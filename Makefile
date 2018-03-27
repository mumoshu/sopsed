.PHONY: open
open:
	scripts/open-in-code

.PHONY: format
format:
	scripts/format-code

.PHONY: build
build: format
	go build -o bin/sopsed ./examples/kube-aws

.PHONY: it
it: build
	sh -c './bin/sopsed kubectl version'

.PHONY: release
release: build
	scripts/release
