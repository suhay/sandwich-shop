GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOWICHINSTALL=$(GOINSTALL) ${GOPATH}/src/github.com/suhay/sandwich-shop/gowich/gowich.go
NODEWICH=cd ${GOPATH}/src/github.com/suhay/sandwich-shop/nodewich && yarn install --production && cd $(shell pwd)

install: 
	$(GOINSTALL) ${GOPATH}/src/github.com/suhay/sandwich-shop/shop/sandwich-shop.go

sandwiches:
	$(GOWICHINSTALL)
	$(NODEWICH)

nodewich:
	$(NODEWICH)

gowich:
	$(GOWICHINSTALL)

clean: 
	$(GOCLEAN)
	rm -f ${GOPATH}/bin/sandwich-shop
	rm -f ${GOPATH}/bin/gowich

gqlgen:
	go run github.com/99designs/gqlgen

generate:
	go generate .

dev:
	MODE=DEV CompileDaemon -directory=./ -color=true -command="./sandwich-shop"
