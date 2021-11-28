.PHONY: shop

gqlgen:
	go run github.com/99designs/gqlgen

generate:
	go generate .

dev:
	MODE=DEV CompileDaemon -directory=./ -color=true -command="./sandwich-shop"

all:
	go build
	cd sandwiches/gowich;	go build

shop:
	go build

gowich:
	cd sandwiches/gowich;	go build