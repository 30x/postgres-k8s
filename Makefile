IMAGE_VERSION=0.0.7-dev


build-cli:
	cd cli && \
	GOOS=darwin GOARCH=amd64 go build -a ${EXTRA_BUILDFLAGS} -o ${GOPATH}/bin/pgctl -ldflags "${EXTRA_LDFLAGS} -X main.BuildVersion=${VERSION}" 

build-and-push: build-postgres push-postgres-to-hub

build-and-push-benchmark: build-benchmark push-benchmark-to-hub

build-postgres:
	docker build -t thirtyx/transicator-postgres-k8s docker-image

build-benchmark:
	docker build -t thirtyx/transicator-postgres-k8s-benchmark benchmark


push-postgres-to-hub:
		docker tag thirtyx/transicator-postgres-k8s thirtyx/transicator-postgres-k8s:$(IMAGE_VERSION)
		docker push thirtyx/transicator-postgres-k8s:$(IMAGE_VERSION)

push-benchmark-to-hub:
		docker tag thirtyx/transicator-postgres-k8s-benchmark thirtyx/transicator-postgres-k8s-benchmark:$(IMAGE_VERSION)
		docker push thirtyx/transicator-postgres-k8s-benchmark:$(IMAGE_VERSION)
