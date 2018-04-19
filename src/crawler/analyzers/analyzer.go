package analyzers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"../rule"
	"../sink"
	"../target"
)

var (
	regAction = regexp.MustCompile(`^\$\[[^\$]*\]`) // $[<0 or n chars except $>]
)

type Result struct {
	Mtag      string
	Targets   []*target.Target
	SinkPacks []*sink.SinkPack
}

func relUrlToAbs(relUrl string, parentUrl string) string {
	test, errTest := url.Parse(relUrl)
	if errTest != nil {
		return relUrl
	}

	// Return absolute URL directly
	if test.IsAbs() {
		return relUrl
	}

	// Parse parent URL
	base, errBase := url.Parse(parentUrl)
	if errBase != nil {
		return relUrl
	}

	// Protocol-relative URL
	if strings.HasPrefix(relUrl, "//") {
		return fmt.Sprintf(`%s:%s`, base.Scheme, relUrl)
	}

	// Absolute URL
	abs, errRel := base.Parse(relUrl)
	if errRel != nil {
		return relUrl
	}

	return abs.String()
}

func newFileTarget(dir string, url string, resultType string) *target.Target {
	t := target.NewTarget()
	t.Url = url
	t.ResultType = resultType

	output := rule.NewObjectOutput()
	output.Name = dir

	t.ObjectOutputs = []*rule.ObjectOutput{output}

	return t
}
