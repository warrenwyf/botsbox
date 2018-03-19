package job

import (
	"testing"
)

func Test_New(t *testing.T) {
	rule := `
	{
		"$every":"5m",
		"$startDay":"w0",
		"$startDayTime":"03:00:00",
		"$entries":["landing_page"],
		"landing_page":{
			"#bgDiv": "saveurl"
		},
	}
	`

	job, err := NewJob("unittest", rule)
	if err != nil {
		t.Fatalf("NewJob error: %v", err)
	}

	t.Log("Interval", job.Interval)
	t.Log("Delay", job.Delay)
}
