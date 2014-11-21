package app

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/vaughan0/go-ini"
	"html/template"
	"log"
	"strings"
)

type SupportDocument struct {
	Authentication string     `json:"authentication""`
	Provisioning   string     `json:"provisioning"`
	PublicKey      *PublicKey `json:"public-key"`
}

type GorgonApp struct {
	Config          ini.File
	Router          *mux.Router
	sessionStore    *sessions.CookieStore
	supportDocument SupportDocument
	publicKey       *PublicKey
	privateKey      *PrivateKey
	templates       *template.Template
	domain          string
}

func NewApp(config_file string) GorgonApp {
	config, err := ini.LoadFile(config_file)
	if err != nil {
		log.Panic(err)
	}

	public_key_filename, _ := config.Get("global", "public_key")
	public_key, err := LoadPublicKey(public_key_filename)
	if err != nil {
		log.Panic(err)
	}

	private_key_filename, _ := config.Get("global", "private_key")
	private_key, err := LoadPrivateKey(private_key_filename)
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

	domain, _ := config.Get("global", "idp_domain")

	app := GorgonApp{
		config,
		mux.NewRouter(),
		sessions.NewCookieStore([]byte("TODO")),
		support_document,
		public_key,
		private_key,
		templates,
		domain,
	}

	app.Router.Handle("/.well-known/browserid", gorgonHandler{app, SupportDocumentHandler})
	app.Router.Handle("/browserid/authentication.html", gorgonHandler{app, AuthenticationHandler})
	app.Router.Handle("/browserid/provisioning.html", gorgonHandler{app, ProvisioningHandler})
	app.Router.Handle("/browserid/generate_certificate.html", gorgonHandler{app, GenerateCertificateHandler})
	app.Router.Handle("/browserid/is_authenticated", gorgonHandler{app, CheckAuthenticatedHandler})

	return app
}
