package minutemarktimer

import (
	"errors"
	"fmt"
	"time"

	"github.com/matt-doug-davidson/timestamps"
)

type HandlerSettings struct {
	Interval string `md:"interval"` // The start delay (ex. 1m, 1h, etc.), immediate if not specified
	Offset   string `md:"offset"`   // The repeat interval (ex. 1m, 1h, etc.), doesn't repeat if not specified
}

type MarkTimer struct {
	Interval      int64
	Offset        int64
	Callback      MinuteMarkTimerCallback
	nextTimestamp int64
}

type MinuteMarkTimer struct {
	timers []*MarkTimer
}

func NewMinuteMarkTimer() *MinuteMarkTimer {
	mmt := &MinuteMarkTimer{timers: make([]*MarkTimer, 0)}
	return mmt
}

type MinuteMarkTimerCallback func()

// Init implements trigger.Init
func (m *MinuteMarkTimer) AddTimer(interval int64, offset int64, callback MinuteMarkTimerCallback) error {

	if interval > 60 {
		fmt.Println("Error: interval must be less than or equal to 60")
		return errors.New("Interval must be less than or equal to 60")
	}
	if 60%interval != 0 {
		fmt.Println("Error: interval must be a factor of 60")
		return errors.New("Interval must be a factor of 60")
	}

	if interval < 2 {
		if offset > 1 {
			fmt.Println("Error: interval of 2 cannot have offset greater than 1")
			return errors.New("Interval of 2 cannot have offset greater than 1")
		}
	}

	//m.addMarkTimer(interval, offset, callback)
	timer := &MarkTimer{
		Interval:      interval,
		Offset:        offset,
		Callback:      callback,
		nextTimestamp: 0,
	}
	m.timers = append(m.timers, timer)

	return nil
}

func (m *MarkTimer) adjust() {
	m.nextTimestamp = calculateNextMark(m.Interval, m.Offset)
}

// func (t *Trigger) addMarkTimer(interval int64, offset int64, callback MinuteMarkTimerCallback) {
// 	timer := &MarkTimer{
// 		Interval:      interval,
// 		Offset:        offset,
// 		Callback:      callback,
// 		nextTimestamp: 0,
// 	}
// 	t.timers = append(t.timers, timer)
// }

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
	nextMarkNanoSecs += timestamps.Nanoseconds(float64(offset) * 60) // InputOffset is seconds. Convert to minutes
	return nextMarkNanoSecs
}

func (m *MinuteMarkTimer) findEarliestNext() int64 {
	earliest := m.timers[0].nextTimestamp
	for _, timer := range m.timers {
		if timer.nextTimestamp < earliest {
			earliest = timer.nextTimestamp
		}
	}
	return earliest
}

func (m *MinuteMarkTimer) findEarliestDelay() time.Duration {
	earliest := m.findEarliestNext()
	nanoSeconds := epochNanoSecNow()
	delayToNextMark := earliest - nanoSeconds
	d := time.Duration(delayToNextMark) * time.Nanosecond
	return d
}

func (m *MinuteMarkTimer) adjustTimers() {
	secondsNow := epochNanoSecNow()
	minutesNow := secondsNow / 60000000000
	for _, timer := range m.timers {
		minutes := timer.nextTimestamp / 60000000000
		if minutesNow == minutes {
			go timer.Callback()
			timer.adjust()
		}
	}
}

var stop chan bool = nil

func (m *MinuteMarkTimer) schedule() chan bool {
	stop := make(chan bool)
	delay := m.findEarliestDelay()
	go func() {
		for {
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
			m.adjustTimers()
			delay = m.findEarliestDelay()
		}
	}()

	return stop
}

// Start implements ext.Trigger.Start
func (m *MinuteMarkTimer) Start() error {
	for _, timer := range m.timers {
		nm := calculateNextMark(timer.Interval, timer.Offset)
		timer.nextTimestamp = nm
	}
	m.schedule()
	return nil
}

// Stop implements ext.Trigger.Stop
func (m *MinuteMarkTimer) Stop() error {
	// Pushing true into the stop channel blocks at this point.
	//stop <- true
	return nil
}
