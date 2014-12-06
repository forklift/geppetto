package unit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/forklift/geppetto/event"
	"github.com/omeid/go-ini"
)

const BASEPATH = "test/"

func Read(unit *Unit) error {

	file, err := os.Open(filepath.Join(BASEPATH, unit.Name))
	if err != nil {
		return err
	}

	return Parse(file, unit)
}

func Parse(reader io.Reader, unit *Unit) error {
	return ini.NewDecoder(reader).Decode(unit)
}

func Make(names []string) []*Unit {
	units := []*Unit{}
	for _, name := range names {
		units = append(units, &Unit{Name: name})
	}
	return units
}

type Unit struct {
	//Unit is the "unit" files for Geppeto.
	Explicit bool

	Name   string
	status event.Status

	Meta    Meta
	Service Service

	//The actuall process on system and it's attributes.
	process *exec.Cmd

	//internals
	prepared bool

	//Transaction
	Ch        chan *event.Event
	Listeners *event.Pipe

	Deps *UnitList

	//Keeping it safe.
	lock sync.Mutex
}

func (u *Unit) Prepare() error {

	if u.prepared {
		return nil
	}

	u.Listeners = event.NewPipe()
	u.Ch = make(chan *event.Event)

	var err error

	err = u.Meta.Prepare()
	if err != nil {
		return err
	}

	err = u.Service.Prepare()
	if err != nil {
		return err
	}

	cred, err := u.Service.BuildCredentails()
	if err != nil {
		return err
	}

	//TODO: Consider the arguments?
	u.process = exec.Command(u.Service.ExecStart)

	u.process.Dir = u.Service.WorkingDirectory
	u.process.Stdin = u.Service.Stdin
	u.process.Stdout = u.Service.Stdout
	u.process.Stderr = u.Service.Stderr

	u.process.SysProcAttr = &syscall.SysProcAttr{
		Chroot:     u.Service.Chroot,
		Credential: cred,
	}

	u.prepared = true
	return nil
}

func (u *Unit) Clean() []error {
	return u.Service.CloseIO() //TODO: Check if we have other deps.
}

func (u *Unit) Status() event.Status {
	return u.status

}

func (u *Unit) Send(s event.Status) {
	//TODO: Notify the channel.
	u.status = s
}

func (u *Unit) Start() error {

	u.lock.Lock()
	defer u.lock.Unlock()

	pipeline, errs, cancel, units := NewPipeline()

	started := pipeline.Do(errs, cancel, units, pipeline.StartUnit)

	//pipeline.
	err := pipeline.Wait(errs, cancel, started)

	if err != nil {
		return err
	}

	err = u.process.Start()

	if err != nil {
		return err
	}

	go u.watch()

	go func() {
		err := u.process.Wait()
		_ = err
		u.Ch <- event.NewEvent(u.Name, event.StatusBye)
	}()
	return err
}

func (u *Unit) watch() {
	for e := range u.Ch {
		u.Listeners.Emit(e)
		//What to do with events?
	}
}
