// Package eventrouter is a simple trie event routing package.
// Handlers can Subscribe to events that will be Published at some later
// date using simple prefix matching.
package eventrouter

import "strings"

// Route encapsulates an Event's route data.
type Route struct {
	parts []string
	index int
}

func (r *Route) next() bool {
	r.index++
	if r.index >= len(r.parts) {
		return false
	}

	return true
}

// Current returns the part of the route that corresponds to the depth of the
// current handler.
func (r Route) Current() string {
	return r.parts[r.index]
}

// Event wraps a payload and route information for passing into Handlers.
type Event struct {
	Route   Route
	Payload interface{}
}

// Handler defines a subscribable Handler for responding to Events.
type Handler interface {
	Handle(Event)
}

type opFn func(map[string][]Handler)

// Router manages a simple trie of Handlers that can respond to Events by
// prefix matching incoming routes against the routes of subscribed Handlers.
type Router struct {
	ops chan opFn
}

// see https://dave.cheney.net/2016/11/13/do-not-fear-first-class-functions for
// inspiration
func (r *Router) loop() {
	handlers := make(map[string][]Handler)
	for op := range r.ops {
		op(handlers)
	}
}

func (r *Router) start() {
	r.ops = make(chan opFn)
	go r.loop()
}

func (r *Router) pushOp(fn opFn) {
	if r.ops == nil {
		r.start()
	}

	r.ops <- fn
}

// New returns a configured Router. This function is for conveniently passing of
// Handlers upon initialization.
func New(ms ...map[string][]Handler) *Router {
	var r Router
	for _, m := range ms {
		for rt, hs := range m {
			for _, h := range hs {
				r.Subscribe(rt, h)
			}
		}
	}
	return &r
}

type routeHandler struct {
	*Router
}

// Handle performs Event routing for the subscribed Handlers.
func (r routeHandler) Handle(e Event) {
	r.pushOp(func(handlers map[string][]Handler) {
		if !e.Route.next() {
			return
		}

		hs := handlers[e.Route.Current()]
		hs = append(hs, handlers["*"]...)

		for _, h := range hs {
			h.Handle(e)
		}
	})
}

// Publish triggers any Handlers subscribed to the route to handle an Event
// containing the provided payload.
func (r *Router) Publish(rt string, p interface{}) {
	routeHandler{r}.Handle(Event{
		Route: Route{
			parts: strings.Split(rt, "."),
			index: -1,
		},
		Payload: p,
	})
}

// Subscribe adds a Handler to the Router to respond to the given route.
func (r *Router) Subscribe(rt string, h Handler) {
	r.pushOp(func(handlers map[string][]Handler) {
		parts := strings.Split(rt, ".")
		if len(parts) > 1 {
			var r Router
			r.Subscribe(strings.Join(parts[1:], "."), h)
			h = routeHandler{&r}
		}

		hs, ok := handlers[parts[0]]
		if !ok {
			hs = make([]Handler, 0, 1)
		}

		hs = append(hs, h)
		handlers[parts[0]] = hs
	})
}

// Unsubscribe removes a specifc Handler from the Router for a given route.
//
// It must be noted that this function is heavily dependent on user input. The
// Handlers subscribed previously *must* be comparable or it will panic.
// See here: https://golang.org/ref/spec#Comparison_operators.
func (r *Router) Unsubscribe(rt string, h Handler) {
	r.pushOp(func(handlers map[string][]Handler) {
		parts := strings.Split(rt, ".")
		hs, ok := handlers[parts[0]]
		if !ok {
			return
		}

		if len(parts) > 1 {
			for _, handler := range hs {
				r, ok := handler.(routeHandler)
				if ok {
					r.Unsubscribe(strings.Join(parts[1:], "."), h)
					h = handler
				}
			}
		}

		for i, handler := range hs {
			if h == handler {
				hs = append(hs[:i], hs[i+1:]...)
				handlers[parts[0]] = hs
				break
			}
		}
	})
}
