package app

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"strings"
)

// sessions.NewCookieStore([]byte("blabla"))

type SupportDocument struct {
	Authentication string     `json:"authentication""`
	Provisioning   string     `json:"provisioning"`
	PublicKey      *PublicKey `json:"public-key"`
}

type GorgonApp struct {
	Router          *mux.Router
	sessionStore    *sessions.CookieStore
	supportDocument SupportDocument
	publicKey       *PublicKey
	privateKey      *PrivateKey
	templates       *template.Template
}

func NewApp() GorgonApp {
	public_key, err := LoadPublicKey("TODO-public-key.pem")
	if err != nil {
		log.Panic(err)
	}

	private_key, err := LoadPrivateKey("TODO-private-key.pem")
	if err != nil {
		log.Panic(err)
	}

	support_document := SupportDocument{
		"/browserid/authentication.html",
		"/browserid/provisioning.html",
		public_key,
	}

	templates := template.New("")
	for _, assetName := range AssetNames() {
		if strings.HasSuffix(assetName, ".html") {
			data, err := Asset(assetName)
			if err != nil {
				log.Panic(err)
			}
			templates.New(assetName).Parse(string(data))
		}
	}

	app := GorgonApp{
		mux.NewRouter(),
		sessions.NewCookieStore([]byte("TODO")),
		support_document,
		public_key,
		private_key,
		templates,
	}

	app.Router.Handle("/.well-known/browserid", gorgonHandler{app, SupportDocumentHandler})
	app.Router.Handle("/browserid/authentication.html", gorgonHandler{app, AuthenticationHandler})
	app.Router.Handle("/browserid/provisioning.html", gorgonHandler{app, ProvisioningHandler})
	app.Router.Handle("/browserid/generate_certificate.html", gorgonHandler{app, GenerateCertificateHandler})
	app.Router.Handle("/browserid/is_authenticated", gorgonHandler{app, CheckAuthenticatedHandler})

	return app
}
