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
	io.ReadWriteCloser
	path *url.URL
}

func (i *UnifiedIO) UnmarshalBinary(raw []byte) error {
	var err error
	i.path, err = url.Parse(string(raw))
	return err
}

func (i *UnifiedIO) Prepare() error {
	//FIXME: Not sure if this is a good idea.
	if i.ReadWriteCloser != nil {
		return nil
	}

	var err error

	//TODO: Close these connections on failur or exit!
	if i.path == nil {
		i.ReadWriteCloser, err = os.Open(os.DevNull)
		return err
	}

	switch i.path.Scheme {

	case "file":
		//TODO: is it better to use a ReadOnly vs WriteOnly for stdin/stdout? or not worth the lines of code.
		i.ReadWriteCloser, err = os.Open(i.path.Path)

	//TODO: Check if they actually work, add more if possible.
	case "tcp", "unix", "unixgram", "udp":
		i.ReadWriteCloser, err = net.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"))

	case "tls":
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		i.ReadWriteCloser, err = tls.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"), tlsConf)

	default:
		err = errors.New("Unsupported Protocol.")
	}

	//TODO: err == nil && i.Stream == nil

	return err
}
