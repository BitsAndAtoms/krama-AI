package main

import (
	"log"
	"net/http"

	"github.com/romana/rlog"
)

func main() {

	loadEnvironmentVariables()

	if !loadSystemProperties() {
		return
	}

	rlog.Debug("Attempting to connect to base components - [ELASTIC, REDIS, MONGO]...")

	connectDB()
	connectRedis()
	connectElastic()

	rlog.Info("Base components - [ELASTIC, REDIS, MONGO] connected ...")

	log.Fatal(http.ListenAndServe(":"+APIPort, routers()))

}
