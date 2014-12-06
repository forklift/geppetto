package unit

import (
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/forklift/geppetto/unit/sys/group"
)

type Service struct {

	//Configures the process start-up type for this service unit. One of : `simple`, `forking`, `oneshot`.
	// `simple` : It is expected that the process configured with `ExecStart` is the main process of the service.
	// `oneshot`: Similar to simple; however, it is expected that the process has to exit before systemd starts follow-up units.
	// `forking`: It is expected that the process configured with `ExecStart` will call `fork()` as part of its start-up. The parent process is expected to exit when start-up is complete and all communication channels are set up. The child continues to run as the main daemon process. This is the behavior of traditional UNIX daemons. You MUST also use the `PIDFile` option, so that systemd can identify the main process of the daemon. Geppetto will proceed with starting follow-up units as soon as the parent process exits.
	Type string

	// A space-separated list of required prerequiests. if the units listed here are not started already, they will not be started and the transaction will fail immediately.
	Prerequests []string

	// Similar to Prerequests, but the opposite. if units listed here are already started, this transaction will fail immediately.
	Conflicts []string

	// A space-separated list of required units to start with this unit, in any order. if any of these units fails, this unit will be cancneld.
	Requires []string

	// A space-separated list of required units to start with this unit, in any order. if any of these units fails, this unit will be NOT cancled if any of these units
	// fail after a successful startup.
	Wants []string

	//TODO: Validate if Before and After values exists in Requires or Wants.

	// A space-separated list of required units to start before this unit is started. The units must exist in Requires or Wants.
	Before []string
	// A space-separated list of required units to start before this unit is started. The units must exist in Requires or Wants.
	After []string

	//Similar in to Preequires. However in addition to this behavior, if any of the units listed suddenly disappears or fails, this unit stops.
	BindsTo []string

	// Takes an absolute directory path. Sets the working directory for executed processes. If not set, defaults to the root directory when systemd is running as a system instance and the respective user's home directory if run as user.
	WorkingDirectory string
	// Sets the Unix user that the processes are executed as. Takes a single user name or ID as argument. If no user is set, the default user will be chosen.
	User string

	// Sets the Unix group that the processes are executed as. Takes a single group name or ID as argument. If no user is set, the default user's group will be chosen.
	Group string

	//Sets the root" for the command, requires "root" user.
	Chroot string

	// Commands with their arguments that are executed when this service is started.
	// This will not be run in a shell, so there is NO pipe or redirects, use the Stin and Stdout.
	ExecStart string

	//The stdin, stdout, and stderr are network aware and accepts the following protocoles:
	// "file", "tcp", "unix", "unixgram", "udp", and "tls".
	// examples: `file://var/log/Unit`, `tls://feeds.example.tld:443`, `unix://var/run/myunit.sock`
	// if nothing given, a null device will be used.
	Stdin  *UnifiedIO //io.Reader
	Stdout *UnifiedIO //io.Writer

	//TODO: Logging.
	Stderr *UnifiedIO //io.Writer

}

func (s *Service) Prepare() error {

	var err error

	if filepath.Base(s.ExecStart) == s.ExecStart {
		s.ExecStart, err = exec.LookPath(s.ExecStart)
	}

	if err != nil {
		return err
	}

	for _, i := range []*UnifiedIO{s.Stdin, s.Stdout, s.Stderr} {

		if i == nil {
			i = &UnifiedIO{}
		}
	}

	return err
}

func (s *Service) ConnectIO() error {
	for _, i := range []*UnifiedIO{s.Stdin, s.Stdout, s.Stderr} {

		if err := i.Connect(); err != nil {
			return err
		}

	}
	return nil
}

func (s *Service) CloseIO() []error {
	var errs []error

	//Close connections.
	for _, i := range []*UnifiedIO{s.Stdin, s.Stdout, s.Stderr} {
		if i != nil {
			err := i.Close()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (s *Service) BuildCredentails() (*syscall.Credential, error) {

	var err error

	cred := &syscall.Credential{}
	usr := &user.User{}

	if s.User == "" {
		usr, err = user.Current()
	} else {
		usr, err = user.Lookup(s.User)
	}

	if err != nil {
		return nil, err
	}

	Uid, err := strconv.ParseUint(usr.Uid, 10, 0)
	Gid, err := strconv.ParseUint(usr.Gid, 10, 0)

	cred.Uid = uint32(Uid)
	cred.Gid = uint32(Gid)

	if s.Group != "" {
		grp, err := group.Lookup(s.Group)
		if err != nil {
			return nil, err
		}

		Gid, err := strconv.ParseUint(grp.Gid, 10, 0)
		cred.Gid = uint32(Gid)
	}

	return cred, nil
}
