package eventrouter

import (
	"strings"
	"sync"
)

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

func New() Router {
	r := &router{
		ops: make(chan func(map[string][]Handler)),
	}
	go r.loop()
	return r
}

func (r *router) Handle(e Event) {
	var wg sync.WaitGroup
	wg.Add(1)

	r.ops <- func(handlers map[string][]Handler) {
		defer wg.Done()

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

	wg.Wait()
}

func (r *router) Publish(rt string, p interface{}) {
	r.Handle(Event{
		Route:   strings.Split(rt, "."),
		index:   -1,
		Payload: p,
	})
}

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
