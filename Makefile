
DEPS := $(wildcard *.go)

cmd/godoctest/godoctest: cmd/godoctest/main.go $(DEPS)
	go build -o $@ $<

.PHONY: test
test:
	./cmd/godoctest/godoctest
