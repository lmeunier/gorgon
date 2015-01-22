package app

import (
	"crypto/tls"
	"testing"

	"github.com/mxk/go-imap/imap"
	"github.com/mxk/go-imap/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthenticator(t *testing.T) {
	// create our app
	app := NewApp("../tests/gorgon.ini")

	// try to use an unknown authenticator
	authenticator, err := NewAuthenticator(app, "unknown authenticator name")
	assert.Nil(t, authenticator)
	assert.Error(t, err)
}

func TestTestAuthenticator(t *testing.T) {
	// create our app
	app := NewApp("../tests/gorgon.ini")

	// create a Test authenticator
	authenticator, err := NewAuthenticator(app, "test")
	assert.IsType(t, TestAuthenticator{}, authenticator)
	testAuthenticator := authenticator.(TestAuthenticator)
	assert.Equal(t, "secretpasswordfortests", testAuthenticator.GlobalPassword)
	assert.NoError(t, err)

	// try to authenticate with the good password
	err = testAuthenticator.Authenticate("foobar", "secretpasswordfortests")
	assert.NoError(t, err)

	// try to authenticate with a wrong password
	err = testAuthenticator.Authenticate("foobar", "wrong password")
	assert.Error(t, err)
}

func TestImapAuthenticator(t *testing.T) {
	// create our app
	app := NewApp("../tests/gorgon.ini")

	// create an Imap authenticator
	authenticator, err := NewAuthenticator(app, "imap")
	assert.IsType(t, ImapAuthenticator{}, authenticator)
	imapAuthenticator := authenticator.(ImapAuthenticator)
	assert.Equal(t, "imap.example.com", imapAuthenticator.Server)
	assert.True(t, imapAuthenticator.TLSVerifyCert)
	assert.NoError(t, err)
}

func TestImapAuthenticate(t *testing.T) {
	var (
		c       *imap.Client
		s       *mock.T
		errMock error
		errAuth error
	)

	// try to authenticate (Auth ok, no TLS)
	s = mock.Server(t,
		"S: * OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] Server ready",
		"C: A1 LOGIN \"alice\" \"verysecret\"",
		"S: A1 OK LOGIN completed",
		"C: A2 CAPABILITY",
		"S: * CAPABILITY IMAP4rev1 AUTH=PLAIN",
		"S: A2 OK Thats all",
	)
	c, _ = s.Dial()
	errAuth = ImapAuthenticate(c, "alice", "verysecret", nil)
	assert.NoError(t, errAuth)
	s.Join(errMock)

	// try to authenticate (Auth ko, no TLS)
	s = mock.Server(t,
		"S: * OK [CAPABILITY IMAP4rev1 AUTH=PLAIN] Server ready",
		"C: A1 LOGIN \"alice\" \"bad password\"",
		"S: A1 NO [AUTHENTICATIONFAILED] Authentication failed",
	)
	c, _ = s.Dial()
	errAuth = ImapAuthenticate(c, "alice", "bad password", nil)
	assert.Error(t, errAuth)
	s.Join(errMock)

	// try to authenticate (Auth ok, TLS skip verify)
	s = mock.Server(t,
		"S: * OK [CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS] Server ready",
		"C: A1 STARTTLS",
		"S: A1 OK Begin TLS Negotation now",
		mock.STARTTLS,
		"C: A2 CAPABILITY",
		"S: * CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS",
		"S: A2 OK Thats all",
		"C: A3 LOGIN \"alice\" \"verysecret\"",
		"S: A3 OK LOGIN completed",
		"C: A4 CAPABILITY",
		"S: * CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS",
		"S: A4 OK Thats all",
	)
	c, _ = s.Dial()
	errAuth = ImapAuthenticate(c, "alice", "verysecret", &tls.Config{InsecureSkipVerify: true})
	assert.NoError(t, errAuth)
	s.Join(errMock)

	// try to authenticate (Auth ok, TLS verify)
	s = mock.Server(t,
		"S: * OK [CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS] Server ready",
		"C: A1 STARTTLS",
		"S: A1 OK Begin TLS Negotation now",
	)
	c, _ = s.Dial()
	errAuth = ImapAuthenticate(c, "alice", "verysecret", nil)
	assert.Error(t, errAuth)
	s.Join(errMock)
}
