.PHONY: check

check: test tools check-fmt check-vet check-misspell  check-lint
.PHONY:  check-fmt
check-fmt:
	@echo Check code is formatted
	@bash -c 'if [ -n "$(gofmt -s -l .)" ]; then echo "Go code is not formatted:"; gofmt -s -d -e .; exit 1;fi'

.PHONY: check-vet
check-vet:
	@echo Check go vet
	@go vet ./...

.PHONY: check-misspell
check-misspell:
	@echo Check misspell
	@misspell ./...

.PHONY: check-lint
check-lint:
	@echo Check lint
	@golint --set_exit_status ./...

.PHONY: tools
tools:
	@echo get tools
	@go get -u github.com/client9/misspell/cmd/misspell
	@go get -u github.com/golang/lint/golint

test:
	# get project dependencies
	go get gopkg.in/go-playground/validator.v9 
	# run tests
	go test -cover -race ./...
