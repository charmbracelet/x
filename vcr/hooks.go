package vcr

import (
	"strings"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

var headersToKeep = map[string]struct{}{
	"accept":       {},
	"content-type": {},
	"user-agent":   {},
}

func hookRemoveHeaders(keepAll bool) recorder.HookFunc {
	return func(i *cassette.Interaction) error {
		if keepAll {
			return nil
		}
		for k := range i.Request.Headers {
			if _, ok := headersToKeep[strings.ToLower(k)]; !ok {
				delete(i.Request.Headers, k)
			}
		}
		for k := range i.Response.Headers {
			if _, ok := headersToKeep[strings.ToLower(k)]; !ok {
				delete(i.Response.Headers, k)
			}
		}
		return nil
	}
}
