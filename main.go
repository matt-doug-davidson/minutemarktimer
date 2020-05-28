package main

import (
	"fmt"
	"time"
)

type Handler func(cntx int64, triggerData interface{}) string

func UnixMicro() int64 {
	now := time.Now()
	return now.UnixNano() / 1000
}

func sleepSeconds(d time.Duration) {
	fmt.Println(d)
	var duration time.Duration = d * time.Second
	fmt.Println(duration)
	time.Sleep(duration)
	fmt.Println(d)
}

func printMinute(description string, minute int64) {
	minuteInNanosecods := minute * 60000000000
	timeT := time.Unix(minuteInNanosecods/1000000000, minuteInNanosecods%1000000000)
	fmt.Printf("%s %d %s\n", description, minute, timeT)
}

func epochSecondsNow() int64 {
	return time.Now().Unix()
}

func epochNanoSecNow() int64 {
	return time.Now().UnixNano()
}

/*
func calculateCurrentMark() int64 {
	// Get current times
	start := time.Now()
	fmt.Println(start)
	seconds := epochSecondsNow()
	// Current mark
	currentMinute := seconds / 60                             // minutes
	minuteOfHour := currentMinute % 60                        // minutes
	currentMarkOfHour := (minuteOfHour / interval) * interval // minutes
	return currentMarkOfHour * 60000000000
}
*/

func calculateNextMark(interval int64, offset int64) int64 {
	// Get current times
	start := time.Now()
	fmt.Println(start)
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

func calculateDelay(interval int64, offset int64) time.Duration {
	// Get current times
	start := time.Now()
	fmt.Println(start)
	seconds := epochSecondsNow()
	// Current mark
	currentMinute := seconds / 60                             // minutes
	minuteOfHour := currentMinute % 60                        // minutes
	currentMarkOfHour := (minuteOfHour / interval) * interval // minutes
	// Next Mark
	nextMarkOfHour := currentMarkOfHour + interval
	nextMarkSeconds := currentMinute + nextMarkOfHour - minuteOfHour
	nextMarkNanoSecs := nextMarkSeconds * 60000000000
	// Delay calculation
	nanoSeconds := epochNanoSecNow()
	delayToNextMark := nextMarkNanoSecs - nanoSeconds
	d := time.Duration(delayToNextMark) * time.Nanosecond
	fmt.Printf("%T", d)
	return d
}

type MarkTimer struct {
	Interval      int64
	Offset        int64
	handler       Handler
	nextTimestamp int64
}

func (m *MarkTimer) adjust() {
	m.nextTimestamp = calculateNextMark(m.Interval, m.Offset)
}

var timers []*MarkTimer

func addMarkTimer(interval int64, offset int64, handler Handler) {
	timer := &MarkTimer{
		Interval:      interval,
		Offset:        offset,
		handler:       handler,
		nextTimestamp: 0,
	}
	timers = append(timers, timer)
	fmt.Println(timers)

}

func findEarliestNext() int64 {
	earliest := timers[0].nextTimestamp
	for _, timer := range timers {
		if timer.nextTimestamp < earliest {
			earliest = timer.nextTimestamp
		}
	}
	return earliest
}

func findEarliestDelay() time.Duration {
	earliest := findEarliestNext()
	nanoSeconds := epochNanoSecNow()
	delayToNextMark := earliest - nanoSeconds
	d := time.Duration(delayToNextMark) * time.Nanosecond
	fmt.Println("findEarliestDelay: Delay", d)
	return d
}

func Start() {
	for _, timer := range timers {
		fmt.Println(timer)
		nm := calculateNextMark(timer.Interval, timer.Offset)
		timer.nextTimestamp = nm
		fmt.Println(timer)
	}
	earliest := findEarliestNext()
	fmt.Println("earliest", earliest)
}

func runHandler(handler Handler) {
	err := handler(1, nil)

	if err != "" {
		fmt.Println("Error running handler: ", err)
	}
}

func adjustTimers() {
	//currentMark := calculateCurrentMark()
	//fmt.Println("currentMark", currentMark)
	secondsNow := epochNanoSecNow()
	minutesNow := secondsNow / 60000000000
	//fmt.Println(minutesNow)
	for _, timer := range timers {
		//fmt.Println(timer)
		minutes := timer.nextTimestamp / 60000000000
		if minutesNow == minutes {
			fmt.Println("timer expired. Handler", timer.handler)
			// go timer.handler(timer.Interval, nil)
			go runHandler(timer.handler)
			timer.adjust()
		}
	}
	//earliest := findEarliestNext()
	//fmt.Println("earliest", earliest)
}

func schedule(what func(), delay time.Duration) chan bool {
	fmt.Println("Enter schedule")
	stop := make(chan bool)
	delay = findEarliestDelay()
	go func() {
		for {
			select {
			case <-time.After(delay):
			case <-stop:
				fmt.Println("Stopping")
				return
			}
			//fmt.Println("Here")
			what() // Just print timestamp in callback
			adjustTimers()
			delay = findEarliestDelay()
		}
	}()

	return stop
}

var stop chan bool = nil

func handler(cntx int64, triggerData interface{}) string {
	fmt.Println("handler()", cntx)
	return "End of handler"
}

func main() {

	fmt.Println("Timers")
	fmt.Println("time.Nanosecond", time.Nanosecond)
	secondsNow := epochSecondsNow()
	fmt.Println("epochSecondsNow()", secondsNow)

	addMarkTimer(1, 1, handler)
	addMarkTimer(2, 1, handler)
	addMarkTimer(3, 1, handler)
	addMarkTimer(4, 1, handler)
	Start()
	ping := func() { fmt.Println(time.Now()) }

	stop = schedule(ping, 5*time.Millisecond)
	time.Sleep(15 * time.Minute)
	stop <- true
	time.Sleep(25 * time.Millisecond)

}
