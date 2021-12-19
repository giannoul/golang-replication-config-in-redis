PWD       		= $(shell pwd)
UID           	= $(shell id -u $$USER)
GID           	= $(shell id -g $$USER)
IMAGE_NAME		= golang-redis-replication-config


# Development
## Build images
.PHONY: build-golang-development
build-golang-development: 
	docker build --target debug -t $(IMAGE_NAME) --file Dockerfile .

# copy dlv binary to host
.PHONY: copy-dvl-from-container
copy-dvl-from-container: build-golang-development
	docker run --rm -ti --entrypoint /bin/sleep -d --name dlv-copier $(IMAGE_NAME) 10 && sudo docker cp dlv-copier:/go/bin/dlv /usr/local/bin/

# Docker containers
.PHONY: debug-container-development
debug-container-development: build-golang-development
	docker run --rm -ti -v $(PWD)/src:/src -u $(UID):$(GID) --workdir /src --entrypoint /bin/bash $(IMAGE_NAME)

.PHONY: run-container-development
run-container-development: build-golang-development
	docker run --rm -ti -p 2345:2345 --security-opt seccomp=unconfined --workdir /src --name golang-dev-env $(IMAGE_NAME)

.PHONY: run-application
run-application: build-golang-development
	docker run --rm -ti --workdir /src --name golang-application --entrypoint go $(IMAGE_NAME) run cmd/main.go

.PHONY: start-container-for-development
start-container-for-development: build-golang-development
	docker run --rm -ti -v $(PWD)/src:/src -u $(UID):$(GID) --workdir /src --name golang-dev-env --entrypoint /bin/bash $(IMAGE_NAME)

.PHONY: start-development-environment
start-development-environment: build-golang-development
	docker run --rm -ti -v $(PWD)/src:/src -u $(UID):$(GID) --workdir /src --name golang-dev-env --entrypoint /bin/bash $(IMAGE_NAME)


.PHONY: inspect-running-container
inspect-running-container: 
	docker exec -ti `docker ps | grep golang-dev-env | awk '{print $$1}'` /bin/bash

.PHONY: start-container-for-inspection
start-container-for-inspection: 
	docker run --rm -ti --entrypoint /bin/bash $(IMAGE_NAME)

# Production
## Build images
.PHONY: build-golang-production
build-golang-production: 
	docker build --target production -t $(IMAGE_NAME)-production --file Dockerfile .

## Build golang binary
.PHONY: build-golang-binary
build-golang-binary: build-golang-production
	docker run --rm --name golang-builder -v $(PWD)/dist:/dist --entrypoint /bin/bash $(IMAGE_NAME)-production \
	-c "go build -o /dist/master-finder cmd/main.go"
	sudo chown $(UID):$(GID) ./dist/master-finder


# Helpers
.PHONY: redis-config-clone
redis-config-clone: 
	cp configurations/vanilla-redis.conf configurations/redis.conf 

# Docker compose
.PHONY: docker-compose-start
docker-compose-start: redis-config-clone
	docker-compose up --detach

.PHONY: docker-compose-logs
docker-compose-logs: 
	docker-compose logs -f 

.PHONY: docker-compose-stop
docker-compose-stop: 
	docker-compose stop && docker-compose down