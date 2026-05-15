BINARY=gospeed
build:
	go build -o $(BINARY) ./cmd/gospeed
run: build
	./$(BINARY)
clean:
	rm -f $(BINARY)
lint:
	go vet ./...
