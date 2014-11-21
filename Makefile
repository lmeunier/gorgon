build:
	go-bindata -o app/bindata.go -pkg app -prefix "./app/data/" ./app/data/...
	go build -o gorgon ./main.go

install_deps:
	go get -u github.com/dgrijalva/jwt-go
	go get -u github.com/gorilla/mux
	go get -u github.com/gorilla/sessions
	go get -u github.com/jteeuwen/go-bindata
