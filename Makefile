VERSION = $(shell sed -n -e 's/[[:blank:]]Version = "\(.*\)"/\1/p' app/gorgon.go)

build:
	go-bindata -o app/bindata.go -pkg app -prefix "./app/data/" ./app/data/...
	go build -o gorgon ./main.go

dist: build
	mkdir -p dist/gorgon-$(VERSION)/
	cp gorgon gorgon.ini.example dist/gorgon-$(VERSION)/
	tar czf dist/gorgon-$(VERSION).tar.gz -C dist gorgon-$(VERSION) --owner=0 --group=0

install_deps:
	go get -u github.com/dgrijalva/jwt-go
	go get -u github.com/gorilla/mux
	go get -u github.com/gorilla/sessions
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/mxk/go-imap/imap
	go get -u github.com/op/go-logging
	go get -u github.com/vaughan0/go-ini
