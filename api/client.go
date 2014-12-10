package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.Status, e.Message)
}

func newError(status int, body []byte) *Error {
	return &Error{Status: status, Message: string(body)}
}

type client struct {
	endpoint *url.URL
	httpConn *http.Client
}

func (client *client) getURL(path string) string {
	url := strings.TrimRight(client.endpoint.String(), "/")
	if client.endpoint.Scheme == "unix" {
		url = ""
	}
	return fmt.Sprintf("%s%s", url, path)
}

func (client *client) do(method string, path string, data interface{}) ([]byte, int, error) {
	var body io.Reader

	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return nil, -1, err
		}
		body = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequest(method, client.getURL(path), body)
	if err != nil {
		return nil, -1, err
	}

	req.Header.Set("User-Agent", "microdocker")
	req.Header.Set("Content-Type", "Content-Type: application/json")

	var res *http.Response
	if client.endpoint.Scheme == "unix" {
		dial, err := net.Dial(client.endpoint.Scheme, client.endpoint.Path)
		if err != nil {
			return nil, -1, err
		}
		newConn := httputil.NewClientConn(dial, nil)
		res, err = newConn.Do(req)
		defer newConn.Close()

	} else {
		res, err = client.httpConn.Do(req)
	}
	if err != nil {
		status := 0
		if res != nil {
			status = res.StatusCode
		}
		return nil, status, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, res.StatusCode, err
	}

	replay, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, err
	}
	return replay, res.StatusCode, nil
}

/*
//Stream

type jsonMessage struct {
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
	Stream   string `json:"stream,omitempty"`
}

func (client *client) stream(method, path string, headers map[string]string, in io.Reader, out io.Writer) error {

	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, client.getURL(path), in)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "microdocker")
	req.Header.Set("Content-Type", "Content-Type: application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	var res *http.Response
	if client.endpoint.Scheme == "unix" {
		dial, err := net.Dial(client.endpoint.Scheme, client.endpoint.Path)
		if err != nil {
			return err
		}
		newConn := httputil.NewclientConn(dial, nil)
		res, err = newConn.Do(req)
		defer newConn.Close()

	} else {
		res, err = client.httpConn.Do(req)
	}
	if err != nil {
		return err
	}

	defer res.Body.Close()
	contentType := res.Header.Get("Content-Type")
	//TODO:  Check this for: POST /containers/(id)/start `|| contentType == "text/plain" `
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return &Error{Status: int(res.StatusCode), Message: string(body)}
	}

	if contentType == "application/json" {

		document := json.NewDecoder(res.Body)
		for {
			var msg jsonMessage
			if err := document.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if msg.Stream != "" {
				fmt.Fprint(out, msg.Stream)
			} else if msg.Progress != "" {
				fmt.Fprintf(out, "%s %s\r", msg.Status, msg.Progress)
			} else if msg.Error != "" {
				return errors.New(msg.Error)
			}
			if msg.Status != "" {
				fmt.Fprintln(out, msg.Status)
			}
		}
	}

	if _, err := io.Copy(out, res.Body); err != nil {
		return err
	}
	if ender, ok := out.(interface {
		End() (int, error)
	}); ok {
		if _, err := ender.End(); err != nil {
			return err
		}
	}
	return nil
} */
