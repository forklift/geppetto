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

//       TODO: For "unix" io.WriteCloser, syscall.Mknod(fullPath, syscall.S_IFIFO|0666, 0)
type WriteCloser struct {
	io.WriteCloser
	UrlPath
}

func (i *WriteCloser) Connect() error {

	var err error

	//TODO: Close these connections on failur or exit!
	if i.path == nil || i.path.String() == "" {
		i.WriteCloser, err = os.OpenFile(os.DevNull, os.SEEK_SET, os.ModeType)
		return err
	}

	switch i.path.Scheme {

	case "file":
		//TODO: is it better to use a ReadOnly vs WriteOnly for stdin/stdout? or not worth the lines of code.
		i.WriteCloser, err = os.Open(i.path.Path)

	//TODO: Check if they actually work, add more if possible.
	case "tcp", "unix", "unixgram", "udp":
		i.WriteCloser, err = net.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"))

	case "tls":
		tlsConf := &tls.Config{InsecureSkipVerify: true}
		i.WriteCloser, err = tls.Dial(i.path.Scheme, strings.TrimPrefix(i.path.String(), i.path.Scheme+"://"), tlsConf)

	default:
		err = errors.New("Unsupported Protocol.")
	}

	return err
}
