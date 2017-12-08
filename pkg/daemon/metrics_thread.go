package daemon

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
)

const MetricsInterval = time.Minute

// MetricsThread is a thread that gathers metrics and periodically prints them in the log.
type MetricsThread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger
}

// NewMetricsThread creates a new metrics thread instance.
func NewMetricsThread(logger log15.Logger) *MetricsThread {
	logger = logger.New("module", "daemon/metrics")

	return &MetricsThread{
		logger: logger,
	}
}

// Start attempts to start the metrics thread.
func (t *MetricsThread) Start() error {
	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.Unlock()

	defer func() {
		t.Lock()
		t.ctx = nil
		t.cfn = nil
		t.Unlock()
	}()

	ticker := time.NewTicker(MetricsInterval)
	defer ticker.Stop()

	var memStats runtime.MemStats

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
func (t *MetricsThread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.ctx != nil && t.cfn != nil {
		t.cfn()
	}

	return nil
}
