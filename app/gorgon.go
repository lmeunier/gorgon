package app

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/vaughan0/go-ini"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// SupportDocument represents the document where domains advertise their
// ability to act as Persona Identity Providers located at:
// "/.well-known/browserid".
type SupportDocument struct {
	Authentication string     `json:"authentication""`
	Provisioning   string     `json:"provisioning"`
	PublicKey      *PublicKey `json:"public-key"`
}

// GorgonApp represents an application used to act as a Persona IdP.
type GorgonApp struct {
	Config          ini.File              // configuration read from a configuration file
	Router          *mux.Router           // routes to URL
	sessionStore    *sessions.CookieStore // users sessions (client side cookie)
	supportDocument SupportDocument       // Persona support document
	publicKey       *PublicKey            // public key for the domain
	privateKey      *PrivateKey           // private key for the domain
	templates       *template.Template    // list of all templates used by the application
	domain          string                // domain name used for this IdP
	Authenticator   Authenticator         // method to authenticate users
	ListenAddress   string                // network address on which the app will listens
}

// NewApp returns a GorgonApp fully configured and initialized. Panic if the
// app can't be initialized.
func NewApp(config_file string) GorgonApp {
	// load the configuration file
	config, err := ini.LoadFile(config_file)
	if err != nil {
		log.Panic(err)
	}

	// load the public key
	public_key_filename, _ := config.Get("global", "public_key")
	public_key, err := LoadPublicKey(public_key_filename)
	if err != nil {
		log.Panic(err)
	}

	// load the private key
	private_key_filename, _ := config.Get("global", "private_key")
	private_key, err := LoadPrivateKey(private_key_filename)
	if err != nil {
		log.Panic(err)
	}

	// create the support document
	support_document := SupportDocument{
		"/browserid/authentication.html",
		"/browserid/provisioning.html",
		public_key,
	}

	// load all "*.html" templates from the data directory
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

	// the domain used for this IdP (should be the domain part of the email address)
	domain, _ := config.Get("global", "idp_domain")

	// the listen network address
	listenAddress, _ := config.Get("global", "listen")

	// create the Gorgon application
	app := GorgonApp{
		config,
		mux.NewRouter(),
		sessions.NewCookieStore([]byte("TODO")),
		support_document,
		public_key,
		private_key,
		templates,
		domain,
		nil,
		listenAddress,
	}

	// create the authentication method
	authenticator_name, _ := config.Get("global", "auth")
	authenticator, err := NewAuthenticator(app, authenticator_name)
	if err != nil {
		log.Panic(err)
	}
	app.Authenticator = authenticator

	// define routes
	app.Router.Handle("/.well-known/browserid", gorgonHandler{app, SupportDocumentHandler})
	app.Router.Handle("/browserid/authentication.html", gorgonHandler{app, AuthenticationHandler})
	app.Router.Handle("/browserid/provisioning.html", gorgonHandler{app, ProvisioningHandler})
	app.Router.Handle("/browserid/generate_certificate.html", gorgonHandler{app, GenerateCertificateHandler})
	app.Router.Handle("/browserid/is_authenticated", gorgonHandler{app, CheckAuthenticatedHandler})

	return app
}

// ListenAndServe listens on the TCP network address provided by the app
// configuration and then serve requests on incoming connections.
func (app GorgonApp) ListenAndServe() error {
	return http.ListenAndServe(app.ListenAddress, app.Router)
}
