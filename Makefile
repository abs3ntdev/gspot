generate:
	buf generate

build:
	go build -ldflags="-X 'github.com/abs3ntdev/gspot/src/cli.Version=$(shell git show -s --date=short --pretty='format:%h (%ad)' HEAD)'" -o dist/ ./cmd/gspot

run: build
	./dist/gspot

daemon: build
	./dist/gspot daemon run

tidy:
	go mod tidy

clean:
	rm -rf dist

uninstall:
	rm -f /usr/bin/gspot
	rm -f /usr/share/zsh/site-functions/_gspot
	rm -f /usr/share/bash-completion/completions/gspot

install:
	cp ./dist/gspot /usr/bin
	cp ./completions/_gspot /usr/share/zsh/site-functions/_gspot
	cp ./completions/gspot /usr/share/bash-completion/completions/gspot
