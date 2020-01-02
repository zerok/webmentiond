fontawesome_version = 5.12.0
fontawesome_archive = fontawesome-pro-$(fontawesome_version)-web.zip

all: bin/webmentiond frontend/fontawesome

bin:
	mkdir -p bin

clean:
	rm -rf bin

bin/webmentiond: $(shell find . -name '*.go') go.mod bin
	cd cmd/webmentiond && go build -o ../../$@

test:
	go test ./... -v

frontend/fontawesome: frontend/$(fontawesome_archive)
	cd frontend && unzip $(fontawesome_archive) && mv "fontawesome-pro-$(fontawesome_version)-web" fontawesome

frontend/$(fontawesome_archive):
	$(error "Please download $(fontawesome_archive) from https://fontawesome.com/download and place it inside ./frontend")

frontend-watch:
	cd frontend && yarn && yarn run webpack --watch

.PHONY: clean all test frontend-watch
