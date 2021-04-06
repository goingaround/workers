# exec.Worker
An object built to execute <_func() error_> parallelly to the caller's processing continuum.

thread safe, not reuseable after Wait() / Abort().

`func NewWorker(queueSize int) (*Worker, error)` - Initializes a new worker

`func NewWorkerWithContext(ctx context.Context, queueSize int) (*Worker, context.Context, error)` - Initializes a new worker with context

- _queueSize_ - is the buffer size of the work queue ( > 0)

returns error on 
 - _queueSize_ < 1
 
`(w *Worker) Do(do func() error)` - sends new work, if queued work count >= queueSize invoker will wait here

`(w *Worker) Wait() error` - waiting for the worker to finish execution and returning the result

`(w *Worker) Done() <-chan struct{}` - support for broadcast done channel

`(w *Worker) Async()` - sets the worker to async mode (= work would be executed as goroutine)

`(w *Worker) Abort()` - manually aborts the execution
