package chunk

import (
	"errors"
	"github.com/stretchr/testify/require"
	"sort"
	"sync"
	"testing"
)

func TestWorker_ChunkSize0(t *testing.T) {
	_, err := NewWorker(0, 0, func(values Values) error { return nil })
	require.EqualError(t, err, "chunk size can not be smaller than 1")
}

func TestWorker_QueueSizeZero(t *testing.T) {
	_, err := NewWorker(1, 0, func(values Values) error { return nil })
	require.EqualError(t, err, "queue size must be bigger than 0")
}

func TestWorker_NilFunc(t *testing.T) {
	_, err := NewWorker(1, 1, nil)
	require.EqualError(t, err, "given function is nil")
}

func TestWorker_OK(t *testing.T) {
	executed, validate := make(testValues, 0), make(testValues, 0)
	cw, err := NewWorker(3, 5, func(values Values) error {
		executed = append(executed, values.(testValues)...)
		return nil
	})
	require.NoError(t, err)
	for i := 1; i <= 5; i++ {
		v := getTestValues(i, i*10)
		cw.Feed(v)
		validate = append(validate, v...)
	}
	require.NoError(t, cw.Wait())
	require.Equal(t, executed, validate)
}

func TestWorker_OK_Async(t *testing.T) {
	executed, validate := make(testValues, 0), make(testValues, 0)
	var m sync.Mutex
	cw, err := NewWorker(3, 5, func(values Values) error {
		m.Lock()
		defer m.Unlock()
		executed = append(executed, values.(testValues)...)
		return nil
	})
	require.NoError(t, err)
	cw.Async()
	for i := 1; i <= 5; i++ {
		v := getTestValues(i, i*10)
		cw.Feed(v)
		validate = append(validate, v...)
	}
	require.NoError(t, cw.Wait())
	sort.Sort(executed)
	sort.Sort(validate)
	require.Equal(t, executed, validate)
}

func TestWorker_FeedAbort(t *testing.T) {
	cw, err := NewWorker(3, 5, func(values Values) error {
		return errors.New("error")
	})
	require.NoError(t, err)
	cw.Feed(getTestValues(0, 10))
	require.Error(t, cw.Wait())
}

func TestWorker_FeedAbort_Async(t *testing.T) {
	cw, err := NewWorker(3, 5, func(values Values) error {
		return errors.New("error")
	})
	require.NoError(t, err)
	cw.Async()
	cw.Feed(getTestValues(0, 10))
	require.Error(t, cw.Wait())
}

func getTestValues(start, count int) testValues {
	var (
		result = make(testValues, 0, count)
		limit  = start + count
	)
	for i := start; i < limit; i++ {
		result = append(result, i)
	}
	return result
}

type testValues []int

func (v testValues) Merge(bv Values) Values {
	if values, ok := bv.(testValues); ok {
		return append(v, values...)
	}
	return v
}

func (v testValues) Range(i, j int) Values {
	return v[i:j]
}

func (v testValues) Len() int {
	return len(v)
}

func (v testValues) Less(i, j int) bool {
	return v[i] < v[j]
}

func (v testValues) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
