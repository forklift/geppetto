package unit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/omeid/go-ini"
)

const BASEPATH = "test/"

func Read(name string) (*Unit, error) {

	file, err := os.Open(filepath.Join(BASEPATH, name))
	if err != nil {
		return nil, err
	}

	return Parse(file, name)
}

func Parse(reader io.Reader, name string) (*Unit, error) {
	u := &Unit{}
	u.Name = name

	return u, ini.NewDecoder(reader).Decode(u)
}

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

type Unit struct {
	//Unit is the "unit" files for Geppeto.

	Name   string
	status Status

	Meta    Meta
	Service Service

	//The actuall process on system and it's attributes.
	process *exec.Cmd

	//internals
	prepared bool
}

func (u *Unit) Prepare() error {

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

func (u *Unit) Status() Status {
	return u.status

}

func (u *Unit) setStatus(s Status) {
	//TODO: Notify the channel.
	u.status = s
}

func (u *Unit) Start() {}
func (u *Unit) Wait()  {}
