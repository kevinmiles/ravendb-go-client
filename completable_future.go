package ravendb

import (
	"sync"
	"time"
)

// TODO: write tests

/*
Note:

future = completableFuture.runAsync(() -> foo())

(or supplyAsync) is replaced with this pattern:

future = newCompletableFuture()
go func() {
	res, err := foo()
	if err != nil {
		future.completeWithError(err)
	} else {
		future.complete(res)
	}
}()
*/

// completableFuture helps porting Java code. Implements only functions needed
// by ravendb.
type completableFuture struct {
	mu sync.Mutex

	completed bool
	// used to wait for Future to finish
	signalCompletion chan bool

	// result generated by the Future, only valid if completed
	result interface{}
	err    error
}

func newCompletableFuture() *completableFuture {
	return &completableFuture{
		// channel with capacity 1 so that complete() can finish the goroutine
		// without waiting for someone to call Get()
		signalCompletion: make(chan bool, 1),
	}
}

func newCompletableFutureAlreadyCompleted(result interface{}) *completableFuture {
	res := newCompletableFuture()
	res.complete(result)
	return res
}

func (f *completableFuture) getState() (bool, interface{}, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.completed, f.result, f.err
}

// must be called with f.mu locked
func (f *completableFuture) markCompleted(result interface{}, err error) {
	f.completed = true
	f.result = result
	f.err = err
	f.signalCompletion <- true
}

// complete marks the future as completed with a given result (which can be nil)
func (f *completableFuture) complete(result interface{}) {
	f.mu.Lock()
	if !f.completed {
		f.markCompleted(result, nil)
	}
	f.mu.Unlock()
}

// completeWithError marks the future as completed with error
func (f *completableFuture) completeWithError(err error) {
	f.mu.Lock()
	if !f.completed {
		f.markCompleted(nil, err)
	}
	f.mu.Unlock()
}

// cancel cancels a future
func (f *completableFuture) cancel(mayInterruptIfRunning bool) {
	// mayInterruptIfRunning is ignored, apparently same happens in Java
	// https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/CompletableFuture.html
	f.completeWithError(&CancellationError{})
}

// IsDone returns true if future has been completed, either with a result or error
func (f *completableFuture) IsDone() bool {
	done, _, _ := f.getState()
	return done
}

// IsCompletedExceptionally returns true if future has been completed due to an error
func (f *completableFuture) IsCompletedExceptionally() bool {
	_, _, err := f.getState()
	return err != nil // implies f.done
}

// isCancelled returns true if future was cancelled by calling cancel()
func (f *completableFuture) isCancelled() bool {
	var isCancelled bool
	_, err, _ := f.getState()
	if err != nil {
		_, isCancelled = err.(*CancellationError)
	}
	return isCancelled
}

// Get waits for completion and returns resulting value or error
// If already completed, returns immediately.
// TODO: hide it, unused in one test
func (f *completableFuture) Get() (interface{}, error) {
	return f.GetWithTimeout(0)
}

func (f *completableFuture) GetWithTimeout(dur time.Duration) (interface{}, error) {
	done, res, err := f.getState()
	if done {
		return res, err
	}

	if dur == 0 {
		// wait for the Future to complete
		<-f.signalCompletion
	} else {
		// wait for the Future to complete or timeout to expire
		select {
		case <-f.signalCompletion:
			// completed, will return the result
		case <-time.After(dur):
			// timed out
			return nil, NewTimeoutError("GetWithTimeout() timed out after %s", dur)
		}
	}

	_, res, err = f.getState()
	return res, err
}
