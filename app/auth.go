package app

import (
	"errors"
	"reflect"
)

var (
	Authenticators = map[string]reflect.Value{
		"test": reflect.ValueOf(NewTestAuthenticator),
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
