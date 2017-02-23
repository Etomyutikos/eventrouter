package eventrouter

import "strings"

// Event wraps a payload and route information for passing into Handlers.
type Event struct {
	Route   []string
	index   int
	Payload interface{}
}

func (e *Event) next() bool {
	e.index++
	if e.index >= len(e.Route) {
		return false
	}

	return true
}

// CurrentPart returns the part of the Event's route that corresponds to
// the depth of the current handler.
func (e Event) CurrentPart() string {
	return e.Route[e.index]
}

type Handler interface {
	Handle(Event)
}

type Router interface {
	Publish(route string, payload interface{})
	Subscribe(route string, handler Handler)
	Unsubscribe(route string, handler Handler)
}

type router struct {
	ops chan func(map[string][]Handler)
}

// see https://dave.cheney.net/2016/11/13/do-not-fear-first-class-functions for
// inspiration
func (r *router) loop() {
	handlers := make(map[string][]Handler)
	for op := range r.ops {
		op(handlers)
	}
}

// New returns a configured Router.
func New() Router {
	r := &router{
		ops: make(chan func(map[string][]Handler)),
	}
	go r.loop()
	return r
}

// Handle performs Event routing for the subscribed Handlers.
func (r *router) Handle(e Event) {
	r.ops <- func(handlers map[string][]Handler) {
		if !e.next() {
			return
		}

		hs, ok := handlers[e.CurrentPart()]
		if !ok {
			hs, ok = handlers["*"]
			if !ok {
				return
			}
		}

		for _, h := range hs {
			h.Handle(e)
		}
	}
}

// Publish triggers any Handlers subscribed to the route to handle an Event
// containing the provided payload.
func (r *router) Publish(rt string, p interface{}) {
	r.Handle(Event{
		Route:   strings.Split(rt, "."),
		index:   -1,
		Payload: p,
	})
}

// Subscribe adds a Handler to the Router to respond to the given route.
func (r *router) Subscribe(rt string, h Handler) {
	r.ops <- func(handlers map[string][]Handler) {
		parts := strings.Split(rt, ".")
		if len(parts) > 1 {
			r := New()
			r.Subscribe(strings.Join(parts[1:], "."), h)
			h = r.(*router)
		}

		hs, ok := handlers[parts[0]]
		if !ok {
			hs = make([]Handler, 0, 1)
		}

		hs = append(hs, h)
		handlers[parts[0]] = hs
	}
}

// Unsubscribe removes a specifc Handler from the Router for a given route.
func (r *router) Unsubscribe(rt string, h Handler) {
	r.ops <- func(handlers map[string][]Handler) {
		parts := strings.Split(rt, ".")
		hs, ok := handlers[parts[0]]
		if !ok {
			return
		}

		if len(parts) > 1 {
			for _, handler := range hs {
				r, ok := handler.(Router)
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
	}
}
