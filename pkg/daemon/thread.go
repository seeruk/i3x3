package daemon

import "context"

// Thread is a generic interface for some sort of process that can be started and stopped. It is not
// necessarily run in the background, and it may not necessarily have to be stopped by calling stop
// (i.e. it could end on it's own).
type Thread interface {
	Start() error
	Stop() error
}

// BackgroundThreadResult represents the result of a background thread's work. It is sent when work
// has finished in the background thread.
type BackgroundThreadResult struct {
	// Error may contain an error that occurred.
	Error error
}

// NewBackgroundThread starts the given Thread, and then waits for the given context to signal that
// it should stop it's work, allowing the thread to (hopefully) end gracefully. The returned channel
// will receive a message when the thread has finished stopping. It should only receive one message.
func NewBackgroundThread(ctx context.Context, thread Thread) <-chan BackgroundThreadResult {
	bail := make(chan struct{}, 1)
	done := make(chan BackgroundThreadResult, 1)

	go func() {
		// First, wait for some kind of signal to do anything.
		select {
		case <-ctx.Done():
		case <-bail:
			// We should only reach this return if there was an error starting the thread.
			return
		}

		// Attempt to stop work. This should be a blocking call, and should unblock once stopping
		// has finished. Otherwise there's no point in using it...
		err := thread.Stop()

		// Once we stop, we need to tell the other routine to not try to send more results.
		bail <- struct{}{}
		done <- BackgroundThreadResult{
			Error: err,
		}
	}()

	go func() {
		// This call should block. Once this stops, we will check to see if we requested a shutdown.
		err := thread.Start()

		select {
		case <-bail:
			return // If a shutdown has already been requested.
		default:
		}

		// If we didn't request a shutdown, we will want to stop the goroutine above from waiting
		// for something to tell it to shutdown (i.e. the context). Then, we'll send a response that
		// may contain an error (there may not always be an error).
		bail <- struct{}{}
		done <- BackgroundThreadResult{
			Error: err,
		}
	}()

	return done
}
