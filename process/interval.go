package process

import (
	"errors"
	"time"
)

type Interval interface {
	Start() time.Time
	End() time.Time
	Duration() time.Duration
	Rest(t time.Time) time.Duration
	Contains(t time.Time) bool
}

type interval struct {
	startTime time.Time
	duration  time.Duration
}

func NewInterval(start time.Time, duration time.Duration) (*interval, error) {
	if duration/time.Hour > 24 {
		err := errors.New(`the duration of the rule 
			should not be more than 24 hours`)
		return nil, err
	}
	if start.After(start.Add(duration)) {
		err := errors.New(`the beginning of the period should be 
			earlier than the end of the period`)
		return nil, err
	}
	i := &interval{
		startTime: start,
		duration:  duration,
	}
	return i, nil
}

func ParseInterval(startTime string, duration time.Duration /* endTime string */) (*interval, error) {
	start, err := time.Parse("15:04:05", startTime)
	if err != nil {
		return nil, err
	}
	return NewInterval(start, duration)
}

func (i interval) Start() time.Time {
	return date(i.startTime)
}

func (i interval) End() time.Time {
	return i.Start().Add(i.duration)
}

func (i interval) Duration() time.Duration {
	return i.duration
}

func (i interval) Rest(t time.Time) time.Duration {
	return i.End().Sub(t)
}

func (i interval) Contains(t time.Time) bool {
	return t.After(i.Start()) && t.Before(i.End())
}

func date(t time.Time) time.Time {
	year, month, day := now().Date()

	hour, min, sec := t.Clock()

	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func now() time.Time {
	return time.Now().UTC()
}

func Now() time.Time {
	return now()
}
