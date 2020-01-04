GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GORUN=$(GOCMD) run
GOWICH=$(GOINSTALL) ${GOPATH}/src/github.com/suhay/sandwich-shop/gowich/gowich.go
NODEWICH=cd ${GOPATH}/src/github.com/suhay/sandwich-shop/nodewich && yarn install --production && cd $(shell pwd)

install: 
	$(GOINSTALL) ${GOPATH}/src/github.com/suhay/sandwich-shop/shop/sandwich-shop.go

sandwiches:
	$(GOWICH)
	$(NODEWICH)

nodewich:
	$(NODEWICH)

gowich:
	$(GOWICH)

clean: 
	$(GOCLEAN)
	rm -f ${GOPATH}/bin/sandwich-shop
	rm -f ${GOPATH}/bin/gowich

gqlgen:
	$(GORUN) github.com/99designs/gqlgen -v

generate:
	go generate .

dev:
	CompileDaemon -directory=./shop -color=true -command="./shop/shop"
