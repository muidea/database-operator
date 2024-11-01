package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/muidea/magicCommon/application"
	"github.com/muidea/magicCommon/foundation/log"

	"supos.ai/operator/database/internal/config"
	"supos.ai/operator/database/internal/core"
)

var listenPort = "8080"
var endpointName = "database-operator"

func initPprofMonitor(listenPort string) {
	addr := ":1" + listenPort

	go func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Criticalf("funcRetErr=http.ListenAndServe||err=%s", err.Error())
		}
	}()
}

func main() {
	flag.StringVar(&listenPort, "ListenPort", listenPort, "listen address")
	flag.StringVar(&endpointName, "EndpointName", endpointName, "endpoint name.")
	flag.Parse()

	initPprofMonitor(listenPort)

	fmt.Printf("%s starting!\n", endpointName)

	config.SetListenPort(listenPort)

	corePtr, coreErr := core.New(endpointName, listenPort)
	if coreErr != nil {
		log.Errorf("create core service failed, err:%s", coreErr.Error())
		return
	}
	err := application.Startup(corePtr)
	if err != nil {
		log.Errorf("application.Startup err:%s", err.Error())
		return
	}

	application.Run()
	application.Shutdown()
}
