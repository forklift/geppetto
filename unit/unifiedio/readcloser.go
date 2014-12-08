package unifiedio

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"os"
	"strings"
)

//FIXME: Do we need a PTY? Perhaps consider: https://github.com/kr/pty

type ReadCloser struct {
	io.ReadCloser
	UrlPath
}

func (i *ReadCloser) Connect() error {

	//TODO: Close these connections on failur or exit!
	var err error

	if i.path == nil || i.path.String() == "" {
		i.ReadCloser, err = os.OpenFile(os.DevNull, os.SEEK_SET, os.ModeType)
		return err
	}

	switch i.path.Scheme {

	case "file":
		//TODO: is it better to use a ReadOnly vs WriteOnly for stdin/stdout? or not worth the lines of code.
		i.ReadCloser, err = os.Open(i.path.Path)

	//TODO: Check if they actually work, add more if possible.
	case "tcp", "unix", "unixgram", "udp":
		i.ReadCloser, err = net.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"))

	case "tls":
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		i.ReadCloser, err = tls.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"), tlsConf)

	default:
		err = errors.New("Unsupported Protocol.")
	}

	return err
}
