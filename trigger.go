package minutemarktimer

import (
	"fmt"

	"github.com/carlescere/scheduler"
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

type Trigger struct {
	timers   []*scheduler.Job
	handlers []trigger.Handler
	logger   log.Logger
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
		t.logger.Info("Interval: ", s.Interval)
		t.logger.Info("Offset: ", s.Offset)

		t.logger.Infof("Type of Interval is %T", s.Interval)
		t.logger.Infof("Type of Offset is %T", s.Offset)

		fmt.Println(handler)
		t.logger.Info("Initialize: Handler loop")
	}

	return nil
}

type MarkTimer struct {
	Interval      int64
	Offset        int64
	handler       trigger.Handler
	nextTimestamp int64
}

func (m *MarkTimer) adjust() {
	//m.nextTimestamp = calculateNextMark(m.Interval, m.Offset)
}

var timers []*MarkTimer

func addMarkTimer(interval int64, offset int64, handler trigger.Handler) {
	timer := &MarkTimer{
		Interval:      interval,
		Offset:        offset,
		handler:       handler,
		nextTimestamp: 0,
	}
	timers = append(timers, timer)
	fmt.Println(timers)

}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {
	t.logger.Info("Starting")

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {
	t.logger.Info("Stopping")
	return nil
}
