package event

import "sync"

func Pipe(From <-chan Event, To chan<- Event) <-chan Event {
	out := make(chan Event)

	go func() {
		for e := range From {
			go func(out chan<- Event, e Event) { out <- e }(To, e)
			go func(out chan<- Event, e Event) { out <- e }(out, e)
		}
		close(out)
	}()

	return out
}

func NewTopic() *Topic {
	return &Topic{
		chans: make(map[string]chan<- Event),
		pipes: make(map[string]*Topic),
	}
}

type Topic struct {
	sync.Mutex
	chans map[string]chan<- Event
	pipes map[string]*Topic
}

func (p *Topic) Watch(ch <-chan Event) {
	for e := range ch {
		p.Emit(e)
	}
}

func (p *Topic) Emit(e Event) {
	p.Lock()
	defer p.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(p.chans) + len(p.pipes))

	for _, ch := range p.chans {
		go func(ch chan<- Event) {
			ch <- e
			wg.Done()
		}(ch)
	}

	for _, pp := range p.pipes {
		go func(pp *Topic) {
			pp.Emit(e)
			wg.Done()
		}(pp)
	}

	wg.Wait()
}

/*
func (p *Topic) Topic(ch <-chan Event) {
	for e := range ch {
		p.Emit(e)
	}
}
*/

func (p *Topic) New(name string) chan Event {
	ch := make(chan Event)
	p.Add(name, ch)
	return ch
}

func (p *Topic) Add(name string, ch chan Event) {
	p.Lock()
	defer p.Unlock()
	p.chans[name] = ch
}

func (p *Topic) AddTopic(name string, pp *Topic) {
	p.Lock()
	defer p.Unlock()
	p.pipes[name] = pp
}

func (p *Topic) Drop(name string) {
	p.Lock()
	defer p.Unlock()
	delete(p.chans, name)
}

func (p *Topic) Count() int {
	p.Lock()
	defer p.Unlock()
	return len(p.chans)
}

func (p *Topic) List() []string {
	p.Lock()
	defer p.Unlock()
	list := []string{}
	for name, _ := range p.chans {
		list = append(list, name)
	}
	return list
}
