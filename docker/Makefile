IMAGE_VERSION=0.1.16


build-image:
	docker build -t thirtyx/postgres .

build-dev: build-image push-to-dev

perform-release: build-image push-to-hub

push-to-dev:
	docker tag thirtyx/postgres thirtyx/postgres:dev
	docker push thirtyx/postgres:dev

push-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres:$(IMAGE_VERSION)
		docker push thirtyx/postgres:$(IMAGE_VERSION)
