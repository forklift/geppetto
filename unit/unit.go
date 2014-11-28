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

	return Parse(file)
}

func Parse(reader io.Reader) (*Unit, error) {
	u := &Unit{}

	err := ini.NewDecoder(reader).Decode(u)
	if err != nil {
		return nil, err
	}

	return u, u.Prepare()
}

type Unit struct {
	//Unit is the "unit" files for Geppeto.

	Meta    Meta
	Service Service

	//The actuall process on system and it's attributes.
	process *exec.Cmd
}

func (u *Unit) Prepare() error {

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

	return nil
}

func (u *Unit) Start() {}
func (u *Unit) Wait()  {}
