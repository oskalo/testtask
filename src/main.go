package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Oleg-Skalozub/testtask/src/infrastructure/config"
	"github.com/Oleg-Skalozub/testtask/src/infrastructure/load"
	"github.com/Oleg-Skalozub/testtask/src/infrastructure/logger"
	"github.com/Oleg-Skalozub/testtask/src/router"

	"github.com/urfave/negroni"
)

var configFile *string
var logFatal = log.New(os.Stderr, "ERROR:\n", 0).Fatalf

func init() {
	configFile = flag.String("config", "", "Configuration file in JSON-format")
}

func main() {
	flag.Parse()

	if len(*configFile) > 0 {
		config.FilePath = *configFile
	}

	err := load.LoadApplicationServices()
	if err != nil {
		logFatal("Failed to initialize : %v", err)
	}

	r := router.NewRouter()

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	rLog := negroni.NewLogger()
	rLog.SetFormat("[{{.Status}} {{.Duration}} {{.Method}}  {{.Path}}] - {{.Request.UserAgent}}")
	rLog.ALogger = logger.Log

	n.Use(rLog)
	n.UseHandler(r)

	log.Fatal(http.ListenAndServe(":3000", n))
}