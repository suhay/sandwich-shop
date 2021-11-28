gqlgen:
	go run github.com/99designs/gqlgen

generate:
	go generate .

dev:
	MODE=DEV CompileDaemon -directory=./ -color=true -command="./sandwich-shop"

build:
	go build
	cd sandwiches/gowich;	go build
