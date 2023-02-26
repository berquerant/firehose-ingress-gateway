.PHONY: generate
generate: re-go-generate

.PHONY: re-go-generate
re-go-generate: clean-go-generated go-generate

.PHONY: clean-go-generated
clean-go-generated:
	find . -name "*_generated.go" -type f -delete

.PHONY: go-generate
go-generate:
	go generate ./...

.PHONY: test
test:
	go test ./...
