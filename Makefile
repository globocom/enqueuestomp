CONTAINER_MQ := $(shell docker container ls -f name=enqueuestomp-activemq -q)

start-activemq:
	@echo "start container..."
	docker container run -d --rm --name enqueuestomp-activemq -p 61613:61613 -p 61616:61616 -p 8161:8161  rmohr/activemq:5.15.9-alpine

stop-activemq:
	@if [ "$(CONTAINER_MQ)" ]; then \
		echo "stop container..."; \
		docker container rm -f $(CONTAINER_MQ);\
	fi; \

test:
	@if [ ! "$(CONTAINER_MQ)" ]; then \
		$(MAKE) start-activemq; \
		sleep 5; \
	fi; \
	go test -count=1 -cover ./...
