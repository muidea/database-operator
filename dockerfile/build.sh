#!/bin/sh

ldconfig

echo GOARCH=$GOARCH MODULE_NAME=$MODULE_NAME

if [ "$MODULE_NAME" != "" ]; then
  rm -rf /go/bin/$MODULE_NAME
    GOOS=linux GOARCH=$GOARCH  CGO_ENABLED=$CGOENABLE go build -mod=vendor -tags netgo,musl  -o /go/bin/$MODULE_NAME supos.ai/operator/database/cmd/database-operator;
else
    rm -rf database-operator
    GOOS=linux GOARCH=amd64  CGO_ENABLED=$CGOENABLE  go build -mod=vendor -tags netgo,musl supos.ai/operator/database/cmd/database-operator;
fi

