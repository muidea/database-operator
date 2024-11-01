export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on
LDFLAGS := -X 'main.time=$(date -u --rfc-3339=seconds)' -X 'main.git=$(git log --pretty=format:"%h" -1)'
PROJECT=database-operator

all: fmt build

build:
	env CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(PROJECT) ./cmd/$(PROJECT)

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f ./bin/$(PROJECT)
