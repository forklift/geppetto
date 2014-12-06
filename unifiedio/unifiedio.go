package unifiedio

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
type IO struct {
	io.ReadWriteCloser
	path *url.URL
}

func New(path string) (*IO, error) {

	io := &IO{}

	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	io.path = url
	return io, nil
}

func (i *IO) Connect() error {

	var (
		err error
		o   io.ReadWriteCloser
	)

	//TODO: Close these connections on failur or exit!
	if i.path == nil {
		i.ReadWriteCloser, err = os.Open(os.DevNull)
		return err
	}

	switch i.path.Scheme {

	case "file":
		//TODO: is it better to use a ReadOnly vs WriteOnly for stdin/stdout? or not worth the lines of code.
		o, err = os.Open(i.path.Path)

	//TODO: Check if they actually work, add more if possible.
	case "tcp", "unix", "unixgram", "udp":
		o, err = net.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"))

	case "tls":
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		o, err = tls.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"), tlsConf)

	default:
		err = errors.New("Unsupported Protocol.")
	}

	//TODO: err == nil && i.Stream == nil

	if err == nil {
		i = &IO{o, i.path}
	}

	return err
}
