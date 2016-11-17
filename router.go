package eventrouter

import "strings"

type Event struct {
	Route   []string
	Payload interface{}
}

type Handler interface {
	Handle(Event)
}

type HandlerFunc func(Event)

func (f HandlerFunc) Handle(e Event) {
	f(e)
}

type Router interface {
	Publish(rt string, p interface{})
	Subscribe(rt string, h Handler)
}

type router struct {
	handlers map[string][]Handler
	routers  map[string]*router
}

func New() Router {
	return &router{
		handlers: make(map[string][]Handler),
	}
}

func (r *router) Subscribe(rt string, h Handler) {
	hs, ok := r.handlers[rt]
	if !ok {
		hs = make([]Handler, 0, 1)
	}

	hs = append(hs, h)
	r.handlers[rt] = hs
}

func (r *router) collect(rt string) []Handler {
	var hss []Handler
	for p, hs := range r.handlers {
		if p == rt || p == "*" {
			hss = append(hss, hs...)
		}
	}
	return hss
}

func (r *router) Publish(rt string, p interface{}) {
	hs := r.collect(rt)

	ps := strings.Split(rt, ".")
	for _, h := range hs {
		h.Handle(Event{
			Route:   ps,
			Payload: p,
		})
	}
}
