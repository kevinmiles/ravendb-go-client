package ravendb

import (
	"sync"
	"time"
)

// TODO: write tests
// TODO: make private to package if not exposed in public APIs

/*
Note:

future = CompletableFuture.runAsync(() -> foo())

(or supplyAsync) is replaced with this pattern:

future = NewCompletableFuture()
go func() {
	res, err := foo()
	if err != nil {
		future.CompleteExceptionally(err)
	} else {
		future.Complete(res)
	}
}()
*/

// CompletableFuture helps porting Java code. Implements only functions needed
// by ravendb.
type CompletableFuture struct {
	mu sync.Mutex

	completed bool
	// used to wait for Future to finish
	signalCompletion chan bool

	// result generated by the Future, only valid if completed
	result interface{}
	err    error
}

func NewCompletableFuture() *CompletableFuture {
	return &CompletableFuture{
		// channel with capacity 1 so that Complete() can finish the goroutine
		// without waiting for someone to call Get()
		signalCompletion: make(chan bool, 1),
	}
}

func NewCompletableFutureAlreadyCompleted(result interface{}) *CompletableFuture {
	res := NewCompletableFuture()
	res.Complete(result)
	return res
}

func (f *CompletableFuture) getState() (bool, interface{}, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.completed, f.result, f.err
}

// must be called with f.mu locked
func (f *CompletableFuture) markCompleted(result interface{}, err error) {
	f.completed = true
	f.result = result
	f.err = err
	f.signalCompletion <- true
}

// Complete marks the future as completed with a given result (which can be nil)
func (f *CompletableFuture) Complete(result interface{}) {
	f.mu.Lock()
	if !f.completed {
		f.markCompleted(result, nil)
	}
	f.mu.Unlock()
}

// CompleteExceptionally marks the future as completed with error
// TODO: maybe rename to CompleteWithError() since doesn't have exceptions
func (f *CompletableFuture) CompleteExceptionally(err error) {
	f.mu.Lock()
	if !f.completed {
		f.markCompleted(nil, err)
	}
	f.mu.Unlock()
}

// Cancel cancels a future
func (f *CompletableFuture) Cancel(mayInterruptIfRunning bool) {
	// mayInterruptIfRunning is ignored, apparently same happens in Java
	// https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/CompletableFuture.html
	f.CompleteExceptionally(NewCancellationError())
}

// IsDone returns true if future has been completed, either with a result or error
func (f *CompletableFuture) IsDone() bool {
	done, _, _ := f.getState()
	return done
}

// IsCompletedExceptionally returns true if future has been completed due to an error
func (f *CompletableFuture) IsCompletedExceptionally() bool {
	_, err, _ := f.getState()
	return err != nil // implies f.done
}

// IsCancelled returns true if future was cancelled by calling Cancel()
func (f *CompletableFuture) IsCancelled() bool {
	var isCancelled bool
	_, err, _ := f.getState()
	if err != nil {
		_, isCancelled = err.(*CancellationError)
	}
	return isCancelled
}

// Get waits for completion and returns resulting value or error
// If already completed, returns immediately.
func (f *CompletableFuture) Get() (interface{}, error) {
	return f.getWithTimeout(0)
}

// GetWithTimeout waits for completion of the future up to dur and returns
// resulting value or error.
// If already completed, returns immediately.
func (f *CompletableFuture) GetWithTimeout(dur time.Duration) (interface{}, error) {
	return f.getWithTimeout(dur)
}

func (f *CompletableFuture) getWithTimeout(dur time.Duration) (interface{}, error) {
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
			return nil, NewTimeoutException("GetWithTimeout() timed out after", dur)
		}
	}

	_, res, err = f.getState()
	return res, err
}
