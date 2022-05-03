BINARY=panzer-plugin

all: deps linux darwin windows

clean:
	go clean
	if [ -f ./target/linux_amd64/${BINARY} ] ; then rm ./target/${BINARY}-linux_amd64 ; fi
	if [ -f ./target/darwin_amd64/${BINARY} ] ; then rm ./target/${BINARY}-darwin_amd64 ; fi
	if [ -f ./target/windows_amd64/${BINARY} ] ; then rm ./target/${BINARY}-windows_amd64 ; fi

deps:
	go get -v ./...

linux: deps
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/${BINARY}-linux_amd64 .

darwin: deps
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/${BINARY}-darwin_amd64 .

darwin-arm64: deps
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o ./target/${BINARY}-darwin_arm64 .

windows: deps
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/${BINARY}-windows_amd64 .
