package eventrouter

import (
	"reflect"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		desc        string
		subscribeRt string
		publishRt   string
		p           interface{}
	}{
		{
			"top-level event",
			"event",
			"event",
			"payload",
		},
		{
			"top-level wildcard event",
			"*",
			"event",
			"payload",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			r := New()

			var called bool
			h := func(e Event) {
				called = true

				expectedRt := strings.Split(test.publishRt, ".")
				if !reflect.DeepEqual(e.Route, expectedRt) {
					t.Fatalf("incorrect route; expected: %v, actual: %v", expectedRt, e.Route)
				}

				if e.Payload != test.p {
					t.Fatalf("incorrect payload; expected: %s, actual: %s", test.p, e.Payload)
				}
			}

			r.Subscribe(test.subscribeRt, h)
			r.Publish(test.publishRt, test.p)

			if !called {
				t.Fatal("handler was never called")
			}
		})
	}
}

func TestRouterMultipleHandlers(t *testing.T) {
	r := New()

	rt := "event"
	p := "payload"

	var called int
	h := func(e Event) {
		called += 1

		expectedRt := []string{"event"}
		if !reflect.DeepEqual(e.Route, expectedRt) {
			t.Fatalf("incorrect route; expected: %v, actual: %v", expectedRt, e.Route)
		}

		if e.Payload != p {
			t.Fatalf("incorrect payload; expected: %s, actual: %s", p, e.Payload)
		}
	}

	r.Subscribe(rt, h)
	r.Subscribe(rt, h)

	r.Publish(rt, p)

	if called != 2 {
		t.Fatalf("handler called count incorrect; expected: %d, actual: %d", 2, called)
	}
}
