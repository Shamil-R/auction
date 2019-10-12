package process

import (
	"errors"
	"time"
)

type Timeline interface {
	Start() time.Time
	End() time.Time
	Added() time.Duration
	Run(start time.Time, end time.Time) <-chan bool
	Stop() error
	Prolong(d time.Duration)
}

type timeline struct {
	start      time.Time
	end        time.Time
	added      time.Duration
	isRun      bool
	chComplete chan bool
	chSkip     chan bool
	chStop     chan bool
}

func newTimeline() *timeline {
	return &timeline{
		chComplete: make(chan bool),
		chSkip:     make(chan bool),
		chStop:     make(chan bool),
	}
}

func (t *timeline) Start() time.Time {
	return t.start
}

func (t *timeline) End() time.Time {
	return t.end.Add(t.added)
}

func (t *timeline) Added() time.Duration {
	return t.added
}

func (t *timeline) Run(start time.Time, end time.Time) <-chan bool {
	if t.isRun {
		t.chStop <- true
	}
	t.start = start
	t.end = end
	t.added = 0
	t.isRun = true
	rest := end.Sub(start)
	go t.runTimer(rest)
	return t.chComplete
}

func (t *timeline) Prolong(d time.Duration) {
	t.chSkip <- true
	t.added += d
	rest := t.end.Add(t.added).Sub(now())
	go t.runTimer(rest)
}

func (t *timeline) Stop() error {
	if !t.isRun {
		return errors.New("timer not started")
	}
	t.chStop <- true
	return nil
}

func (t *timeline) runTimer(d time.Duration) {
	timer := time.NewTimer(d)
	select {
	case <-t.chSkip:
	case <-t.chStop:
		t.isRun = false
		t.chComplete <- false
	case <-timer.C:
		t.isRun = false
		t.chComplete <- true
	}
}
