build: gspot

gspot: $(shell find . -name '*.go')
	go build -ldflags="-X 'git.asdf.cafe/abs3nt/gspot/src/components/cli.Version=$(shell git show -s --date=short --pretty='format:%h (%ad)' HEAD)'" -o dist/ .

run:
	go run main.go

tidy:
	go mod tidy

clean:
	rm -rf bin

uninstall:
	rm -f /usr/bin/gspot
	rm -f /usr/share/zsh/site-functions/_gspot
	rm -f /usr/share/bash-completion/completions/gspot

install:
	cp ./dist/gspot /usr/bin
	cp ./completions/_gspot /usr/share/zsh/site-functions/_gspot
	cp ./completions/gspot /usr/share/bash-completion/completionsgspotg
