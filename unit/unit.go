package unit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	Listeners *event.Pipe

	Deps *UnitList
}

func (u *Unit) Prepare() error {

	u.Listeners = event.NewPipe()

	if u.prepared {
		return nil
	}

	var err error

	err = u.Meta.Prepare()
	if err != nil {
		return err
	}

	err = u.Service.Prepare()
	if err != nil {
		return err
	}

	cred, err := u.Service.Credentails()
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

func (u *Unit) Status() event.Status {
	return u.status

}

func (u *Unit) setStatus(s event.Status) {
	//TODO: Notify the channel.
	u.status = s
}

func (u *Unit) Start() error {

	return nil
}
func (u *Unit) Wait() {}
