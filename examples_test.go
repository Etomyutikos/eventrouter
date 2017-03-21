package eventrouter

import (
	"fmt"
	"sync"
)

// SomeHandler implements the eventrouter.Handler interface.
type SomeHandler struct {
	handle func(Event)
}

func (h SomeHandler) Handle(e Event) {
	h.handle(e)
}

func Example() {
	// zero-value initialization
	var r Router

	// event handling is asynchronous; this is here so we can see output
	var wg sync.WaitGroup
	wg.Add(1)

	// will respond to any Publish beginning "first"
	// e.g., "first", "first.second", "first.*", ...
	r.Subscribe("first", SomeHandler{
		handle: func(e Event) {
			fmt.Printf("e: %#v", e)
			wg.Done()
		},
	})

	// Publish does not block
	r.Publish("first", "some payload")
	wg.Wait()

	// Output: e: eventrouter.Event{Route:eventrouter.Route{parts:[]string{"first"}, index:0}, Payload:"some payload"}
}

func Example_prefixMatching() {
	// event handling is asynchronous; these are here so we can guarantee output order
	var firstWg sync.WaitGroup
	firstWg.Add(1)

	var secondWg sync.WaitGroup
	secondWg.Add(1)

	var thirdWg sync.WaitGroup
	thirdWg.Add(1)

	handlers := map[string][]Handler{
		"first": []Handler{
			SomeHandler{
				handle: func(e Event) {
					fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
					firstWg.Done()
				},
			},
		},
		"first.second": []Handler{
			SomeHandler{
				handle: func(e Event) {
					firstWg.Wait()
					fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
					secondWg.Done()
				},
			},
		},
		"first.second.third": []Handler{
			SomeHandler{
				handle: func(e Event) {
					secondWg.Wait()
					fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
					thirdWg.Done()
				},
			},
		},
	}

	// initialize with map of string to slice of Handler
	r := New(handlers)

	// Publish does not block
	r.Publish("first.second.third", "some payload")
	thirdWg.Wait()

	// Output:
	// e.Route.Current(): first
	// e.Route.Current(): second
	// e.Route.Current(): third
}

func Example_wildcardMatching() {
	// zero-value initialization
	var r Router

	// event handling is asynchronous; this is here so we can see output
	var wg sync.WaitGroup
	wg.Add(3)

	// will respond to any Publish beginning "first"
	r.Subscribe("first", SomeHandler{
		handle: func(e Event) {
			fmt.Printf("e.Route: %v\n", e.Route.parts)
			fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
			wg.Done()
		},
	})

	// will respond to any Publish
	r.Subscribe("*", SomeHandler{
		handle: func(e Event) {
			fmt.Printf("e.Route: %v\n", e.Route.parts)
			fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
			wg.Done()
		},
	})

	// will respond to any two-part Publish
	r.Subscribe("*.*", SomeHandler{
		handle: func(e Event) {
			fmt.Printf("e.Route: %v\n", e.Route.parts)
			fmt.Printf("e.Route.Current(): %s\n", e.Route.Current())
			wg.Done()
		},
	})

	// Publish does not block
	r.Publish("first.second", "some payload")
	wg.Wait()

	// Output:
	// e.Route: [first second]
	// e.Route.Current(): first
	// e.Route: [first second]
	// e.Route.Current(): first
	// e.Route: [first second]
	// e.Route.Current(): second
}
