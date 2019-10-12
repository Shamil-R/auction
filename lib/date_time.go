package time_lib

import (
	"gitlab/nefco/auction/errors"
	"time"
)

const (
	DateTimeFormat           = "2006-01-02T15:04:05Z"
)

type DateTime string

func (d DateTime) Validate() error {
	_ , err :=  time.Parse(DateTimeFormat, d.String())
	if err != nil {
		return errors.BadRequest("invalid date param")
	}
	return nil
}

func (d DateTime) String() string {
	return string(d)
}

func (d DateTime) UTC() time.Time {
	t, _ := time.Parse(DateTimeFormat, d.String())
	return t
}

func (d DateTime) After(t time.Time) bool {
	return d.UTC().After(t)
}

func (d DateTime) Before(t time.Time) bool {
	return d.UTC().Before(t)
}

func NewDateTime(t time.Time) DateTime {
	s := t.UTC().Format(DateTimeFormat)
	return DateTime(s)
}

func ParseStringToDateTime(timeString string) (DateTime, error) {
	t, err := time.Parse(DateTimeFormat, timeString)
	if err != nil {
		return "", err
	}
	dt := NewDateTime(t)
	if err := dt.Validate(); err != nil {
		return "", err
	}

	return dt, nil
}

func (d DateTime) ValidateDateTime() error {
	_, err := time.Parse(DateTimeFormat, d.String())
	return err
}