package eventrouter

import "strings"

type Event struct {
	Route   []string
	Payload interface{}
}

type Handler func(Event)

type Router interface {
	Publish(e string, p interface{})
	Subscribe(e string, h Handler)
}

type router struct {
	handlers map[string][]Handler
}

func New() Router {
	return &router{
		handlers: make(map[string][]Handler),
	}
}

func (r *router) Subscribe(e string, h Handler) {
	hs, ok := r.handlers[e]
	if !ok {
		hs = make([]Handler, 0, 1)
	}

	hs = append(hs, h)
	r.handlers[e] = hs
}

func (r *router) collect(e string) []Handler {
	var hss []Handler
	for p, hs := range r.handlers {
		if p == e || p == "*" {
			hss = append(hss, hs...)
		}
	}
	return hss
}

func (r *router) Publish(e string, p interface{}) {
	hs := r.collect(e)

	ps := strings.Split(e, ".")
	for _, h := range hs {
		h(Event{
			Route:   ps,
			Payload: p,
		})
	}
}
