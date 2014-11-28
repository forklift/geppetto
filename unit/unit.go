package unit

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

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

type Meta struct {
	// The version of Gepetto that this Unit was written for, Gepetto will try it's best
	// to treat the Unit like the version specified.
	// Gepetto will refuse to run a Unit If version is missing, invalid, or incompatible.
	Gepetto string

	// A free-form string describing of the Unit. This is intended for use in UIs to show descriptive information along with the service name.
	// The description should contain a name that means something to the end user.
	// Good example: `Nginx Server and Reverse Proxy`.
	// Bad  example: `Nginx`      (too specific and meaningless for people who do not know Nginx).
	// Bad  example: `Web Server` (too generic).
	// Bad  example: `Nginx high-performance light-weight HTTP and Reverse Proxy Server` (too long).
	Description string
}

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

	// Commands with their arguments that are executed when this service is started.
	// This will not be run in a shell, so there is NO pipe or redirects, variables. If your program needs it, fix your program, or write a wrapper.
	ExecStart string

	//These commens will be
	Stdin  UnifiedIO //io.Reader
	Stdout UnifiedIO //io.Writer
}

type Unit struct {
	//Unit is the "unit" files for Geppeto.

	Meta    Meta
	Service Service

	//The actuall process on system.
	process *exec.Cmd
}

func (u *Unit) Prepare() error {
	return nil
}

func (u *Unit) Start() {}
func (u *Unit) Wait()  {}
