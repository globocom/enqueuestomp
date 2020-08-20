CONTAINER_NAME := enqueuestomp_activemq

start-activemq:
	@echo "start container..."
	docker container run -d --rm --name $(CONTAINER_NAME) -p 61613:61613 -p 61616:61616 -p 8161:8161  rmohr/activemq:5.15.9-alpine
	@sleep 10

stop-activemq:
	@echo "stop container..."
	docker container rm -f $(CONTAINER_NAME)

test:
	@if [ ! $(shell docker container ls -f name=$(CONTAINER_NAME) -q) ]; then \
		$(MAKE) start-activemq; \
	fi; \
	go test -v -count=1 -cover -race ./...

test-coverage:
	@if [ ! $(shell docker container ls -f name=$(CONTAINER_NAME) -q) ]; then \
		$(MAKE) start-activemq; \
	fi; \
	go test -v -count=1 -race -cover -covermode atomic -coverprofile coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

ci:
	@sleep 10
	go test -count=1 -cover -race ./...
