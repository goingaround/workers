package exec

import (
	"errors"
	"github.com/goingaround/parallel"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWorker_QueueSizeZero(t *testing.T) {
	_, err := NewWorker(0)
	require.EqualError(t, err, "queue size must be bigger than 0")
}

func TestWorker_AbortManually(t *testing.T) {
	w, err := NewWorker(2)
	require.NoError(t, err)
	op, err := parallel.NewOperation(multiplyFunc(func() {
		w.Do(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
	}, 5)...)
	require.NoError(t, err)

	require.NoError(t, op.Run(0, 0))
	w.Abort()
	require.Error(t, w.Wait())
}

func TestWorker_AbortManually_Async(t *testing.T) {
	w, err := NewWorker(2)
	require.NoError(t, err)
	w.Async()
	op, err := parallel.NewOperation(multiplyFunc(func() {
		w.Do(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
	}, 5)...)
	require.NoError(t, err)

	require.NoError(t, op.Run(0, 0))
	w.Abort()
	require.Error(t, w.Wait())
}

func TestWorker_AbortForError(t *testing.T) {
	w, err := NewWorker(2)
	require.NoError(t, err)
	op, err := parallel.NewOperation(multiplyFunc(func() {
		w.Do(func() error { return errors.New("error") })
	}, 3)...)
	require.NoError(t, err)

	require.NoError(t, op.Run(0, 0))
	require.Error(t, w.Wait())
}

func TestWorker_AbortForError_Async(t *testing.T) {
	w, err := NewWorker(2)
	require.NoError(t, err)
	w.Async()
	op, err := parallel.NewOperation(multiplyFunc(func() {
		w.Do(func() error { return errors.New("error") })
	}, 3)...)
	require.NoError(t, err)

	require.NoError(t, op.Run(0, 0))
	require.Error(t, w.Wait())
}

func multiplyFunc(f func(), n int) []func() {
	if n <= 0 {
		return []func(){}
	}
	res := make([]func(), 0, n)
	for i := 0; i < n; i++ {
		res = append(res, f)
	}
	return res
}
