BINARY=http-buffer-agent
build:
		go build -o ${BINARY} -ldflags "-X main.Version=${VERSION}"
		go test -v
install:
		go install
release:
		go clean
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-darwin-amd64-${VERSION} -ldflags "-X main.Version=${VERSION}"
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-amd64-${VERSION} -ldflags "-X main.Version=${VERSION}"
		go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${BINARY}-windows-amd64-${VERSION}.exe -ldflags "-X main.Version=${VERSION}"
		go clean
clean:
		go clean
		# rm -f http-buffer-agent-*

.PHONY:  clean build