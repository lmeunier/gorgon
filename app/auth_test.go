package app

import (
	"testing"

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
