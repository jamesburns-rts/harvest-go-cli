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
	return 60 * (float64(h) - h.Hours())
}

func (h Hours) Hours() float64 {
	return math.Floor(float64(h))
}

func (h Hours) Duration() time.Duration {
	return time.Hour*time.Duration(h.Hours()) + time.Minute*time.Duration(h.Minutes())
}

var anyLetters = regexp.MustCompile("^.*[a-z]+.*$")

func ParseHours(str string) (Hours, error) {
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "")
	if anyLetters.MatchString(str) {
		d, err := time.ParseDuration(str)
		if err != nil {
			return 0, err
		}
		return Hours(d.Hours()), nil
	}
	hours, err := strconv.ParseFloat(str, 64)
	return Hours(hours), err
}

func (h *Hours) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	*h, err = ParseHours(str)
	return err
}

func (h *Hours) Marshal() (out []byte, err error) {
	return []byte(fmt.Sprintf("%0.2f", float64(*h))), nil
}
