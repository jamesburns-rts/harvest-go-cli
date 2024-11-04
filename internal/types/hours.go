package types

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Hours float64

func (h Hours) Minutes() float64 {
	if h >= 0 {
		return 60 * (float64(h) - h.Hours())
	}
	return 60 * (h.Hours() - float64(h))
}

func (h Hours) Hours() float64 {
	if h >= 0 {
		return math.Floor(float64(h))
	}
	return math.Ceil(float64(h))
}

func (h Hours) Duration() time.Duration {
	return time.Hour*time.Duration(h.Hours()) + time.Minute*time.Duration(h.Minutes())
}

var anyLetters = regexp.MustCompile("^.*[a-z]+.*$")

func ParseHours(str string) (*Hours, error) {
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "")

	if str == "" {
		return nil, nil
	}

	var hours float64
	if anyLetters.MatchString(str) {
		d, err := time.ParseDuration(str)
		if err != nil {
			return nil, err
		}
		hours = d.Hours()
	} else {
		var err error
		if hours, err = strconv.ParseFloat(str, 64); err != nil {
			return nil, err
		}
	}
	h := Hours(hours)
	return &h, nil
}

func (h *Hours) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	if hours, err := ParseHours(str); err != nil {
		return err
	} else {
		*h = *hours
		return nil
	}
}

func (h *Hours) Marshal() (out []byte, err error) {
	return []byte(fmt.Sprintf("%0.2f", float64(*h))), nil
}

func (h *Hours) FloatPtr() *float64 {
	if h == nil {
		return nil
	}
	return (*float64)(h)
}
