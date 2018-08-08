test:
	go test --race ./swagger
	go test --race ./parser

cover:
	rm -f *.coverprofile
	go test -coverprofile=swagger.coverprofile ./swagger
	go test -coverprofile=parser.coverprofile ./parser
	gover
	go tool cover -html=gover.coverprofile
	rm -f *.coverprofile

build:
	go build -o swaggo main.go

.PHONY: test cover build
