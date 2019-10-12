package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))
	return []byte(stamp), nil
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" {
		return nil
	}

	// TODO: удалить после того как все рейсы пересохранят Duration в новом формате
	if dur, err := strconv.ParseUint(s, 10, 64); err == nil {
		*d = Duration(time.Duration(dur) * time.Minute)
		return nil
	}

	s, err := strconv.Unquote(s)
	if err != nil {
		return err
	}

	t, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(t)

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
