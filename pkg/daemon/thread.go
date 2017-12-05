package daemon

import "context"

// Thread is a generic interface for some sort of process that can be started and stopped. It is not
// necessarily run in the background, and it may not necessarily have to be stopped by calling stop
// (i.e. it could end on it's own).
type Thread interface {
	// Start starts some work. The work only end when Stop is called, or it may end on it's own.
	Start() error
	// Stop attempts to stop a started thread.
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

	// Wait for a request to stop from the context, in the background.
	go func() {
		// First, wait for some kind of signal to do anything.
		select {
		case <-ctx.Done():
		case <-bail:
			// If we hit here, it means that the thread stopped on it's own. This might mean there
			// was an error.
			return
		}

		// Attempt to stop work. This should be a blocking call, and should unblock once stopping
		// has finished. Otherwise there's no point in using it...
		err := thread.Stop()

		// Once we stop, we need to tell the other goroutine to not try to send more results. Then
		// we send our result from stopping.
		bail <- struct{}{}
		done <- BackgroundThreadResult{
			Error: err,
		}
	}()

	// Start the thread, and wait for it to stop, or be stopped.
	go func() {
		// This call should block. Once this stops, we will check to see if we requested a shutdown.
		// If we didn't (i.e. if the other thread hasn't told us to bail) then this thread must have
		// stopped on it's own (maybe an error).
		err := thread.Start()
		if err == nil {
			bail <- struct{}{}
			done <- BackgroundThreadResult{
				Error: err,
			}
			return
		}

		select {
		case <-bail:
			// If a stop was requested, we will have already sent the result (the result from
			// stopping), therefore, we shouldn't send another result
			return
		default:
		}

		// If the Thread didn't stop because it was told to (but rather because it finished for
		// whatever reason), then we need to tell the other goroutine to bail, as there's no point
		// in it trying to tell a Thread that's already stopped to stop.
		bail <- struct{}{}

		// Once we've told the other thread to bail, send the result of the Thread's start call.
		done <- BackgroundThreadResult{
			Error: err,
		}
	}()

	return done
}
