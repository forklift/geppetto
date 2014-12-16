package unit

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/forklift/geppetto/event"
	"github.com/mattn/go-shellwords"
	"github.com/omeid/go-ini"
)

var BasePath = "/etc/geppetto/services"

func New(name string) (*Unit, error) {
	u := &Unit{Name: name}
	return u, Read(u)
}

func Read(unit *Unit) error {

	file, err := os.Open(filepath.Join(BasePath, unit.Name))
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

	// Reserved names `gGod`, `gCurrentTransaction`
	Name string

	status event.Type

	Meta    Meta
	Service Service

	//The actuall process on system and it's attributes.
	process           *exec.Cmd
	processExitStatus chan error

	//internals
	prepared bool

	//Transaction
	Events    chan event.Event
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
	u.Events = make(chan event.Event)

	var err error

	err = u.Meta.Prepare()
	if err != nil {
		return fmt.Errorf("Meta: %s", err)
	}

	err = u.Service.Prepare()
	if err != nil {
		return err
	}

	cred, err := u.Service.BuildCredentails()

	u.processExitStatus = make(chan error, 1)

	//TODO: We need to export, $HOME, $USER, $EDITOR, et all. Should be done at Engine level, or deamon?
	shellwords.ParseEnv = true
	cmd, err := shellwords.Parse(u.Service.ExecStart)
	if err != nil {
		return err
	}

	u.process = exec.Command(cmd[0])
	u.process.Args = cmd

	u.process.Dir = u.Service.WorkingDirectory

	u.process.SysProcAttr = &syscall.SysProcAttr{
		Setsid:     true,
		Credential: cred,
	}

	if u.Service.Chroot != "" && u.Service.Chroot != "/" {
		u.process.SysProcAttr.Chroot = u.Service.Chroot
	}

	u.prepared = true
	return nil
}

func (u *Unit) Clean() []error {
	return u.Service.CloseIO() //TODO: Check if we have other deps.
}

func (u *Unit) Type() event.Type {
	return u.status

}

func (u *Unit) Start() chan event.Event {

	u.lock.Lock()
	events := make(chan event.Event)

	go func() {
		defer u.lock.Unlock()
		defer close(events)

		/*pipeline, errs, cancel, units := NewPipeline()

		started := pipeline.Do(errs, cancel, units, pipeline.StartUnit)

		go func() {
			u.Deps.ForEach(func(u *Unit) {
				units <- u
			})
			close(units)
		}()
		//pipeline.
		err := pipeline.Wait(errs, cancel, started)

		if err != nil {
			return err
		}*/

		var err error
		u.process.Stdin, u.process.Stdout, u.process.Stderr, err = u.Service.ConnectIO()
		if err != nil {
			events <- event.New(u.Name, event.ProcessConnectFailed, err.Error())
			return
		}

		err = u.process.Start()
		go u.processWatch()

		if err != nil {
			events <- event.New(u.Name, event.ProcessStartFailed, err.Error())
			return
		}

	}()

	return events
}

func (u *Unit) Stop(sender string) chan event.Event {
	out := make(chan event.Event)

	go func() {
		u.stop(out, sender)
		close(out)
	}()

	return out
}

func (u *Unit) stop(out chan event.Event, sender string) error {

	u.Listeners.Add("transaction-"+sender, out)
	defer u.Listeners.Drop("transaction-" + sender)

	out <- event.New(u.Name, event.UnitStopping, sender)

	//TODO: Check Deps.
	//If multiple. Drop the sender from Listners and return nil.

	pipeline, errs, cancel, units := NewPipeline()

	stopped := pipeline.Do(errs, cancel, units, func(u *Unit) error {
		return u.stop(out, u.Name)
	})

	go u.Deps.ForEach(func(u *Unit) {
		units <- u
	}).Then(func() { close(units) })

	err := pipeline.Wait(errs, cancel, stopped)

	if err != nil {
		return err
	}

	u.process.Process.Signal(syscall.SIGKILL)

	//Wait for process to exit.
	//TODO: Timeout!
	exitStatus := <-u.processExitStatus

	return exitStatus

}

func (u *Unit) processWatch() {
	u.processExitStatus <- u.process.Wait()
	close(u.processExitStatus)
}

func (u *Unit) unitWatch() {

	for {
		select {
		case e := <-u.Events:
			switch e.Type {
			}
		case exitStatus := <-u.processExitStatus:
			_ = exitStatus
			return
		}
	}
}
