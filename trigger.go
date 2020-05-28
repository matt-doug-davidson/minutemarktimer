package minutemarktimer

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

type HandlerSettings struct {
	Interval string `md:"interval"` // The start delay (ex. 1m, 1h, etc.), immediate if not specified
	Offset   string `md:"offset"`   // The repeat interval (ex. 1m, 1h, etc.), doesn't repeat if not specified
}

var triggerMd = trigger.NewMetadata(&HandlerSettings{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &Trigger{}, nil
}

type MarkTimer struct {
	Interval      int64
	Offset        int64
	handler       trigger.Handler
	nextTimestamp int64
}

type Trigger struct {
	//timers   []*scheduler.Job
	handlers []trigger.Handler
	logger   log.Logger
	timers   []*MarkTimer
}

// Init implements trigger.Init
func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	t.handlers = ctx.GetHandlers()
	t.logger = ctx.Logger()

	s := &HandlerSettings{}

	//Handlers slice data is available here
	for _, handler := range t.handlers {

		err := metadata.MapToStruct(handler.Settings(), s, true)
		if err != nil {
			t.logger.Error("Mapping metadata to struct failed", err.Error())
			return err
		}

		// Convert to int64
		interval, err := strconv.ParseInt(s.Interval, 10, 64)
		if err != nil {
			t.logger.Error("Interval conversion failed. Error:", err.Error())
			return err
		}

		if interval > 60 {
			t.logger.Error("Interval must be less than or equal to 60")
			return errors.New("Interval must be less than or equal to 60")
		}
		if 60%interval != 0 {
			t.logger.Error("Interval must be a factor of 60")
			return errors.New("Interval must be a factor of 60")
		}

		offset, err := strconv.ParseInt(s.Offset, 10, 64)
		if err != nil {
			t.logger.Error("Offset conversion failed. Error:", err.Error())
			return err
		}
		if interval < 2 {
			if offset > 1 {
				t.logger.Error("Interval of 2 cannot have offset greater than 1")
				return errors.New("Interval of 2 cannot have offset greater than 1")
			}
		}

		t.addMarkTimer(interval, offset, handler)
	}

	return nil
}

func (m *MarkTimer) adjust() {
	m.nextTimestamp = calculateNextMark(m.Interval, m.Offset)
}

func (t *Trigger) addMarkTimer(interval int64, offset int64, handler trigger.Handler) {
	timer := &MarkTimer{
		Interval:      interval,
		Offset:        offset,
		handler:       handler,
		nextTimestamp: 0,
	}
	t.timers = append(t.timers, timer)
}

func epochSecondsNow() int64 {
	return time.Now().Unix()
}

func epochNanoSecNow() int64 {
	return time.Now().UnixNano()
}

func calculateNextMark(interval int64, offset int64) int64 {
	// Get current times
	seconds := epochSecondsNow()
	// Current mark
	currentMinute := seconds / 60                             // minutes
	minuteOfHour := currentMinute % 60                        // minutes
	currentMarkOfHour := (minuteOfHour / interval) * interval // minutes
	// Next Mark
	nextMarkOfHour := currentMarkOfHour + interval
	nextMarkSeconds := currentMinute + nextMarkOfHour - minuteOfHour
	nextMarkNanoSecs := nextMarkSeconds * 60000000000
	return nextMarkNanoSecs
}

func (t *Trigger) findEarliestNext() int64 {
	earliest := t.timers[0].nextTimestamp
	for _, timer := range t.timers {
		if timer.nextTimestamp < earliest {
			earliest = timer.nextTimestamp
		}
	}
	return earliest
}

func (t *Trigger) findEarliestDelay() time.Duration {
	earliest := t.findEarliestNext()
	nanoSeconds := epochNanoSecNow()
	delayToNextMark := earliest - nanoSeconds
	d := time.Duration(delayToNextMark) * time.Nanosecond
	return d
}

func (t *Trigger) runHandler(handler trigger.Handler) {
	//t.logger.Info("Running handle ", time.Now())
	_, err := handler.Handle(context.Background(), nil)
	if err != nil {
		t.logger.Error("Error running handler: ", err.Error())
	}
}

func (t *Trigger) adjustTimers() {
	secondsNow := epochNanoSecNow()
	minutesNow := secondsNow / 60000000000
	for _, timer := range t.timers {
		minutes := timer.nextTimestamp / 60000000000
		if minutesNow == minutes {
			go t.runHandler(timer.handler)
			timer.adjust()
		}
	}
}

var stop chan bool = nil

func (t *Trigger) schedule() chan bool {
	stop := make(chan bool)
	delay := t.findEarliestDelay()
	go func() {
		for {
			select {
			case <-time.After(delay):
			case <-stop:
				t.logger.Info("Stopping")
				return
			}
			t.adjustTimers()
			delay = t.findEarliestDelay()
		}
	}()

	return stop
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {
	for _, timer := range t.timers {
		nm := calculateNextMark(timer.Interval, timer.Offset)
		timer.nextTimestamp = nm
	}
	t.schedule()
	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {
	// Pushing true into the stop channel blocks at this point.
	//stop <- true
	return nil
}
