package unit

import "github.com/omeid/semver"

type Meta struct {
	// The version of Gepetto that this Unit was written for, Gepetto will try it's best
	// to treat the Unit like the version specified.
	// Gepetto will refuse to run a Unit If version is missing, invalid, or incompatible.
	Geppetto        string
	GeppettoVersion *semver.Version
	// A free-form string describing of the Unit. This is intended for use in UIs to show descriptive information along with the service name.
	// The description should contain a name that means something to the end user.
	// Good example: `Nginx Server and Reverse Proxy`.
	// Bad example: `Nginx` (too specific and meaningless for people who do not know Nginx).
	// Bad example: `Web Server` (too generic).
	// Bad example: `Nginx high-performance light-weight HTTP and Reverse Proxy Server` (too long).
	Description string
}

func (m *Meta) Prepare() error {

	var err error

	m.GeppettoVersion, err = semver.NewVersion(m.Geppetto)

	//TODO: Check compatibility here.
	return err

}
