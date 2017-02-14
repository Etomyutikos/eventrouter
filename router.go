package eventrouter

import "strings"

type Event struct {
	Route   []string
	index   int
	Payload interface{}
}

func (e Event) CurrentPart() string {
	return e.Route[e.index]
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
}

func New() Router {
	return &router{
		handlers: make(map[string][]Handler),
	}
}

func new(rt string, h Handler) *router {
	r := &router{
		handlers: make(map[string][]Handler),
	}
	r.Subscribe(rt, h)

	return r
}

// TODO(Erik): support multithreaded operations; https://dave.cheney.net/2016/11/13/do-not-fear-first-class-functions
// TODO(Erik): does branching work like I expect?

func (r *router) Handle(e Event) {
	e.index++
	hs, ok := r.handlers[e.CurrentPart()]
	if !ok {
		hs, ok = r.handlers["*"]
		if !ok {
			return
		}
	}

	for _, h := range hs {
		h.Handle(e)
	}
}

func (r *router) Subscribe(rt string, h Handler) {
	parts := strings.Split(rt, ".")
	if len(parts) > 1 {
		h = new(strings.Join(parts[1:], "."), h)
	}

	hs, ok := r.handlers[parts[0]]
	if !ok {
		hs = make([]Handler, 0, 1)
	}

	hs = append(hs, h)
	r.handlers[parts[0]] = hs
}

func (r *router) Publish(rt string, p interface{}) {
	r.Handle(Event{
		Route:   strings.Split(rt, "."),
		index:   -1,
		Payload: p,
	})
}
