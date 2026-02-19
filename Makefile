run:
	@go run cmd/address.go

query:
	curl http://localhost:8081/search?q=$(q)
fmt:
	@go install mvdan.cc/gofumpt@latest
	@gofumpt -l -w -extra ./.
	@go install github.com/daixiang0/gci@latest
	@go mod tidy