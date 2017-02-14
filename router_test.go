package eventrouter

import (
	"reflect"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	tests := []struct {
		desc           string
		subscribeRts   []string
		publishRt      string
		expectedCalled int
	}{
		{
			"top-level event",
			[]string{"event"},
			"event",
			1,
		},
		{
			"top-level wildcard event",
			[]string{"*"},
			"event",
			1,
		},
		{
			"second-level event",
			[]string{"first.second"},
			"first.second",
			1,
		},
		{
			"second-level wildcard first event",
			[]string{"*.second"},
			"first.second",
			1,
		},
		{
			"second-level wildcard second event",
			[]string{"first.*"},
			"first.second",
			1,
		},
		{
			"top-level event 2 handlers",
			[]string{"event", "event"},
			"event",
			2,
		},
		{
			"3 2 1 wildcards",
			[]string{"*", "*.*", "*.*.*"},
			"first.second.third",
			3,
		},
		{
			"top-level longer event",
			[]string{"first"},
			"first.second.third",
			1,
		},
		{
			"handler subscribed deeper than published",
			[]string{"first.second.third"},
			"first",
			0,
		},
		{
			"branching handlers",
			[]string{"first", "first.*", "first.second", "first.*.third", "first.second.third"},
			"first.second.third",
			5,
		},
		{
			"no handlers",
			[]string{"first"},
			"none",
			0,
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

			if called != test.expectedCalled {
				t.Fatalf("handler called count incorrect; expected: %d, actual: %d", test.expectedCalled, called)
			}
		})
	}
}
