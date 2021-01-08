COMMIT_HASH = $(shell git rev-parse --short HEAD)
COMMIT = $(shell git rev-parse HEAD)
RET = $(shell git describe --contains $(COMMIT_HASH) 1>&2 2> /dev/null; echo $$?)
PWD = $(shell pwd)
USER = $(shell whoami)
buildTime = $(shell date +%Y-%m-%dT%H:%M:%S%z)
PROJ_NAME = iscsi-target-api
RELEASE_TAG = v0.1
DOCKER_REPO = ogre0403

ifeq ($(RET),0)
    TAG = $(shell git describe --contains $(COMMIT_HASH))
else
	TAG = $(USER)-$(COMMIT_HASH)
endif


run:
	./bin/${PROJ_NAME} --logtostderr=true --v=2

test:
	go test  ./...  -v  -cover -count=1 --logtostderr=true


build:
	rm -rf bin/${PROJ_NAME}
	go mod vendor
	go build -mod=vendor  \
	-ldflags '-X "main.buildTime='"${buildTime}"'" -X "main.commitID='"${COMMIT}"'"' \
	-o bin/${PROJ_NAME} cmd/main.go

clean:
	rm -rf bin/*

run-in-docker:
	docker run -d \
	--rm --privileged  \
	-v /etc/tgt/:/etc/tgt/ \
	-v /var/run:/var/run \
	-v /var/lib/iscsi/:/var/lib/iscsi/ \
	-p 8811:8811 \
	${DOCKER_REPO}/${PROJ_NAME}:$(TAG)


build-img:
	docker build -t ${DOCKER_REPO}/${PROJ_NAME}:$(TAG) .


build-in-docker:
	rm -rf bin/*
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor \
	-ldflags '-X "main.buildTime='"${buildTime}"'" -X "main.commitID='"${COMMIT}"'"' \
	-a -installsuffix cgo -o bin/${PROJ_NAME} cmd/main.go


