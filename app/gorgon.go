package app

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/op/go-logging"
	"github.com/vaughan0/go-ini"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	Version = "0.1.0"
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
	Config        ini.File              // configuration read from a configuration file
	Router        *mux.Router           // routes to URL
	SessionStore  *sessions.CookieStore // users sessions (client side cookie)
	PublicKey     *PublicKey            // public key for the domain
	PrivateKey    *PrivateKey           // private key for the domain
	Templates     *template.Template    // list of all templates used by the application
	Domain        string                // domain name used for this IdP
	Authenticator Authenticator         // method to authenticate users
	ListenAddress string                // network address on which the app will listens
	Logger        *logging.Logger       // Logger for this app
}

// NewApp returns a GorgonApp fully configured and initialized. Panic if the
// app can't be initialized.
func NewApp(config_file string) GorgonApp {
	// initialize the logger
	logger := logging.MustGetLogger("gorgon")
	var format = logging.MustStringFormatter(
		"[%{time:" + time.RFC3339 + "} %{level}] %{message}",
	)
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	bf := logging.NewBackendFormatter(backend1, format)
	logging.SetBackend(bf)

	// load the configuration file
	config, err := ini.LoadFile(config_file)
	if err != nil {
		logger.Fatal("Unable to load configuration file '" + config_file + "': " + err.Error())
	}

	// load the public key
	public_key_filename, _ := config.Get("global", "public_key")
	public_key, err := LoadPublicKey(public_key_filename)
	if err != nil {
		logger.Fatal("Unable to load public key '" + public_key_filename + "': " + err.Error())
	}

	// load the private key
	private_key_filename, _ := config.Get("global", "private_key")
	private_key, err := LoadPrivateKey(private_key_filename)
	if err != nil {
		logger.Fatal("Unable to load private key '" + private_key_filename + "': " + err.Error())
	}

	// load all "*.html" templates from the data directory
	templates := template.New("")
	for _, assetName := range AssetNames() {
		if strings.HasSuffix(assetName, ".html") {
			data, err := Asset(assetName)
			if err != nil {
				logger.Fatal("Unable to load template '" + assetName + "': " + err.Error())
			}
			templates.New(assetName).Parse(string(data))
		}
	}

	// the domain used for this IdP (should be the domain part of the email address)
	domain, _ := config.Get("global", "idp_domain")

	// the listen network address
	listenAddress, _ := config.Get("global", "listen")

	// the session secret key
	session_secret_key, _ := config.Get("global", "session_secret_key")
	if session_secret_key == "" {
		logger.Fatal("The 'session_secret_key' is empty.")
	}
	if len(session_secret_key) != 64 && len(session_secret_key) != 32 {
		logger.Fatalf("The 'session_secret_key' must have a length of 32 or 64 bytes (currently: %d).", len(session_secret_key))
	}

	// create the Gorgon application
	app := GorgonApp{
		config,
		mux.NewRouter(),
		sessions.NewCookieStore([]byte(session_secret_key)),
		public_key,
		private_key,
		templates,
		domain,
		nil,
		listenAddress,
		logger,
	}

	// create the authentication method
	authenticator_name, _ := config.Get("global", "auth")
	authenticator, err := NewAuthenticator(app, authenticator_name)
	if err != nil {
		logger.Panic("Unable to create auth backend '" + authenticator_name + "': " + err.Error())
	}
	app.Authenticator = authenticator

	// define routes
	app.Router.Handle(
		"/.well-known/browserid",
		GorgonHandler{&app, SupportDocumentHandler}).
		Methods("GET").
		Name("support_document")

	app.Router.Handle(
		"/.well-known/browserid/_gorgon/authentication",
		GorgonHandler{&app, AuthenticationHandler}).
		Methods("GET", "POST").
		Name("authentication")

	app.Router.Handle(
		"/.well-known/browserid/_gorgon/provisioning",
		GorgonHandler{&app, ProvisioningHandler}).
		Methods("GET").
		Name("provisioning")

	app.Router.Handle(
		"/.well-known/browserid/_gorgon/generate_certificate",
		GorgonHandler{&app, GenerateCertificateHandler}).
		Methods("POST").
		Name("generate_certificate")

	app.Router.Handle(
		"/.well-known/browserid/_gorgon/is_authenticated",
		GorgonHandler{&app, CheckAuthenticatedHandler}).
		Methods("GET").
		Name("check_authenticate")

	return app
}

// ListenAndServe listens on the TCP network address provided by the app
// configuration and then serve requests on incoming connections.
func (app GorgonApp) ListenAndServe() error {
	return http.ListenAndServe(app.ListenAddress, app.Router)
}

// GetSupportDocument returns a SupportDocument struct for the GorgonApp.
func (app *GorgonApp) GetSupportDocument() SupportDocument {
	// create the support document
	authentication_url, _ := app.Router.Get("authentication").URL()
	provisioning_url, _ := app.Router.Get("provisioning").URL()
	return SupportDocument{
		authentication_url.String(),
		provisioning_url.String(),
		app.PublicKey,
	}
}
