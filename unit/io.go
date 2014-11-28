package unit

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

//FIXME: Do we need a PTY? Perhaps consider: https://github.com/kr/pty
type UnifiedIO struct {
	path   *url.URL
	Stream io.ReadWriter
}

func (i *UnifiedIO) UnmarshalBinary(raw []byte) error {
	var err error
	i.path, err = url.Parse(string(raw))
	return err
}

func (i *UnifiedIO) Prepare() error {
	//FIXME: Not sure if this is a good idea.
	if i.Stream != nil {
		return nil
	}

	if i.path.Path == "os.Stdout" {
		i.Stream = os.Stdout
		return nil
	}

	var err error
	switch i.path.Scheme {

	case "file":
		//TODO: is it better to use a ReadOnly vs WriteOnly for stdin/stdout? or not worth the lines of code.
		i.Stream, err = os.Open(i.path.Path)

	//TODO: Check if they actually work, add more if possible.
	case "tcp", "unix", "unixgram", "udp":
		i.Stream, err = net.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"))

	case "tls":
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		i.Stream, err = tls.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"), tlsConf)

	default:
		err = errors.New("Unsupported Protocol.")
	}

	//TODO: err == nil && i.Stream == nil

	return err
}
