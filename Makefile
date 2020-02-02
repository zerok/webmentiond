fontawesome_version = 5.12.0
fontawesome_archive = fontawesome-pro-$(fontawesome_version)-web.zip
MAIL_FROM ?= no-reply@zerokspot
ALLOWED_TARGET_DOMAINS ?= zerokspot.com

all: bin/webmentiond frontend/fontawesome

prepare:
	test -f .envrc || cp envrc-dist .envrc

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

docker:
	docker build -t zerok/webmentiond:latest .

run-server:
	docker run --rm \
		-e "MAIL_USER=$(MAIL_USER)" \
		-e "MAIL_PORT=$(MAIL_PORT)" \
		-e "MAIL_HOST=$(MAIL_HOST)" \
		-e "MAIL_PASSWORD=$(MAIL_PASSWORD)" \
		-e "MAIL_FROM=no-reply@zerokspot.com" \
		-v $(PWD)/data:/data \
		-p 8080:8080 \
		zerok/webmentiond:latest \
		--addr 0.0.0.0:8080 \
		--auth-jwt-secret testsecret \
		--auth-admin-emails $(AUTH_ADMIN_MAILS) \
		--allowed-target-domains $(ALLOWED_TARGET_DOMAINS)

.PHONY: clean all test frontend-watch docker
