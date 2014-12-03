package event

type Status string

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
	StatusLoaded        Status = "Loaded."
	StatusAlreadyLoaded Status = "Already Loaded."
)

type Event struct {
	unit   string
	status Status
}

func (e *Event) Unit() string {
	return e.unit
}

func (e *Event) Status() Status {
	return e.status
}

func (e *Event) String() string {
	return string(e.Status())
}

func NewPipe() *Pipe {
	return &Pipe{make(map[string]chan *Event)}
}

type Pipe struct {
	chans map[string]chan *Event
}

func (p *Pipe) Emit(e *Event) {
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
	p.chans[name] = ch
}

func (p *Pipe) Drop(name string) {
	delete(p.chans, name)
}
