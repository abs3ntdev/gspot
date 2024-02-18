build: gospt-ng

gospt-ng: $(shell find . -name '*.go')
	go build -ldflags="-X 'git.asdf.cafe/abs3nt/gospt-ng/src/components/cli.Version=$(shell git rev-parse --short HEAD)'" -o dist/ .

run:
	go run main.go

tidy:
	go mod tidy

clean:
	rm -rf bin

uninstall:
	rm -f /usr/bin/gospt-ng
	rm -f /usr/share/zsh/site-functions/_gospt-ng
	rm -f /usr/share/bash-completion/completions/gospt-ng

install:
	cp ./dist/gospt-ng /usr/bin
	cp ./completions/_gospt-ng /usr/share/zsh/site-functions/_gospt-ng
	cp ./completions/gospt-ng /usr/share/bash-completion/completions/gospt-ng
