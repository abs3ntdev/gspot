build: 
	go build -ldflags="-X 'github.com/abs3ntdev/gspot/src/components/cli.Version=$(shell git show -s --date=short --pretty='format:%h (%ad)' HEAD)'" -o dist/ ./cmd/gspot
	go build -o dist/ ./cmd/gspot-daemon

rundaemon: build
	./dist/gspot-daemon

run: build
	./dist/gspot

tidy:
	go mod tidy

clean:
	rm -rf dist

uninstall:
	rm -f /usr/bin/gspot
	rm -f /usr/bin/gspot-daemon
	rm -f /usr/share/zsh/site-functions/_gspot
	rm -f /usr/share/bash-completion/completions/gspot

install:
	cp ./dist/gspot /usr/bin
	cp ./dist/gspot-daemon /usr/bin
	cp ./completions/_gspot /usr/share/zsh/site-functions/_gspot
	cp ./completions/gspot /usr/share/bash-completion/completionsgspotg
