package daemon

import "context"

// Thread is a generic interface for some sort of process that can be started and stopped. It is not
// necessarily run in the background, and it may not necessarily have to be stopped by calling stop
// (i.e. it could end on it's own).
type Thread interface {
	Start() error
	Stop() error
}

// NewBackgroundThread starts the given thread, and then waits for the given context to signal that
// it should stop it's work, allowing the thread to end gracefully. The returned channel will
// receive a message when the thread has finished stopping.
func NewBackgroundThread(ctx context.Context, thread Thread) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		<-ctx.Done()

		// @TODO: It'd be nice to have some logging around this.
		// If our context says we're done, signal the thread to stop.
		thread.Stop()
		close(done)
	}()

	thread.Start()

	return done
}
