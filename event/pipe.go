package event

import "sync"

func NewPipe() *Pipe {
	return &Pipe{
		chans: make(map[string]chan<- Event),
	}
}

type Pipe struct {
	sync.Mutex
	chans map[string]chan<- Event
}

func (p *Pipe) Watch(ch <-chan Event) {
	for e := range ch {
		p.Emit(e)
	}
}

func (p *Pipe) Emit(e Event) {
	p.Lock()
	defer p.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(p.chans))

	for _, ch := range p.chans {
		go func(ch chan<- Event) {
			ch <- e
			wg.Done()
		}(ch)
	}

	wg.Wait()
}
func (p *Pipe) Pipe(ch <-chan Event) {
	for e := range ch {
		p.Emit(e)
	}
}

func (p *Pipe) New(name string) chan Event {
	ch := make(chan Event)
	p.Add(name, ch)
	return ch
}

func (p *Pipe) Add(name string, ch chan Event) {
	p.Lock()
	defer p.Unlock()
	p.chans[name] = ch
}

func (p *Pipe) Drop(name string) {
	p.Lock()
	defer p.Unlock()
	delete(p.chans, name)
}

func (p *Pipe) Count() int {
	p.Lock()
	defer p.Unlock()
	return len(p.chans)
}

func (p *Pipe) List() []string {
	p.Lock()
	defer p.Unlock()
	list := []string{}
	for name, _ := range p.chans {
		list = append(list, name)
	}
	return list
}
