package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// GorgonHandler implements the Handler interface to add the ability to access
// our GorgonApp from handlers.
type GorgonHandler struct {
	App    *GorgonApp
	Handle func(*GorgonApp, http.ResponseWriter, *http.Request) error
}

// ServeHTTP add the ability to access our GorgonApp from handlers.
func (gh GorgonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := gh.Handle(gh.App, w, r)

	if err != nil {
		gh.App.Logger.Error(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// SupportDocumentHandler returns the SupportDocument in a JSON encoded response.
func SupportDocumentHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	support_document := app.GetSupportDocument()
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
	ctx["App"] = app
	ctx["Email"] = ""

	session, _ := app.SessionStore.Get(r, "persona-auth")

	if r.Method == "POST" {
		// the user submitted the HTML form
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
			ctx["ValidationError"] = true

			app.Logger.Warning("Authentication failed for '" + username + "': " + err.Error())
		}
	}
	session.Save(r, w)

	if emails, ok := r.URL.Query()["email"]; ok {
		ctx["Email"] = emails[0]
	}

	// render the template
	ctx["Session"] = session
	return app.Templates.ExecuteTemplate(w, "authentication.html", ctx)
}

// ProvisioningHandler returns the content of hidden iframe. The content
// depends if the user have an active session or not.
func ProvisioningHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	ctx := make(map[string]interface{})
	session, _ := app.SessionStore.Get(r, "persona-auth")
	generate_certificate_url, _ := app.Router.Get("generate_certificate").URL()
	ctx["Session"] = session
	ctx["generate_certificate_url"] = generate_certificate_url

	return app.Templates.ExecuteTemplate(w, "provisioning.html", ctx)
}

// GenerateCertificateHandler is called via an AJAX request from the
// provisioning page when the user is authenticated. This handler returns a
// generated certificate from informations provided in the query string.
func GenerateCertificateHandler(app *GorgonApp, w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := app.SessionStore.Get(r, "persona-auth")

	// parse data from the POST body
	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch `email` from POST data
	email := ""
	if vals, ok := r.PostForm["email"]; ok {
		email = vals[0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the email received form the AJAX request is the same as the
	// email in the current session. We need to be sure the user is
	// authenticated with the same email address before creating a certificate.
	// This is very important to avoid forged requests to obtain a valid
	// certificate for any email address.
	if email != session.Values["authenticated_as"] {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch `cert_duration` from POST data
	cert_duration := time.Duration(0)
	if vals, ok := r.PostForm["cert_duration"]; ok {
		num_seconds, err := strconv.Atoi(vals[0])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return err
		}
		cert_duration = time.Duration(num_seconds) * time.Second
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch `public_key` from POST data
	var pubkey map[string]string
	if vals, ok := r.PostForm["public_key"]; ok {
		err := json.Unmarshal([]byte(vals[0]), &pubkey)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return err
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// with all theses informations, we can now generate a certificate
	certificate, err := CreateCertificate(app.PrivateKey, app.PublicKey, email, cert_duration, pubkey, app.Domain)
	if err != nil {
		if _, ok := err.(*CertDurationError); ok {
			app.Logger.Warning(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return nil
		}
		w.WriteHeader(http.StatusBadRequest)
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
	session, _ := app.SessionStore.Get(r, "persona-auth")
	_, ok := session.Values["authenticated_as"]

	if !ok {
		w.WriteHeader(http.StatusForbidden)
	}

	return
}
