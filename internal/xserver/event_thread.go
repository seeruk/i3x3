package xserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/inconshreveable/log15"
)

// EventThread is a thread that polls (sadly) for X events, notifying an channel whenever an event
// occurs, so it may be reacted to.
type EventThread struct {
	sync.Mutex

	ctx    context.Context
	cfn    context.CancelFunc
	logger log15.Logger
	xconn  *xgb.Conn

	outCh chan<- struct{}
}

// NewEventThread creates a new instance of event thread.
func NewEventThread(logger log15.Logger, outCh chan<- struct{}) *EventThread {
	logger = logger.New("module", "xserver/eventThread")

	return &EventThread{
		logger: logger,
		outCh:  outCh,
	}
}

// Start attempts to stop the event thread.
func (t *EventThread) Start() error {
	var err error

	t.Lock()
	t.ctx, t.cfn = context.WithCancel(context.Background())
	t.xconn, err = initialiseXEnvironment()
	t.Unlock()

	if err != nil {
		return err
	}

	t.logger.Info("thread started")

	defer func() {
		t.logger.Info("thread stopped")
	}()

	for {
		ev, xerr := t.xconn.WaitForEventContext(t.ctx)
		if xerr != nil {
			return xerr
		}

		if ev == nil {
			continue
		}

		t.handleEvent()
	}

	return nil
}

// Stop attempts to stop the event thread.
func (t *EventThread) Stop() error {
	t.Lock()
	defer t.Unlock()

	if t.ctx != nil && t.cfn != nil {
		t.cfn()
	}

	if t.xconn != nil {
		t.xconn.Close()
	}

	return nil
}

// handleEvent notifies the given out channel that some event occurred.
func (t *EventThread) handleEvent() {
	t.outCh <- struct{}{}
}

// initialiseXEnvironment sets up the connection to the X server, and prepare our randr
// configuration so we can act appropriately whenever an event happens.
func initialiseXEnvironment() (*xgb.Conn, error) {
	x, err := xgb.NewConn()
	if err != nil {
		return nil, fmt.Errorf("error establishing X connectiong: %v", err)
	}

	err = randr.Init(x)
	if err != nil {
		return nil, fmt.Errorf("error initialising randr: %v", err)
	}

	// Get the root window on the default screen.
	root := xproto.Setup(x).DefaultScreen(x).Root

	// Choose which events we watch for.
	events := randr.NotifyMaskScreenChange

	// Subscribe to some events.
	err = randr.SelectInputChecked(x, root, uint16(events)).Check()
	if err != nil {
		return nil, fmt.Errorf("error subscribing to events: %v", err)
	}

	return x, nil
}
