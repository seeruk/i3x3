package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
)

const Interval = time.Minute

// Thread is a thread that gathers metrics and periodically prints them in the log.
type Thread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger
}

// NewThread creates a new metrics thread instance.
func NewThread(logger log15.Logger) *Thread {
	logger = logger.New("module", "daemon/metrics")

	return &Thread{
		logger: logger,
	}
}

// Start attempts to start the metrics thread.
func (t *Thread) Start() error {
	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.Unlock()

	defer func() {
		t.Lock()
		t.ctx = nil
		t.cfn = nil
		t.Unlock()
	}()

	ticker := time.NewTicker(Interval)
	defer ticker.Stop()

	var memStats runtime.MemStats

	t.logger.Info("thread started")

	defer func() {
		t.logger.Info("thread stopped")
	}()

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&memStats)

			t.logger.Debug("metrics",
				"goroutines", runtime.NumGoroutine(),
				"heapAllocBytes", memStats.Alloc,
			)
		case <-t.ctx.Done():
			return t.ctx.Err()
		}
	}
}

// Stop attempts to stop the metrics thread.
func (t *Thread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.ctx != nil && t.cfn != nil {
		t.cfn()
	}

	return nil
}
