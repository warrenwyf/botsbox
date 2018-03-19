package job

import (
	"errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	maxDuration = time.Duration(math.MaxInt64)
)

type Job struct {
	Title string
	Rule  string

	Interval time.Duration
	Delay    time.Duration
}

func NewJob(title string, rule string) (*Job, error) {
	job := &Job{
		Title: title,
		Rule:  rule,

		Interval: time.Hour,
		Delay:    0,
	}

	// Parse $every
	every := gjson.Get(rule, "$every") // 5s | 10m | 1h
	if every.Exists() {
		str := strings.ToLower(every.String())
		duration, err := parseDuration(str)
		if err != nil {
			return nil, errors.New("Parse $every error")
		}

		job.Interval = duration
	} else {
		return nil, errors.New("Job must have $every definition")
	}

	// Parse $startDay and $startDayTime
	now := time.Now()
	nextTime := time.Now()
	startDay := gjson.Get(rule, "$startDay") // w0 | m15 | y125
	if startDay.Exists() {
		str := strings.ToLower(startDay.String())
		if strings.HasPrefix(str, "w") { // Sunday = 0, ..., 6
			weekDay, err := strconv.Atoi(strings.TrimPrefix(str, "w"))
			if err != nil {
				return nil, err
			}

			if weekDay < 0 || weekDay > 6 {
				return nil, errors.New("$startDay out of range")
			}

			currentWeekDay := int(nextTime.Weekday())
			if weekDay > currentWeekDay {
				nextTime = nextTime.AddDate(0, 0, weekDay-currentWeekDay)
			} else if weekDay < currentWeekDay {
				nextTime = nextTime.AddDate(0, 0, 7+weekDay-currentWeekDay)
			}
		} else if strings.HasPrefix(str, "m") { // 1, 2, ..., 30
			monthDay, err := strconv.Atoi(strings.TrimPrefix(str, "m"))
			if err != nil {
				return nil, err
			}

			if monthDay < 1 || monthDay > 30 {
				return nil, errors.New("$startDay out of range")
			}

			currentMonthDay := nextTime.Day()
			if monthDay > currentMonthDay {
				nextTime = nextTime.AddDate(0, 0, monthDay-currentMonthDay)
			} else if monthDay < currentMonthDay {
				nextTime = nextTime.AddDate(0, 1, monthDay-currentMonthDay)
			}
		} else if strings.HasPrefix(str, "y") { // 1, 2, ..., 365
			yearDay, err := strconv.Atoi(strings.TrimPrefix(str, "y"))
			if err != nil {
				return nil, err
			}

			if yearDay < 1 || yearDay > 365 {
				return nil, errors.New("$startDay out of range")
			}

			currentYearDay := nextTime.YearDay()
			if yearDay > currentYearDay {
				nextTime = nextTime.AddDate(0, 0, yearDay-currentYearDay)
			} else if yearDay < currentYearDay {
				nextTime = nextTime.AddDate(1, 0, yearDay-currentYearDay)
			}
		}
	}
	startDayTime := gjson.Get(rule, "$startDayTime") // 19:00:00
	if startDayTime.Exists() {
		str := startDayTime.String()
		timeOfDay, err := time.Parse("15:04:05", str)
		if err != nil {
			return nil, err
		}

		h, m, s := timeOfDay.Clock()
		dur := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
		ch, cm, cs := nextTime.Clock()
		currentDur := time.Duration(ch)*time.Hour + time.Duration(cm)*time.Minute + time.Duration(cs)*time.Second

		if dur < currentDur {
			nextTime = nextTime.AddDate(0, 0, 1)
		}
		nextTime = nextTime.Add(dur - currentDur)
	}
	job.Delay = nextTime.Sub(now)

	return job, nil
}

func NewJobWithFile(title string, rulePath string) (*Job, error) {
	bytes, err := ioutil.ReadFile(rulePath)
	if err != nil {
		return nil, err
	}

	return NewJob(title, string(bytes))
}

func parseDuration(str string) (time.Duration, error) { // 5s | 30m | 1h | 2d
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

		return time.Duration(day) * time.Hour, nil
	}

	return maxDuration, errors.New("Wrong format")
}

func (self *Job) Fn() {

}
