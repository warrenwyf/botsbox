package util

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	maxDuration = time.Duration(math.MaxInt64)
)

func ParseDuration(str string) (time.Duration, error) { // 5s | 30m | 1h | 2d
	if strings.HasSuffix(str, "s") {
		sec, err := strconv.Atoi(strings.TrimSuffix(str, "s"))
		if err != nil {
			return maxDuration, err
		}

		if sec <= 0 {
			return maxDuration, errors.New("Duration must greater than zero")
		}

		return time.Duration(sec) * time.Second, nil
	} else if strings.HasSuffix(str, "m") {
		min, err := strconv.Atoi(strings.TrimSuffix(str, "m"))
		if err != nil {
			return maxDuration, err
		}

		if min <= 0 {
			return maxDuration, errors.New("Duration must greater than zero")
		}

		return time.Duration(min) * time.Minute, nil
	} else if strings.HasSuffix(str, "h") {
		hour, err := strconv.Atoi(strings.TrimSuffix(str, "h"))
		if err != nil {
			return maxDuration, err
		}

		if hour <= 0 {
			return maxDuration, errors.New("Duration must greater than zero")
		}

		return time.Duration(hour) * time.Hour, nil
	} else if strings.HasSuffix(str, "d") {
		day, err := strconv.Atoi(strings.TrimSuffix(str, "d"))
		if err != nil {
			return maxDuration, err
		}

		if day <= 0 {
			return maxDuration, errors.New("Duration must greater than zero")
		}

		return time.Duration(day) * 24 * time.Hour, nil
	}

	return maxDuration, errors.New("Wrong format")
}
