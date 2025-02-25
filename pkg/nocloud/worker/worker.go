package worker

import "sync"

type work struct {
	function func() error
	retries  int
}

type workersQueue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	queue  []work
	closed bool
}

func newWorkersQueue() *workersQueue {
	wq := &workersQueue{
		queue: make([]work, 0),
	}
	wq.cond = sync.NewCond(&wq.mu)
	return wq
}

func (q *workersQueue) Add(w work) {
	q.mu.Lock()
	q.queue = append(q.queue, w)
	q.mu.Unlock()
	q.cond.Signal()
}

func (q *workersQueue) Get() (work, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.queue) == 0 && !q.closed {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		return work{}, false
	}
	task := q.queue[0]
	q.queue = q.queue[1:]
	return task, true
}

func (q *workersQueue) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.cond.Broadcast()
}

type Worker struct {
	q          *workersQueue
	maxWorkers int
	maxRetries int

	errorLog string
	failed   bool

	tasksWg   sync.WaitGroup
	workersWg sync.WaitGroup
}

func NewWorker(maxWorkers int, maxRetries int) *Worker {
	return &Worker{
		q:          newWorkersQueue(),
		maxWorkers: maxWorkers,
		maxRetries: maxRetries,
	}
}

func (w *Worker) Add(f func() error) {
	w.tasksWg.Add(1)
	w.q.Add(work{function: f})
}

func (w *Worker) Start() (success bool, errorLog string) {
	for i := 0; i < w.maxWorkers; i++ {
		w.workersWg.Add(1)
		go w.worker()
	}
	w.tasksWg.Wait()
	w.q.Close()
	w.workersWg.Wait()

	success = !w.failed
	errorLog = w.errorLog
	return
}

func (w *Worker) worker() {
	defer w.workersWg.Done()
	for {
		task, ok := w.q.Get()
		if !ok {
			return
		}
		if err := task.function(); err != nil {
			w.errorLog += err.Error() + "\n"
			task.retries++
			if task.retries > w.maxRetries {
				w.failed = true
				w.tasksWg.Done()
			} else {
				w.q.Add(task)
			}
		} else {
			w.tasksWg.Done()
		}
	}
}
