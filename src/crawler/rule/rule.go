package rule

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"../../common/util"
)

const (
	defaultTimeout     = time.Duration(math.MaxInt64)
	defaultInterval    = time.Hour
	defaultDelay       = time.Duration(0)
	defaultConcurrency = 10
)

type Rule struct {
	Timeout     time.Duration
	Interval    time.Duration
	Delay       time.Duration
	Concurrency int

	Entries         []*Entry
	TargetTemplates map[string]*TargetTemplate

	json gjson.Result
}

func NewRule() *Rule {
	return &Rule{
		Timeout:     defaultTimeout,
		Interval:    defaultInterval,
		Delay:       defaultDelay,
		Concurrency: defaultConcurrency,

		Entries:         []*Entry{},
		TargetTemplates: map[string]*TargetTemplate{},
	}
}

func NewRuleWithContent(ruleContent string) (*Rule, error) {
	if !gjson.Valid(ruleContent) {
		return nil, errors.New("Invalid rule content")
	}

	r := NewRule()
	r.json = gjson.Parse(ruleContent)

	r.Timeout, _ = parseTimeout(&r.json)
	r.Interval, _ = parseInterval(&r.json)
	r.Delay, _ = parseDelay(&r.json)
	r.Concurrency, _ = parseConcurrency(&r.json)

	r.Entries = parseEntries(&r.json)
	r.TargetTemplates = parseTargetTemplates(&r.json)

	return r, nil
}

func parseTimeout(json *gjson.Result) (time.Duration, error) {
	// Parse $timeout
	timeoutElem := json.Get("$timeout") // 5s | 10m | 1h | 7d
	if timeoutElem.Exists() {
		str := strings.ToLower(timeoutElem.String())
		duration, err := util.ParseDuration(str)
		if err != nil {
			return defaultTimeout, errors.New("Parse $timeout error")
		}

		return duration, nil
	}

	return defaultTimeout, nil
}

func parseInterval(json *gjson.Result) (time.Duration, error) {
	// Parse $every
	everyElem := json.Get("$every") // 5s | 10m | 1h | 7d
	if everyElem.Exists() {
		str := strings.ToLower(everyElem.String())
		duration, err := util.ParseDuration(str)
		if err != nil {
			return defaultInterval, errors.New("Parse $every error")
		}

		return duration, nil
	}

	return defaultInterval, nil
}

func parseDelay(json *gjson.Result) (time.Duration, error) {
	now := time.Now()

	// Parse $startDay and $startDayTime
	nextTime := now.Add(0)
	startDayElem := json.Get("$startDay") // w0 | m15 | y125
	if startDayElem.Exists() {
		str := strings.ToLower(startDayElem.String())
		if strings.HasPrefix(str, "w") { // Sunday = 0, ..., 6
			weekDay, err := strconv.Atoi(strings.TrimPrefix(str, "w"))
			if err != nil {
				return defaultDelay, err
			}

			if weekDay < 0 || weekDay > 6 {
				return defaultDelay, errors.New("$startDay out of range")
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
				return defaultDelay, err
			}

			if monthDay < 1 || monthDay > 30 {
				return defaultDelay, errors.New("$startDay out of range")
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
				return defaultDelay, err
			}

			if yearDay < 1 || yearDay > 365 {
				return defaultDelay, errors.New("$startDay out of range")
			}

			currentYearDay := nextTime.YearDay()
			if yearDay > currentYearDay {
				nextTime = nextTime.AddDate(0, 0, yearDay-currentYearDay)
			} else if yearDay < currentYearDay {
				nextTime = nextTime.AddDate(1, 0, yearDay-currentYearDay)
			}
		}
	}
	startDayTimeElem := json.Get("$startDayTime") // 19:00:00
	if startDayTimeElem.Exists() {
		str := startDayTimeElem.String()
		timeOfDay, err := time.Parse("15:04:05", str)
		if err != nil {
			return defaultDelay, errors.New("Parse $startDayTime error")
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

	// Calculate Delay, zero means job should be executed immediately
	return nextTime.Sub(now), nil
}

func parseConcurrency(json *gjson.Result) (int, error) {
	// Parse $concurrency
	concurrencyElem := json.Get("$concurrency")
	if concurrencyElem.Exists() {
		return int(concurrencyElem.Int()), nil
	}

	return defaultConcurrency, nil
}

func parseEntries(json *gjson.Result) []*Entry {
	entries := []*Entry{}

	entriesElem := json.Get("$entries")
	if entriesElem.Exists() {
		entriesElem.ForEach(func(kElem, entryElem gjson.Result) bool {
			entry := NewEntryWithJson(&entryElem)
			entries = append(entries, entry)

			return true
		})
	}

	return entries
}

func parseTargetTemplates(json *gjson.Result) map[string]*TargetTemplate {
	tts := map[string]*TargetTemplate{}

	json.ForEach(func(kElem, vElem gjson.Result) bool {
		name := kElem.String()
		if !strings.HasPrefix(name, "$") {
			tts[name] = NewTargetTemplateWithJson(&vElem)
		}

		return true
	})

	return tts
}
