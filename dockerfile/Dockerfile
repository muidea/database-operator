FROM registry.supos.ai/library/golang-builder:1.19.9_amd64  AS builder

# define source dir
WORKDIR /src

# add source dir current dir must be in project root
ADD . .

# set env for go build
ENV GOOS linux
ENV GOARCH amd64
ENV CGOENABLE 0

# get make args for make
ARG MODULE_NAME
RUN chmod +x ./dockerfile/build.sh
RUN ./dockerfile/build.sh
RUN chmod +x ./dockerfile/bootstrap.sh
RUN chmod +x /go/bin/$MODULE_NAME

FROM registry.supos.ai/library/golang-runner:3.18.0_amd64

LABEL Author="rangh"
LABEL Application=database-operator

WORKDIR /var/app

ARG MODULE_NAME
ENV MODULE_NAME=$MODULE_NAME

COPY --from=builder /go/bin/$MODULE_NAME .
COPY --from=builder /src/dockerfile/bootstrap.sh .

ENTRYPOINT [ "/var/app/bootstrap.sh" ]

EXPOSE 8080
