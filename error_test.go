package asynctask_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/go-asynctask"
	"github.com/stretchr/testify/assert"
)

type structError struct{}

func (pe structError) Error() string {
	return "Error from struct type"
}

type pointerError struct{}

func (pe *pointerError) Error() string {
	return "Error from pointer type"
}

func getPanicTask(sleepDuration time.Duration) asynctask.AsyncFunc {
	return func(ctx context.Context) (interface{}, error) {
		time.Sleep(sleepDuration)
		panic("yo")
	}
}

func getErrorTask(errorString string, sleepDuration time.Duration) asynctask.AsyncFunc {
	return func(ctx context.Context) (interface{}, error) {
		time.Sleep(sleepDuration)
		return nil, errors.New(errorString)
	}
}

func TestTimeoutCase(t *testing.T) {
	t.Parallel()
	ctx, cancelFunc := newTestContextWithTimeout(t, 3*time.Second)
	defer cancelFunc()

	tsk := asynctask.Start(ctx, getCountingTask(10, 200*time.Millisecond))
	_, err := tsk.WaitWithTimeout(ctx, 300*time.Millisecond)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expecting DeadlineExceeded")

	// the last Wait error should affect running task
	// I can continue wait with longer time
	rawResult, err := tsk.WaitWithTimeout(ctx, 2*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, 9, rawResult)

	// any following Wait should complete immediately
	rawResult, err = tsk.WaitWithTimeout(ctx, 2*time.Nanosecond)
	assert.NoError(t, err)
	assert.Equal(t, 9, rawResult)
}

func TestPanicCase(t *testing.T) {
	t.Parallel()
	ctx, cancelFunc := newTestContextWithTimeout(t, 3*time.Second)
	defer cancelFunc()

	tsk := asynctask.Start(ctx, getPanicTask(200*time.Millisecond))
	_, err := tsk.WaitWithTimeout(ctx, 300*time.Millisecond)
	assert.True(t, errors.Is(err, asynctask.ErrPanic), "expecting ErrPanic")
}

func TestErrorCase(t *testing.T) {
	t.Parallel()
	ctx, cancelFunc := newTestContextWithTimeout(t, 3*time.Second)
	defer cancelFunc()

	tsk := asynctask.Start(ctx, getErrorTask("dummy error", 200*time.Millisecond))
	_, err := tsk.WaitWithTimeout(ctx, 300*time.Millisecond)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, asynctask.ErrPanic), "not expecting ErrPanic")
	assert.False(t, errors.Is(err, context.DeadlineExceeded), "not expecting DeadlineExceeded")
	assert.Equal(t, "dummy error", err.Error())
}

func TestPointerErrorCase(t *testing.T) {
	t.Parallel()
	ctx, cancelFunc := newTestContextWithTimeout(t, 3*time.Second)
	defer cancelFunc()

	// nil point of a type that implement error
	var pe *pointerError = nil
	// pass this nil pointer to error interface
	var err error = pe
	// now you get a non-nil error
	assert.False(t, err == nil, "reason this test is needed")

	tsk := asynctask.Start(ctx, func(ctx context.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		var pe *pointerError = nil
		return "Done", pe
	})

	result, err := tsk.Wait(ctx)
	assert.NoError(t, err)
	assert.Equal(t, result, "Done")
}

func TestStructErrorCase(t *testing.T) {
	t.Parallel()
	ctx, cancelFunc := newTestContextWithTimeout(t, 3*time.Second)
	defer cancelFunc()

	// nil point of a type that implement error
	var se structError
	// pass this nil pointer to error interface
	var err error = se
	// now you get a non-nil error
	assert.False(t, err == nil, "reason this test is needed")

	tsk := asynctask.Start(ctx, func(ctx context.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		var se structError
		return "Done", se
	})

	result, err := tsk.Wait(ctx)
	assert.NoError(t, err)
	assert.Equal(t, result, "Done")
}
