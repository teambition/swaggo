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

.PHONY: test cover
