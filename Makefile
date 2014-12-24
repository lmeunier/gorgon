VERSION = $(shell sed -n -e 's/[[:blank:]]Version = "\(.*\)"/\1/p' app/gorgon.go)
OSARCHS = linux/amd64 linux/arm darwin/amd64 freebsd/amd64 openbsd/amd64 freebsd/arm netbsd/amd64 netbsd/arm
PROGRAMS = $(foreach OSARCH,$(OSARCHS),gorgon_$(subst /,_,$(OSARCH)))

.PHONY: build dist

build:
	go-bindata -o app/bindata.go -pkg app -prefix "./app/data/" ./app/data/...
	gox -output "build/gorgon_{{.OS}}_{{.Arch}}" -osarch "$(OSARCHS)"

dist: build $(PROGRAMS)

$(PROGRAMS):
	@mkdir -p dist
	tar czf "dist/$(@)-$(VERSION).tar.gz" gorgon.ini.example -C build "$(@)" --owner=0 --group=0

clean:
	rm -rf app/bindata.go dist/ build/

install_deps:
	go get -u github.com/dgrijalva/jwt-go
	go get -u github.com/gorilla/mux
	go get -u github.com/gorilla/sessions
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/mitchellh/gox
	go get -u github.com/mxk/go-imap/imap
	go get -u github.com/op/go-logging
	go get -u github.com/vaughan0/go-ini
