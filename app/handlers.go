package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type gorgonHandler struct {
	app    GorgonApp
	handle func(GorgonApp, http.ResponseWriter, *http.Request)
}

func (gh gorgonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gh.handle(gh.app, w, r)
}

func SupportDocumentHandler(app GorgonApp, w http.ResponseWriter, r *http.Request) {
	support_document := app.supportDocument
	b, err := json.Marshal(support_document)
	if err != nil {
		log.Panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func AuthenticationHandler(app GorgonApp, w http.ResponseWriter, r *http.Request) {
	ctx := make(map[string]interface{})
	ctx["Email"] = ""

	session, _ := app.sessionStore.Get(r, "persona-auth")

	if r.Method == "POST" {
		ctx["Email"] = r.FormValue("email")
		password := r.FormValue("password")

		if password == "verysecret" {
			session.Values["authenticated_as"] = ctx["Email"]
		} else {
			delete(session.Values, "authenticated_as")
			ctx["ValidationError"] = "Authentication failure"
		}
	}
	session.Save(r, w)

	if emails, ok := r.URL.Query()["email"]; ok {
		ctx["Email"] = emails[0]
	}

	ctx["Session"] = session
	err := app.templates.ExecuteTemplate(w, "authentication.html", ctx)
	if err != nil {
		log.Panic(err)
	}
}

func ProvisioningHandler(app GorgonApp, w http.ResponseWriter, r *http.Request) {
	ctx := make(map[string]interface{})
	session, _ := app.sessionStore.Get(r, "persona-auth")
	ctx["Session"] = session

	err := app.templates.ExecuteTemplate(w, "provisioning.html", ctx)
	if err != nil {
		log.Panic(err)
	}
}

func GenerateCertificateHandler(app GorgonApp, w http.ResponseWriter, r *http.Request) {
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

	// TODO vérifier la public_key
	certificate, err := CreateCertificate(private_key, public_key, email, cert_duration, pubkey, app.domain)
	if err != nil {
		log.Panic(err)
	}

	w.Write(certificate)
}

func CheckAuthenticatedHandler(app GorgonApp, w http.ResponseWriter, r *http.Request) {
	session, _ := app.sessionStore.Get(r, "persona-auth")
	_, ok := session.Values["authenticated_as"]

	if !ok {
		log.Println("Check: *NOT* authenticated")
		w.WriteHeader(http.StatusForbidden)
	}

	log.Println("Check: authenticated")
}
