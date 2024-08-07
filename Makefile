build:
	@echo 'Building the project...'
	go build -o cashFlowManager/cmd/service/main

run: build
	@echo 'Running the project...'
	./cashFlowManager/cmd/service/main

clean:
	@echo 'Cleaning...'
	go clean
	rm -f cashFlowManager/cmd/service/main
lint:
	@echo 'Linting the project...'
	gofumpt -w .
	go mod tidy
	golangci-lint run --fix -c .golangci.yaml
