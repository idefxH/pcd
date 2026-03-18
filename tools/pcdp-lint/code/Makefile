.PHONY: build clean test install

BINARY_NAME=spec-lint
TEMPLATE_DIR=/usr/share/post-coding/templates/

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -X main.TemplateDir=$(TEMPLATE_DIR)' -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

install: build
	mkdir -p $(DESTDIR)/usr/bin
	install -m 755 $(BINARY_NAME) $(DESTDIR)/usr/bin/$(BINARY_NAME)
	mkdir -p $(DESTDIR)/usr/share/post-coding/templates

rpm: build
	rpmbuild -ba spec-lint.spec

deb: 
	debuild -us -uc