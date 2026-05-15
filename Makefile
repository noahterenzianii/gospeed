BINARY=gospeed
build:
	go build -o $(BINARY) .
run: build
	./$(BINARY)
clean:
	rm -f $(BINARY)
lint:
	go vet ./...
