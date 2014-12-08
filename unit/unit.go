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
	"github.com/omeid/go-ini"
)

var BasePath = "/etc/geppetto/"

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

	status event.Status

	Meta    Meta
	Service Service

	//The actuall process on system and it's attributes.
	process *exec.Cmd
	signals chan syscall.Signal

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

	u.signals = make(chan syscall.Signal)

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

	//TODO: Consider the arguments?
	u.process = exec.Command(u.Service.ExecStart)

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
		return err
	}

	fmt.Printf("u.process %+v\n", u.process)
	err = u.process.Start()
	fmt.Printf("err %+v\n", err)
	fmt.Println("started...")

	if err != nil {
		return err
	}

	go u.watch()

	return nil
}

func (u *Unit) signal(s syscall.Signal) error {
	return u.process.Process.Signal(s)
}

func (u *Unit) watch() {

	err := make(chan error)
	go func() {
		err <- u.process.Wait()
		close(err)
	}()

	for {
		select {
		case e := <-u.Events:
			_ = e
		case s := <-u.signals:
			_ = s
		case err := <-err:
			_ = err
			return
		}
	}
}
