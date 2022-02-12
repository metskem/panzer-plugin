BINARY=panzer-plugin

all: deps tests linux darwin windows

clean:
	go clean
	if [ -f ./target/linux_amd64/${BINARY} ] ; then rm ./target/linux_amd64/${BINARY} ; fi
	if [ -f ./target/darwin_amd64/${BINARY} ] ; then rm ./target/darwin_amd64/${BINARY} ; fi
	if [ -f ./target/windows_amd64/${BINARY} ] ; then rm ./target/windows_amd64/${BINARY} ; fi

deps:
	go get -v ./...

tests: deps
	go test ./...

linux: tests
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/linux_amd64/${BINARY} .

darwin: tests
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/darwin_amd64/${BINARY} .

windows: tests
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./target/windows_amd64/${BINARY} .
