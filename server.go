package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/eug48/fhir-server/middleware"
	"github.com/eug48/fhir/auth"
	"github.com/eug48/fhir/server"
)

func main() {
	port := flag.Int("port", 3001, "Port to listen on")
	reqLog := flag.Bool("reqlog", false, "Enables request logging -- do NOT use in production")
	mongodbHostPort := flag.String("mongodbHostPort", "localhost:27017", "MongoDB host:port")
	databaseName := flag.String("databaseName", "fhir", "MongoDB database name to use")
	enableXML := flag.Bool("enableXML", false, "Enable support for the FHIR XML encoding")
	flag.Parse()

	fmt.Printf("MongoDB host: %s\n", *mongodbHostPort)

	var MyConfig = server.Config{
		ServerURL:             fmt.Sprintf("http://localhost:%d", *port),
		IndexConfigPath:       "config/indexes.conf",
		DatabaseHost:          *mongodbHostPort,
		DatabaseName:          *databaseName,
		DatabaseSocketTimeout: 2 * time.Minute,
		DatabaseOpTimeout:     90 * time.Second,
		DatabaseKillOpPeriod:  10 * time.Second,
		Auth:                  auth.None(),
		EnableCISearches:      true,
		CountTotalResults:     true,
		ReadOnly:              false,
		EnableXML:             *enableXML,
		EnableHistory:         true,
		Debug:                 true,
	}
	s := server.NewServer(MyConfig)
	if *reqLog {
		s.Engine.Use(server.RequestLoggerHandler)
	}
	// s.Engine.Use(server.RequestLoggerHandler)

	// Mutex middleware to work around the lack of proper transactions in MongoDB (at least until MongoDB 4.0)
	s.Engine.Use(client_specified_mutexes.Middleware())
	s.Engine.StaticFile("metadata", "capability_statement.json")

	s.Run()
}
