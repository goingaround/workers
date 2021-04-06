package chunk

import (
	"context"
	"errors"
	"github.com/goingaround/workers/exec"
	"sync"
)

// Worker is built to run Func for chunks of Values
// it appends Values via Feed func, and calls exec.Worker for execution when needed
// this kind of structure is giving the advantage of
// continue processing new Values while executing (qSize+1) ready chunks
// once called Wait this obj is not to be fed with more work
type Worker struct {
	mu     sync.Mutex
	xw     *exec.Worker
	values Values
	f      Func
	size   int
}

type Func func(Values) error

func NewWorker(chunkSize, queueSize int, f Func) (*Worker, error) {
	xw, _, err := NewWorkerWithContext(context.Background(), chunkSize, queueSize, f)
	return xw, err
}

func NewWorkerWithContext(ctx context.Context, chunkSize, queueSize int, f Func) (*Worker, context.Context, error) {

	if chunkSize < 1 {
		return nil, nil, errors.New("chunk size can not be smaller than 1")
	}
	if f == nil {
		return nil, nil, errors.New("given function is nil")
	}

	xw, wCtx, err := exec.NewWorkerWithContext(ctx, queueSize)
	if err != nil {
		return nil, nil, err
	}

	cw := &Worker{
		xw:     xw,
		values: &pseudoValues{},
		f:      f,
		size:   chunkSize,
	}

	return cw, wCtx, nil
}

func (w *Worker) Feed(v Values) {
	if v == nil || v.Len() == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.values = w.values.Merge(v)

	if w.values.Len() >= w.size {

		for v = w.values; v.Len() >= w.size; v = v.Range(w.size, v.Len()) {
			w.do(v.Range(0, w.size))
		}
		w.values = v
	}
}

func (w *Worker) Wait() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// exec leftover if needed
	if w.values.Len() > 0 {
		w.do(w.values)
		w.values = w.values.Range(0, 0)
	}
	return w.xw.Wait()
}

func (w *Worker) do(v Values) {
	w.xw.Do(func() error { return w.f(v) })
}

func (w *Worker) Async() {
	w.xw.Async()
}

func (w *Worker) Done() <-chan struct{} {
	return w.xw.Done()
}

func (w *Worker) Abort() {
	w.xw.Abort()
}
