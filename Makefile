IMAGE_VERSION=0.0.3-dev

build-and-push: build-postgres push-postgres-to-hub

build-and-push-benchmark: build-benchmark push-benchmark-to-hub

build-postgres:
	docker build -t thirtyx/postgres docker-image

build-benchmark:
	docker build -t thirtyx/postgres-benchmark benchmark

build-dev: compile build-image push-to-dev

perform-release: compile build-image push-to-hub

push-postgres-to-dev:
	docker tag thirtyx/postgres thirtyx/postgres:dev
	docker push thirtyx/postgres:dev

push-postgres-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres:$(IMAGE_VERSION)
		docker push thirtyx/postgres:$(IMAGE_VERSION)

push-benchmark-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres-benchmark:$(IMAGE_VERSION)
		docker push thirtyx/postgres-benchmark:$(IMAGE_VERSION)
