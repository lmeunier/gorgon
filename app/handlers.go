package app

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// gorgonHandler implements the Handler interface to add the ability to access
// our GorgonApp from handlers.
type gorgonHandler struct {
	app    *GorgonApp
	handle func(*GorgonApp, http.ResponseWriter, *http.Request) error
}

// ServeHTTP add the ability to access our GorgonApp from handlers.
func (gh gorgonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := gh.handle(gh.app, w, r)

	if err != nil {
		gh.app.Logger.Error(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SupportDocumentHandler returns the SupportDocument in a JSON encoded response.
func SupportDocumentHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	support_document := app.supportDocument
	b, err := json.Marshal(support_document)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}

// AuthenticationHandler is responsible of presenting the auth form and
// authenticating the user using the app Authenticator and the
// username/password provided by the user.
// If the user is successfully authenticated, the "persona-auth" cookie is
// updated with the username used in the authentication process.
func AuthenticationHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	ctx := make(map[string]interface{})
	ctx["Email"] = ""

	session, _ := app.sessionStore.Get(r, "persona-auth")

	if r.Method == "POST" {
		// the user submitted the HTMl form
		username := r.FormValue("email")
		password := r.FormValue("password")

		ctx["Email"] = username

		// try to authenticate the user
		err := app.Authenticator.Authenticate(username, password)
		if err == nil {
			// the authentication process is ok
			// add the username in the session
			session.Values["authenticated_as"] = username
		} else {
			// the authentication process failed
			// remove the username from the session
			delete(session.Values, "authenticated_as")
			// notify the user
			ctx["ValidationError"] = "Authentication failure"

			app.Logger.Warning("Authentication failed for '" + username + "': " + err.Error())
		}
	}
	session.Save(r, w)

	if emails, ok := r.URL.Query()["email"]; ok {
		ctx["Email"] = emails[0]
	}

	// render the template
	ctx["Session"] = session
	return app.templates.ExecuteTemplate(w, "authentication.html", ctx)
}

// ProvisioningHandler returns the content of hidden iframe. The content
// depends if the user have an active session or not.
func ProvisioningHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	ctx := make(map[string]interface{})
	session, _ := app.sessionStore.Get(r, "persona-auth")
	ctx["Session"] = session

	return app.templates.ExecuteTemplate(w, "provisioning.html", ctx)
}

// GenerateCertificateHandler is called via an AJAX request from the
// provisioning page when the user is authenticated. This handler returns a
// generated certificate from informations provided in the query string.
func GenerateCertificateHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := app.sessionStore.Get(r, "persona-auth")
	email := ""
	if vals, ok := r.URL.Query()["email"]; ok {
		email = vals[0]
	}
	if email != session.Values["authenticated_as"] {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cert_duration := 0
	if vals, ok := r.URL.Query()["cert_duration"]; ok {
		cert_duration, _ = strconv.Atoi(vals[0])
	}
	var pubkey map[string]string
	if vals, ok := r.URL.Query()["public_key"]; ok {
		json.Unmarshal([]byte(vals[0]), &pubkey)
	}

	private_key := app.privateKey
	public_key := app.publicKey

	// with all theses informations, we can now generate a certificate
	certificate, err := CreateCertificate(private_key, public_key, email, cert_duration, pubkey, app.domain)
	if err != nil {
		return
	}

	// send the certificate to the browser
	w.Write(certificate)
	return
}

// CheckAuthenticateHandler checks if the user has an active session (the user
// is authenticated). If the user is not authenticated returns an HTTP code 403
// (Forbidden), else returns an HTTP code 200 (OK).
func CheckAuthenticatedHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := app.sessionStore.Get(r, "persona-auth")
	_, ok := session.Values["authenticated_as"]

	if !ok {
		w.WriteHeader(http.StatusForbidden)
	}

	return
}
