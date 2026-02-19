run:
	@go run cmd/address.go

query:
	curl http://localhost:8081/search?q=$(q) --silent | jq -r
fmt:
	@go install mvdan.cc/gofumpt@latest
	@gofumpt -l -w -extra ./.
	@go install github.com/daixiang0/gci@latest
	@go mod tidy