package eventrouter

import (
	"reflect"
	"strings"
	"testing"
)

func TestSingleHandlers(t *testing.T) {
	tests := []struct {
		desc        string
		subscribeRt string
		publishRt   string
	}{
		{
			"top-level event",
			"event",
			"event",
		},
		{
			"top-level wildcard event",
			"*",
			"event",
		},
		{
			"second-level event",
			"first.second",
			"first.second",
		},
		{
			"second-level wildcard first event",
			"*.second",
			"first.second",
		},
		{
			"second-level wildcard second event",
			"first.*",
			"first.second",
		},
		{
			"third-level wildcard first event",
			"*.second.third",
			"first.second.third",
		},
		{
			"third-level wildcard second event",
			"first.*.third",
			"first.second.third",
		},
		{
			"third-level wildcard third event",
			"first.second.*",
			"first.second.third",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			r := New()

			p := "payload"

			var called bool
			h := HandlerFunc(func(e Event) {
				called = true

				expectedRt := strings.Split(test.publishRt, ".")
				if !reflect.DeepEqual(e.Route, expectedRt) {
					t.Fatalf("incorrect route; expected: %v, actual: %v", expectedRt, e.Route)
				}

				if e.Payload != p {
					t.Fatalf("incorrect payload; expected: %s, actual: %s", p, e.Payload)
				}
			})

			r.Subscribe(test.subscribeRt, h)
			r.Publish(test.publishRt, p)

			if !called {
				t.Fatal("handler was never called")
			}
		})
	}
}

func TestMultipleHandlers(t *testing.T) {
	r := New()

	rt := "event"
	p := "payload"

	var called int
	h := HandlerFunc(func(e Event) {
		called += 1

		expectedRt := []string{"event"}
		if !reflect.DeepEqual(e.Route, expectedRt) {
			t.Fatalf("incorrect route; expected: %v, actual: %v", expectedRt, e.Route)
		}

		if e.Payload != p {
			t.Fatalf("incorrect payload; expected: %s, actual: %s", p, e.Payload)
		}
	})

	r.Subscribe(rt, h)
	r.Subscribe(rt, h)

	r.Publish(rt, p)

	if called != 2 {
		t.Fatalf("handler called count incorrect; expected: %d, actual: %d", 2, called)
	}
}
