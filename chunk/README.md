# chunk.Worker
An object built to execute _chunk.Func_ for chunks of _chunk.Values_.

thread safe, not reuseable after Wait() / Abort().

`func NewWorker(chunkSize int, queueSize int, f Func) (*Worker, error)` - Initializes a new worker

`func NewWorkerWithContext(ctx context.Context, chunkSize int, queueSize int, f Func) (*Worker, context.Context, error)` - Initializes a new worker with context

- _chunkSize_ - is the execution length trigger ( > 0)
- _queueSize_ - is the buffer size of the work queue ( > 0)
- _f_ - is the function to be executed on the chunks `func(Values) error`

returns error on 
 - _chunkSize_ < 1
 - _queueSize_ < 1
 - _f_ == nil
 
`(w *Worker) Feed(v Values)` - feeds the worker with more values

`(w *Worker) Wait() error` - waiting for the worker to finish execution and returning the result

`(w *Worker) Done() <-chan struct{}` - support for broadcast done channel

`(w *Worker) Async()` - sets the worker to async mode (= work would be executed as goroutine)

`(w *Worker) Abort()` - manually aborts the execution