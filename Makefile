BIN := cfst

$(BIN): $(wildcard *.go) $(wildcard */*.go)
	CGO_ENABLED=0 go build -ldflags='-s -d -buildid=' .

clean:
	rm -rf $(BIN)
	go clean
