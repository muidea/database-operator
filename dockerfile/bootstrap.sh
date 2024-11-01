#!/bin/sh

EXTRA_ARGS=$EXTRA_ARGS
if [ "$LISTENPORT" ]; then
    EXTRA_ARGS=$EXTRA_ARGS' -ListenPort='$LISTENPORT
fi

if [ "$ENDPOINTNAME" ]; then
    EXTRA_ARGS=$EXTRA_ARGS' -EndpointName='$ENDPOINTNAME
fi

/var/app/${MODULE_NAME} $EXTRA_ARGS
