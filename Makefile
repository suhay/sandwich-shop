GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOWICH=$(GOINSTALL) $(shell pwd)/gowich/gowich.go
NODEWICH=cd $(shell pwd)/nodewich && yarn install --production && cd $(shell pwd)

install: 
	$(GOINSTALL) $(shell pwd)/shop/sandwich-shop.go

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