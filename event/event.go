package event

import "sync"

type Status string

func (s Status) Error() string {
	return string(s)
}

const (
/*
	not-found
	active
	loaded
	inactive
	waiting
	running
	exited
	dead
*/
//TODO: Socket activeation??
//listening

//FUTURE:
//elapsed
//mounted
//plugged
//stub
)

const (
	StatusLoaded                 Status = "Loaded."
	StatusAlreadyLoaded          Status = "Already Loaded."
	StatusTransactionRegistering Status = "Registering Transaction."
	StatusTransactionRegistered  Status = "Registering Transaction."
	StatusBye                    Status = "Bye."
)

func NewEvent(from string, status Status) *Event {
	return &Event{from, status}
}

type Event struct {
	from   string
	status Status
}

func (e *Event) From() string {
	return e.from
}

func (e *Event) Status() Status {
	return e.status
}

func (e *Event) String() string {
	return string(e.Status())
}

func NewPipe() *Pipe {
	return &Pipe{chans: make(map[string]chan<- *Event)}
}

type Pipe struct {
	sync.Mutex
	chans map[string]chan<- *Event
}

func (p *Pipe) Emit(e *Event) {
	p.Lock()
	defer p.Unlock()

	//TODO: Should we wait for every transaction to process the event before sending another event?
	//var wg sync.WaitGroup
	//wg.Add(len(u.transactions))

	for _, ch := range p.chans {
		go func() {
			ch <- e
			//wg.Done()
		}()
	}
	//wg.Wait()
}

func (p *Pipe) Add(name string, ch chan *Event) {
	p.Lock()
	defer p.Unlock()
	p.chans[name] = ch
}

func (p *Pipe) Drop(name string) {
	p.Lock()
	defer p.Unlock()
	delete(p.chans, name)
}
