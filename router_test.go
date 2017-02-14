package eventrouter

import (
	"reflect"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		desc         string
		subscribeRts []string
		publishRt    string
	}{
		{
			"top-level event",
			[]string{"event"},
			"event",
		},
		{
			"top-level wildcard event",
			[]string{"*"},
			"event",
		},
		{
			"second-level event",
			[]string{"first.second"},
			"first.second",
		},
		{
			"second-level wildcard first event",
			[]string{"*.second"},
			"first.second",
		},
		{
			"second-level wildcard second event",
			[]string{"first.*"},
			"first.second",
		},
		{
			"top-level event 2 handlers",
			[]string{"event", "event"},
			"event",
		},
		{
			"3 2 1 wildcards",
			[]string{"*", "*.*", "*.*.*"},
			"first.second.third",
		},
		{
			"top-level longer event",
			[]string{"first"},
			"first.second.third",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			r := New()
			p := "payload"

			var called int
			h := HandlerFunc(func(e Event) {
				called++

				expectedRt := strings.Split(test.publishRt, ".")
				if !reflect.DeepEqual(e.Route, expectedRt) {
					t.Fatalf("incorrect route; expected: %v, actual: %v", expectedRt, e.Route)
				}

				if e.Payload != p {
					t.Fatalf("incorrect payload; expected: %s, actual: %s", p, e.Payload)
				}
			})

			for _, rt := range test.subscribeRts {
				r.Subscribe(rt, h)
			}

			r.Publish(test.publishRt, p)

			expectedCount := len(test.subscribeRts)
			if called != expectedCount {
				t.Fatalf("handler called count incorrect; expected: %d, actual: %d", expectedCount, called)
			}
		})
	}
}

// three wildcards, only one publish part: panic
// does branching work like I expect?
