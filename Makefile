IMAGE_VERSION=0.0.1

build-and-push: build push-to-hub

build:
	docker build -t thirtyx/postgres .

build-dev: compile build-image push-to-dev

perform-release: compile build-image push-to-hub

push-to-dev:
	docker tag thirtyx/postgres thirtyx/postgres:dev
	docker push thirtyx/postgres:dev

push-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres:$(IMAGE_VERSION)
		docker push thirtyx/postgres:$(IMAGE_VERSION)
