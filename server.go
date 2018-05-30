package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/intervention-engine/fhir/auth"
	"github.com/intervention-engine/fhir/server"
	"github.com/mitre/fhir-server/middleware"
)

func main() {
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	flag.Parse()

	mongoHost, mongoHostDefined := os.LookupEnv("MONGO_PORT_27017_TCP_ADDR")
	if !mongoHostDefined {
		mongoHost = "localhost"
	}
	mongoPort, mongoPortDefined := os.LookupEnv("MONGO_PORT_27017_TCP_PORT")
	if !mongoPortDefined {
		mongoPort = "27017"
	}

	mongoHostPort := fmt.Sprintf("%s:%s", mongoHost, mongoPort)
	fmt.Printf("Mongo host: %s\n", mongoHostPort)

	var MyConfig = server.Config{
		ServerURL:             "http://localhost:3001",
		IndexConfigPath:       "config/indexes.conf",
		DatabaseHost:          mongoHostPort,
		DatabaseName:          "fhir",
		DatabaseSocketTimeout: 2 * time.Minute,
		DatabaseOpTimeout:     90 * time.Second,
		DatabaseKillOpPeriod:  10 * time.Second,
		Auth:                  auth.None(),
		EnableCISearches:      true,
		CountTotalResults:     true,
		ReadOnly:              false,
		EnableXML:             true,
		Debug:                 true,
	}
	s := server.NewServer(MyConfig)
	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}
	// s.Engine.Use(server.RequestLoggerHandler)

	// Mutex middleware to work around the lack of proper transactions in MongoDB (at least until MongoDB 4.0)
	s.Engine.Use(client_specified_mutexes.Middleware())

	s.Run()
}
