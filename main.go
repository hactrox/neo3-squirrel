package main

import (
	"flag"
	"fmt"
	"neo3-squirrel/config"
	"neo3-squirrel/log"
	"net/http"
)

var pprofEnabled bool
var pprofPort int

func init() {
	flag.BoolVar(&pprofEnabled, "pprof", false, "enable pprof")
	flag.IntVar(&pprofPort, "p", 6060, "pprof port number")
}

func main() {
	flag.Parse()
	config.Load(true)
	log.Init(config.DebugMode())

	if pprofEnabled {
		enablePProf()
	}

	select {}
}

func enablePProf() {
	if pprofPort < 1 || pprofPort > 65535 {
		panic("Incorrect pprof port")
	}

	go func() {
		url := fmt.Sprintf("localhost:%d", pprofPort)
		log.Debug(http.ListenAndServe(url, nil))
	}()
}
