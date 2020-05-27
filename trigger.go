package minutemarktimer

import (
	"github.com/carlescere/scheduler"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

type HandlerSettings struct {
	StartInterval  string `md:"startDelay"`     // The start delay (ex. 1m, 1h, etc.), immediate if not specified
	RepeatInterval string `md:"repeatInterval"` // The repeat interval (ex. 1m, 1h, etc.), doesn't repeat if not specified
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

	return nil
}

// Start implements ext.Trigger.Start
func (t *Trigger) Start() error {
	return nil
}

// Stop implements ext.Trigger.Stop
func (t *Trigger) Stop() error {
	return nil
}
