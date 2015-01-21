package app

import (
	"crypto/tls"
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

// Authenticator is an interface representing a method to authenticate a user.
// Only one function must be implemented: Authenticate(username, password
// string) that returns nil when the username/password pair is correct or an
// error.
type Authenticator interface {
	Authenticate(username, password string) error
}

// NewAuthenticator returns an Authenticator based on the provided name. The
// authenticator is configured from app.Config.
func NewAuthenticator(app GorgonApp, name string) (authenticator Authenticator, err error) {
	if _, ok := Authenticators[name]; !ok {
		err := errors.New("Authenticator '" + name + "' does not exist.")
		return nil, err
	}

	// call the function to create the authenticator, the app is the only
	// argument passed to the function
	params := []reflect.Value{reflect.ValueOf(app)}
	results := Authenticators[name].Call(params)

	// cast the first returned value as an Authenticator interface ...
	authenticator = results[0].Interface().(Authenticator)
	if results[1].Interface() != nil {
		// ... and the second returned value as an error
		err = results[1].Interface().(error)
	}
	return
}

// TestAuthenticator implements the Authenticator interface and is a very
// simple auth method whose first goal is to test a Gorgon app. This
// authenticator should not be used in production.
//
// An example configuration looks like this:
//
// [global]
// ...
// auth = test
//
// [auth:test]
// global_password = myverysecretpassword
//
type TestAuthenticator struct {
	GlobalPassword string // global password to authenticate all users
}

// Authenticate uses a global password to authenticate all users.
func (a TestAuthenticator) Authenticate(username, password string) (err error) {
	if a.GlobalPassword != password {
		err = errors.New("TestAuthenticator: authentication failed")
	}
	return
}

// NewTestAuthenticator returns a populated TestAuthenticator.
func NewTestAuthenticator(app GorgonApp) (Authenticator, error) {
	app.Logger.Warning("You are using the test authenticator. Do *NOT* use this " +
		"authenticator in a production environment.")
	global_password, ok := app.Config.Get("auth:test", "global_password")
	if !ok {
		panic("'global_password' variable missing from 'auth:test' section")
	}

	authenticator := TestAuthenticator{global_password}
	return authenticator, nil
}

// ImapAuthenticator implements the Authenticator interface to authenticate
// users against an Imap server. The username (email) and password are passed
// without modification to the Imap server.
//
// An example configuration looks like this:
//
// [global]
// ...
// auth = imap
//
// [auth:imap]
// server = imap.example.com
// verify_cert = true
//
type ImapAuthenticator struct {
	Server        string // hostname or address of an Imap server
	TLSVerifyCert bool   // should verify the certificate presented by the server
}

// Authenticate uses an Imap server to authenticate users. The username (email)
// and password are passed without modification to the Imap server.
func (a ImapAuthenticator) Authenticate(username, password string) (err error) {
	client, err := imap.Dial(a.Server)
	if client != nil {
		defer client.Logout(30 * time.Second)
	}
	if err != nil {
		return
	}

	tlsConfig := tls.Config{
		InsecureSkipVerify: a.TLSVerifyCert,
	}

	return ImapAuthenticate(client, username, password, &tlsConfig)
}

// ImapAuthenticate tries to authenticate a user. If the IMAP server advertive
// the STARTTLS capability, the connection switches to TLS and use the provided
// *tls.Config. If the authentication is successful, returns nil, else returns
// an error.
func ImapAuthenticate(client *imap.Client, username, password string, tlsConfig *tls.Config) (err error) {
	if client.Caps["STARTTLS"] {
		if _, err = client.StartTLS(tlsConfig); err != nil {
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

// NewImapAuthenticator returns a populated ImapAuthenticator
func NewImapAuthenticator(app GorgonApp) (Authenticator, error) {
	server, ok := app.Config.Get("auth:imap", "server")
	if !ok {
		panic("'server' variable missing from 'auth:imap' section")
	}
	verifyCert, ok := app.Config.Get("auth:imap", "verify_cert")
	if !ok {
		verifyCert = "true"
	}

	authenticator := ImapAuthenticator{
		Server:        server,
		TLSVerifyCert: verifyCert == "true",
	}
	return authenticator, nil
}
