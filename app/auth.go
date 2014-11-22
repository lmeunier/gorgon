package app

import (
	"errors"
	"github.com/mxk/go-imap/imap"
	"reflect"
	"time"
)

var (
	Authenticators = map[string]reflect.Value{
		"test": reflect.ValueOf(NewTestAuthenticator),
		"imap": reflect.ValueOf(NewImapAuthenticator),
	}
)

func NewAuthenticator(app GorgonApp, name string) (authenticator Authenticator, err error) {
	if _, ok := Authenticators[name]; !ok {
		err := errors.New("Authenticator '" + name + "' does not exist.")
		return nil, err
	}

	params := []reflect.Value{reflect.ValueOf(app)}
	results := Authenticators[name].Call(params)

	authenticator = results[0].Interface().(Authenticator)
	if results[1].Interface() != nil {
		err = results[1].Interface().(error)
	}
	return
}

type Authenticator interface {
	Authenticate(username, password string) error
}

type TestAuthenticator struct {
	global_password string
}

func (a TestAuthenticator) Authenticate(username, password string) (err error) {
	if a.global_password != password {
		err = errors.New("TestAuthenticator: authentication failed")
	}
	return
}

func NewTestAuthenticator(app GorgonApp) (Authenticator, error) {
	global_password, ok := app.Config.Get("auth:test", "global_password")
	if !ok {
		panic("'global_password' variable missing from 'auth:test' section")
	}

	authenticator := TestAuthenticator{global_password}
	return authenticator, nil
}

type ImapAuthenticator struct {
	server string
}

func (a ImapAuthenticator) Authenticate(username, password string) (err error) {
	client, err := imap.Dial(a.server)
	if client != nil {
		defer client.Logout(30 * time.Second)
	}
	if err != nil {
		return
	}

	if client.Caps["STARTTLS"] {
		if _, err = client.StartTLS(nil); err != nil {
			return
		}
	}

	if client.State() == imap.Login {
		if _, err = client.Login(username, password); err != nil {
			return
		}
	}

	return
}

func NewImapAuthenticator(app GorgonApp) (Authenticator, error) {
	server, ok := app.Config.Get("auth:imap", "server")
	if !ok {
		panic("'server' variable missing from 'auth:imap' section")
	}

	authenticator := ImapAuthenticator{server}
	return authenticator, nil
}
