package exec

import (
	"context"
	"errors"
	"github.com/goingaround/waitqueue"
	"sync"
)

type Worker struct {
	state       State
	end1, wait1 sync.Once
	waitQueue   *waitqueue.WaitQueue
	work        chan work
	done        chan struct{}
	async       bool
	cancel      context.CancelFunc
}

type work struct {
	do  func() error
	fin bool
}

func NewWorker(queueSize int) (*Worker, error) {
	w, _, err := NewWorkerWithContext(context.Background(), queueSize)
	return w, err
}

func NewWorkerWithContext(ctx context.Context, queueSize int) (*Worker, context.Context, error) {
	wq, err := waitqueue.New(queueSize)
	if err != nil {
		return nil, nil, err
	}

	wCtx, cancel := context.WithCancel(ctx)

	w := &Worker{
		work:      make(chan work),
		done:      make(chan struct{}),
		state:     NewState(state{code: execDO}),
		waitQueue: wq,
		cancel:    cancel,
	}
	go w.up(wCtx)

	return w, wCtx, nil
}

func (w *Worker) up(ctx context.Context) {
	defer close(w.done)
	for {
		select {
		case <-ctx.Done(): // aborted
			return

		case x := <-w.work:
			if x.fin { // done
				return
			}
			w.do(x)
		}
	}
}

func (w *Worker) do(x work) {
	if w.async {
		go func() {
			defer w.waitQueue.Deq()
			// extra safety for async case when one goroutine aborted on error
			// other executing goroutines will not proceed, further work is avoided
			if w.state.Err() == nil {
				if err := x.do(); err != nil {
					w.abort(err)
				}
			}
		}()
	} else {
		defer w.waitQueue.Deq()
		if err := x.do(); err != nil {
			w.abort(err)
		}
	}
}

func (w *Worker) Do(f func() error) {
	if w.state.Code() == execDO {

		w.waitQueue.Enq() // invoker is stuck here if queue size reached

		x := work{do: f, fin: false}
		select {
		case <-w.done:
			// when got done signal, work is not relevant
			// and waitQueue shall be dequeued
			w.waitQueue.Deq()
			return

		case w.work <- x: // send work
			return
		}
	}
}

func (w *Worker) Wait() error {
	if w.state.Code() == execDO {
		w.wait1.Do(func() {
			// not accepting more work after wait was called
			w.state.Set(state{code: execWAIT})
			// wait for all work to be done
			w.waitQueue.Wait()
			// shut down normally
			w.end1.Do(func() {
				w.state.Set(state{code: execDONE})
				w.work <- work{do: func() error { return nil }, fin: true}
			})
		})
	}
	<-w.done // wait for done signal
	return w.state.Err()
}

func (w *Worker) Async() {
	w.async = true
}

func (w *Worker) Done() <-chan struct{} {
	return w.done
}

func (w *Worker) Abort() {
	w.abort(errors.New("aborted manually"))
}

func (w *Worker) abort(err error) {
	w.end1.Do(func() {
		w.state.Set(state{code: execABORT, err: err})
		w.cancel()
	})
}
