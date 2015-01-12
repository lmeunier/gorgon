package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
	gorgon "github.com/lmeunier/gorgon/app"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

func GetAuthCookie(username string, codecs ...securecookie.Codec) (*http.Cookie, error) {
	var authCookie http.Cookie
	decodedValue := make(map[interface{}]interface{})
	decodedValue["authenticated_as"] = username
	encodedValue, err := securecookie.EncodeMulti("persona-auth", decodedValue, codecs...)
	if err != nil {
		return nil, err
	}
	authCookie = http.Cookie{
		Name:  "persona-auth",
		Value: encodedValue,
	}
	return &authCookie, nil
}

func TestSupportDocument(t *testing.T) {
	// create our app
	app := gorgon.NewApp("tests/gorgon.ini")
	logging.SetBackend(logging.NewMemoryBackend(0))

	req, _ := http.NewRequest("GET", "", nil)
	w := httptest.NewRecorder()

	handle := gorgon.GorgonHandler{&app, gorgon.SupportDocumentHandler}
	handle.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusOK, "SupportDocument should return a 200 HTTP status")

	if contentType := w.Header().Get("content-type"); contentType != "application/json" {
		t.Error("No 'Content-Type: application/json' header")
	}

	var supportDocument map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &supportDocument); err != nil {
		t.Error("Error when decoding SupportDocument response from JSON: %s", err.Error())
	}

	assert.Equal(t,
		supportDocument["authentication"],
		"/.well-known/browserid/_gorgon/authentication",
		"The authentication URL in the SupportDocument must match the given URL",
	)

	assert.Equal(t,
		supportDocument["provisioning"],
		"/.well-known/browserid/_gorgon/provisioning",
		"The provisioning URL in the SupportDocument must match the given URL",
	)

	publicKey := supportDocument["public-key"].(map[string]interface{})
	assert.Equal(t, publicKey["algorithm"], "RS",
		"The public key algorithm must be RS",
	)
	assert.Equal(t, publicKey["e"], "65537",
		"The public key exponent must match the exponent of the public key in the tests/ folder",
	)
	assert.Equal(t, publicKey["n"], "23889456486321623473918665255343177635128710173157669424372875359372320736455301355970875593938909218087760986617640898578679844226631652037745248513549355637358220697959372250835697775454508812368942374589593646559767646333952478128009848675984203746915093496186538141115897660344039231014275918654719281984603972728249059019777484941535089356009107883030164047835349598045162040165036179729458456956010346765443578628841250841421231205777808469647531245961509586074089345598248954243867408603319475227403111600221142603559700387465893230884454715818763876460572694605425473855790697450998747164996944357297066585551",
		"The public key modulus must match the modulus of the public key in the tests/ folder",
	)
}

func TestProvisioningPage(t *testing.T) {
	// create our app
	app := gorgon.NewApp("tests/gorgon.ini")
	logging.SetBackend(logging.NewMemoryBackend(0))

	// the handle that will be tested
	handle := gorgon.GorgonHandler{&app, gorgon.ProvisioningHandler}

	// cookie used to authenticate our user
	authCookie, err := GetAuthCookie("user@example.com", app.SessionStore.Codecs...)
	assert.NoError(t, err)

	var (
		req  *http.Request
		w    *httptest.ResponseRecorder
		body string
	)

	// TEST: no auth cookie
	req, _ = http.NewRequest("GET", "", nil)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.Contains(t, body, "navigator.id.raiseProvisioningFailure",
		"raiseProvisioningFailure must be called because the user is not yet authenticated",
	)

	// TEST: with auth cookie
	req, _ = http.NewRequest("GET", "", nil)
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.NotContains(t, body, "navigator.id.raiseProvisioningFailure",
		"raiseProvisioningFailure must *NOT* be called because the user is authenticated",
	)
	assert.Contains(t, body, "navigator.id.registerCertificate",
		"registerCertificate must be called",
	)

	// TEST: malformed cookie
	malformedAuthCookie := authCookie
	malformedAuthCookie.Value = malformedAuthCookie.Value + "BAD"
	req, _ = http.NewRequest("GET", "", nil)
	req.AddCookie(malformedAuthCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.Contains(t, body, "navigator.id.raiseProvisioningFailure",
		"raiseProvisioningFailure must be called because the user is not yet authenticated",
	)
}

func TestAuthenticationPage(t *testing.T) {
	// create our app
	app := gorgon.NewApp("tests/gorgon.ini")
	logging.SetBackend(logging.NewMemoryBackend(0))

	// the handle that will be tested
	handle := gorgon.GorgonHandler{&app, gorgon.AuthenticationHandler}

	// cookie used to authenticate our user
	authCookie, err := GetAuthCookie("user@example.com", app.SessionStore.Codecs...)
	assert.NoError(t, err)

	var (
		req  *http.Request
		w    *httptest.ResponseRecorder
		data url.Values
		body string
	)

	// TEST: display form
	req, _ = http.NewRequest("GET", "", nil)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.Contains(t, body, "<input id=\"input_email\" type=\"text\" name=\"email\"",
		"The authentication page must display a form with an Email field",
	)
	assert.Contains(t, body, "<input id=\"input_password\" type=\"password\" name=\"password\"",
		"The authentication page must display a form with a Password field",
	)
	assert.NotContains(t, body, "Authentication failed!",
		"No error message when displaying the form for the first time (on a GET request)",
	)

	// TEST: submit form with bad credentials
	data = url.Values{}
	data.Set("email", "badpassword@example.com")
	data.Add("password", "badpassword")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.Contains(t, body, "<input id=\"input_email\" type=\"text\" name=\"email\"",
		"The authentication page must display a form with an Email field",
	)
	assert.Contains(t, body, "<input id=\"input_password\" type=\"password\" name=\"password\"",
		"The authentication page must display a form with a Password field",
	)
	assert.Contains(t, body, "Authentication failed!",
		"The error message must be displayed",
	)

	// TEST: submit form with good credentials
	data = url.Values{}
	data.Set("email", "user@example.com")
	data.Add("password", "secretpasswordfortests")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.NotContains(t, body, "<form method=\"POST\">",
		"The authentication page must *NOT* display a form",
	)
	assert.Contains(t, body, "navigator.id.completeAuthentication",
		"completeAuthentication must be called",
	)

	// create an http.Response to decode cookies
	resp := http.Response{
		Header: w.Header(),
	}
	var incomingCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "persona-auth" {
			incomingCookie = cookie
		}
	}
	assert.NotNil(t, incomingCookie, "The 'persona-auth' cookie must be set")

	// try to decode the secure cookie
	decodedValue := make(map[interface{}]interface{})
	err = securecookie.DecodeMulti(incomingCookie.Name, incomingCookie.Value, &decodedValue, app.SessionStore.Codecs...)
	if assert.NoError(t, err) {
		assert.Equal(t, decodedValue["authenticated_as"], "user@example.com",
			"The username in the cookie must be the same as the one POSTed",
		)
	}

	// TEST: already authenticated
	req, _ = http.NewRequest("GET", "", nil)
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	body = w.Body.String()
	assert.NotContains(t, body, "<form method=\"POST\">",
		"The authentication page must *NOT* display a form",
	)
	assert.Contains(t, body, "navigator.id.completeAuthentication",
		"completeAuthentication must be called",
	)
}

func TestCheckAuthenticated(t *testing.T) {
	// create our app
	app := gorgon.NewApp("tests/gorgon.ini")
	logging.SetBackend(logging.NewMemoryBackend(0))

	// the handle that will be tested
	handle := gorgon.GorgonHandler{&app, gorgon.CheckAuthenticatedHandler}

	// cookie used to authenticate our user
	authCookie, err := GetAuthCookie("user@example.com", app.SessionStore.Codecs...)
	assert.NoError(t, err)

	var (
		req *http.Request
		w   *httptest.ResponseRecorder
	)

	// TEST: not authenticated
	req, _ = http.NewRequest("GET", "", nil)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusForbidden)

	// TEST: good cookie
	req, _ = http.NewRequest("GET", "", nil)
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusOK)

	// TEST: malformed cookie
	malformedAuthCookie := authCookie
	malformedAuthCookie.Value = malformedAuthCookie.Value + "BAD"
	req, _ = http.NewRequest("GET", "", nil)
	req.AddCookie(malformedAuthCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestGenerateCertificate(t *testing.T) {
	// create our app
	app := gorgon.NewApp("tests/gorgon.ini")
	logging.SetBackend(logging.NewMemoryBackend(0))

	// the handle that will be tested
	handle := gorgon.GorgonHandler{&app, gorgon.GenerateCertificateHandler}

	// cookie used to authenticate our user
	authCookie, err := GetAuthCookie("user@example.com", app.SessionStore.Codecs...)
	assert.NoError(t, err)

	var (
		data url.Values
		req  *http.Request
		w    *httptest.ResponseRecorder
	)

	// TEST: no auth cookie
	data = url.Values{}
	data.Set("email", "user@example.com")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: email is missing
	data = url.Values{}
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: mismatch emails
	data = url.Values{}
	data.Set("email", "otheruser@example.com")
	data.Add("public_key", "{}")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: cert_duration is missing
	data = url.Values{}
	data.Set("email", "user@example.com")
	data.Add("public_key", "{}")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: public_key is missing
	data = url.Values{}
	data.Set("email", "user@example.com")
	data.Add("cert_duration", "3600")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: malformed public_key
	data = url.Values{}
	data.Set("email", "user@example.com")
	data.Add("cert_duration", "3600")
	data.Add("public_key", "this is a malformed JSON string")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusBadRequest)

	// TEST: check returned certificate
	data = url.Values{}
	data.Set("email", "user@example.com")
	data.Add("cert_duration", "3600")
	data.Add("public_key", "{\"algorithm\":\"DS\",\"y\":\"foobar\"}")
	req, _ = http.NewRequest("POST", "", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	handle.ServeHTTP(w, req)
	assert.Equal(t, w.Code, http.StatusOK)

	token, err := jwt.Parse(w.Body.String(), func(token *jwt.Token) (interface{}, error) {
		return app.PublicKey.PublicKey, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, token.Claims["iss"], "test.example.com")
	assert.Equal(t, token.Claims["public-key"], map[string]interface{}{"algorithm": "DS", "y": "foobar"})
	assert.Equal(t, token.Claims["principal"].(map[string]interface{})["email"], "user@example.com")

	iat := time.Unix(int64(token.Claims["iat"].(float64)/1000), 0)
	assert.True(t, iat.Before(time.Now()))

	exp := time.Unix(int64(token.Claims["exp"].(float64)/1000), 0)
	assert.True(t, exp.After(time.Now()))
}
