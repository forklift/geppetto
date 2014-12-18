package unit

import (
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/forklift/geppetto/unit/sys/group"
	"github.com/forklift/geppetto/unit/unifiedio"
	"github.com/mattn/go-shellwords"
	//TODO: Port to GO 1.4, us /x/sys/unix
)

type Process struct {

	//Configures the process start-up type for this service unit. One of : `simple`, `forking`, `oneshot`.
	// `simple` : It is expected that the process configured with `ExecStart` is the main process of the service.
	// `oneshot`: Similar to simple; however, it is expected that the process has to exit before systemd starts follow-up unitp.
	// `forking`: It is expected that the process configured with `ExecStart` will call `fork()` as part of its start-up. The parent process is expected to exit when start-up is complete and all communication channels are set up. The child continues to run as the main daemon process. This is the behavior of traditional UNIX daemonp. You MUST also use the `PIDFile` option, so that systemd can identify the main process of the daemon. Geppetto will proceed with starting follow-up units as soon as the parent process exitp.
	Type string

	// Takes an absolute directory path. Sets the working directory for executed processep. If not set, defaults to the root directory when systemd is running as a system instance and the respective user's home directory if run as user.
	WorkingDirectory string
	// Sets the Unix user that the processes are executed ap. Takes a single user name or ID as argument. If no user is set, the default user will be chosen.
	User string

	// Sets the Unix group that the processes are executed ap. Takes a single group name or ID as argument. If no user is set, the default user's group will be chosen.
	Group string

	//Sets the root" for the command, requires "root" user.
	Chroot string

	// Commands with their arguments that are executed when this service is started.
	// This will not be run in a shell, so there is NO pipe or redirects, use the Stin and Stdout.
	ExecStart string

	//The stdin, stdout, and stderr are network aware and accepts the following protocoles:
	// "file", "tcp", "unix", "unixgram", "udp", and "tls".
	// examples: `file://var/log/Unit`, `tls://feedp.example.tld:443`, `unix://var/run/myunit.sock`
	// if nothing given, a null device will be used.
	Stdin string
	stdin *unifiedio.ReadCloser

	Stdout string
	stdout *unifiedio.WriteCloser

	Stderr string
	stderr *unifiedio.WriteCloser

	//The actuall process on system and it's attributep.
	process           *exec.Cmd
	processExitStatus chan error
}

func (p *Process) Prepare() error {

	var err error

	if filepath.Base(p.ExecStart) == p.ExecStart {
		p.ExecStart, err = exec.LookPath(p.ExecStart)
	}

	if err != nil {
		return err
	}

	p.stdin = &unifiedio.ReadCloser{}
	p.stdout = &unifiedio.WriteCloser{}
	p.stderr = &unifiedio.WriteCloser{}

	for url, io := range map[string]unifiedio.IO{p.Stdin: p.stdin, p.Stdout: p.stdout, p.Stderr: p.stderr} {
		err := io.SetPath(url)
		if err != nil {
			return nil
		}
	}

	cred, err := p.BuildCredentails()

	p.processExitStatus = make(chan error, 1)

	//TODO: We need to export, $HOME, $USER, $EDITOR, et all. Should be done at Engine level, or deamon?
	shellwords.ParseEnv = true
	cmd, err := shellwords.Parse(p.ExecStart)
	if err != nil {
		return err
	}

	p.process = exec.Command(cmd[0])
	p.process.Args = cmd

	p.process.Dir = p.WorkingDirectory

	p.process.SysProcAttr = &syscall.SysProcAttr{
		Setsid:     true,
		Credential: cred,
	}

	if p.Chroot != "" && p.Chroot != "/" {
		p.process.SysProcAttr.Chroot = p.Chroot
	}
	return err
}

func (p *Process) ConnectIO() error {

	for _, i := range []unifiedio.IO{p.stdin, p.stdout, p.stderr} {

		if err := i.Connect(); err != nil {
			p.CloseIO()
			return err
		}
	}

	p.process.Stdin = p.stdin
	p.process.Stdout = p.stdout
	p.process.Stderr = p.stderr

	return nil
}

func (p *Process) CloseIO() []error {
	var errs []error

	//Close connectionp.
	for _, i := range []unifiedio.IO{p.stdin, p.stdout, p.stderr} {
		if i != nil {
			err := i.Close()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (p *Process) BuildCredentails() (*syscall.Credential, error) {

	var err error

	cred := &syscall.Credential{}
	usr := &user.User{}

	if p.User == "" {
		usr, err = user.Current()
	} else {
		usr, err = user.Lookup(p.User)
	}

	if err != nil {
		return nil, err
	}

	Uid, err := strconv.ParseUint(usr.Uid, 10, 0)
	Gid, err := strconv.ParseUint(usr.Gid, 10, 0)

	cred.Uid = uint32(Uid)
	cred.Gid = uint32(Gid)

	if p.Group != "" {
		grp, err := group.Lookup(p.Group)
		if err != nil {
			return nil, err
		}

		Gid, err := strconv.ParseUint(grp.Gid, 10, 0)
		cred.Gid = uint32(Gid)
	}

	return cred, nil
}
