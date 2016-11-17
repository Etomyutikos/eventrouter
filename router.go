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
	handlers map[string]Handler
}

func New() Router {
	return &router{
		handlers: make(map[string]Handler),
	}
}

func (r *router) Subscribe(e string, h Handler) {
	r.handlers[e] = h
}

func (r *router) collect(e string) []Handler {
	var hs []Handler
	for p, h := range r.handlers {
		if p == e || p == "*" {
			hs = append(hs, h)
		}
	}
	return hs
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
