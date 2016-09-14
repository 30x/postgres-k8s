IMAGE_VERSION=0.1.16


compile:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/posgres-agent .

build-image:
	docker build -t thirtyx/postgres .

build-dev: compile build-image push-to-dev

perform-release: compile build-image push-to-hub

push-to-dev:
	docker tag thirtyx/postgres thirtyx/postgres:dev
	docker push thirtyx/postgres:dev

push-to-hub:
		docker tag thirtyx/postgres thirtyx/postgres:$(IMAGE_VERSION)
		docker push thirtyx/postgres:$(IMAGE_VERSION)
