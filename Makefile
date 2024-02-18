build: gospt-ng

gospt-ng: $(shell find . -name '*.go')
	go build -ldflags="-X 'git.asdf.cafe/abs3nt/gospt-ng/src/components/cli.Version=$(shell git rev-parse --short HEAD)'" -o dist/ .

run:
	go run main.go

tidy:
	go mod tidy

clean:
	rm -f gospt-ng
	rm -rf completions

uninstall:
	rm -f /usr/bin/gospt-ng
	rm -f /usr/share/zsh/site-functions/_gospt-ng
	rm -f /usr/share/bash-completion/completions/gospt-ng
	rm -f /usr/share/fish/vendor_completions.d/gospt-ng.fish

install:
	cp gospt-ng /usr/bin
	cp ./completions/zsh_autocomplete /usr/share/zsh/site-functions/_gospt-ng
