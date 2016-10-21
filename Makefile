IMAGE_VERSION=0.0.3-dev

build-and-push: build-postgres push-postgres-to-hub

build-and-push-benchmark: build-benchmark push-benchmark-to-hub

build-postgres:
	docker build -t thirtyx/postgres docker-image

build-benchmark:
	docker build -t thirtyx/postgres-benchmark benchmark


push-postgres-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres:$(IMAGE_VERSION)
		docker push thirtyx/postgres:$(IMAGE_VERSION)

push-benchmark-to-hub:
		docker tag thirtyx/postgres-benchmark thirtyx/postgres-benchmark:$(IMAGE_VERSION)
		docker push thirtyx/postgres-benchmark:$(IMAGE_VERSION)
