//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	go_http "net/http"
	_ "net/http/pprof"

	"github.com/couchbase/query/accounting"
	acct_resolver "github.com/couchbase/query/accounting/resolver"
	config_resolver "github.com/couchbase/query/clustering/resolver"
	datastore_package "github.com/couchbase/query/datastore"
	"github.com/couchbase/query/datastore/resolver"
	"github.com/couchbase/query/datastore/system"
	"github.com/couchbase/query/logging"
	log_resolver "github.com/couchbase/query/logging/resolver"
	"github.com/couchbase/query/server"
	"github.com/couchbase/query/server/http"
	"github.com/couchbase/query/util"
	"runtime/debug"
)

var DATASTORE = flag.String("datastore", "", "Datastore address (http://URL or dir:PATH or mock:)")
var CONFIGSTORE = flag.String("configstore", "stub:", "Configuration store address (http://URL or stub:)")
var ACCTSTORE = flag.String("acctstore", "gometrics:", "Accounting store address (http://URL or stub:)")
var NAMESPACE = flag.String("namespace", "default", "Default namespace")
var TIMEOUT = flag.Duration("timeout", 0*time.Second, "Server execution timeout, e.g. 500ms or 2s; use zero or negative value to disable")
var READONLY = flag.Bool("readonly", false, "Read-only mode")
var SIGNATURE = flag.Bool("signature", true, "Whether to provide signature")
var METRICS = flag.Bool("metrics", true, "Whether to provide metrics")
var REQUEST_CAP = flag.Int("request-cap", 1024, "Maximum number of queued requests")
var REQUEST_SIZE_CAP = flag.Int("request-size-cap", server.MAX_REQUEST_SIZE, "Maximum size of a request")
var SCAN_CAP = flag.Int("scan-cap", 0, "Maximum buffer size for primary index scans; use zero or negative value to disable")
var SERVICERS = flag.Int("servicers", 4*runtime.NumCPU(), "Servicer count")
var PLUS_SERVICERS = flag.Int("plus-servicers", 16*runtime.NumCPU(), "Plus servicer count")
var MAX_PARALLELISM = flag.Int("max-parallelism", 1, "Maximum parallelism per query; use zero or negative value to disable")
var ORDER_LIMIT = flag.Int64("order-limit", 0, "Maximum LIMIT for ORDER BY clauses; use zero or negative value to disable")
var MUTATION_LIMIT = flag.Int64("mutation-limit", 0, "Maximum LIMIT for data modification statements; use zero or negative value to disable")
var HTTP_ADDR = flag.String("http", ":8093", "HTTP service address")
var HTTPS_ADDR = flag.String("https", ":18093", "HTTPS service address")
var CERT_FILE = flag.String("certfile", "", "HTTPS certificate file")
var KEY_FILE = flag.String("keyfile", "", "HTTPS private key file")
var SSL_MINIMUM_PROTOCOL = flag.String("ssl_minimum_protocol", "tlsv1", "TLS minimum version ('tlsv1'/'tlsv1.1'/'tlsv1.2')")
var LOGGER = flag.String("logger", "", "Logger implementation")
var LOG_LEVEL = flag.String("loglevel", "info", "Log level: debug, trace, info, warn, error, severe, none")
var DEBUG = flag.Bool("debug", false, "Debug mode")
var KEEP_ALIVE_LENGTH = flag.Int("keep-alive-length", server.KEEP_ALIVE_DEFAULT, "maximum size of buffered result")
var STATIC_PATH = flag.String("static-path", "static", "Path to static content")
var PIPELINE_CAP = flag.Int("pipeline-cap", 512, "Maximum number of items each execution operator can buffer")
var PIPELINE_BATCH = flag.Int("pipeline-batch", 16, "Number of items execution operators can batch")
var ENTERPRISE = flag.Bool("enterprise", true, "Enterprise mode")
var GCPERCENT = flag.Int("gc-percent", getOsGogc(), "Go runtime garbage collection target percentage")

//cpu and memory profiling flags
var CPU_PROFILE = flag.String("cpuprofile", "", "write cpu profile to file")
var MEM_PROFILE = flag.String("memprofile", "", "write memory profile to this file")

// Monitoring API
var COMPLETED_THRESHOLD = flag.Int("completed-threshold", 1000, "cache completed query lasting longer than this many milliseconds")
var COMPLETED_LIMIT = flag.Int("completed-limit", 4000, "maximum number of completed requests")

func getOsGogc() int {

	// since there is no getter for GOGC, we get it to see if its 100, and immediately set it again if it wasn't
	origGCSetting := debug.SetGCPercent(100)

	if origGCSetting != 100 {
		debug.SetGCPercent(origGCSetting)
	}

	return int(origGCSetting)
}

func main() {
	HideConsole(true)
	defer HideConsole(false)

	// useful for getting list of go-routines
	go go_http.ListenAndServe("localhost:6060", nil)

	flag.Parse()

	if *LOGGER != "" {
		logger, _ := log_resolver.NewLogger(*LOGGER)
		if logger == nil {
			fmt.Printf("Invalid logger: %s\n", *LOGGER)
			os.Exit(1)
		}

		logging.SetLogger(logger)
	}

	if *DEBUG {
		logging.SetLevel(logging.DEBUG)
	} else {
		level := logging.INFO

		if *LOG_LEVEL != "" {
			lvl, ok := logging.ParseLevel(*LOG_LEVEL)
			if ok {
				level = lvl
			}
		}

		logging.SetLevel(level)
	}

	datastore, err := resolver.NewDatastore(*DATASTORE)
	if err != nil {
		logging.Errorp(err.Error())
		logging.Errorf("Shutting down.")
		os.Exit(1)
	}
	datastore_package.SetDatastore(datastore)

	configstore, err := config_resolver.NewConfigstore(*CONFIGSTORE)
	if err != nil {
		logging.Errorp("Could not connect to configstore",
			logging.Pair{"error", err},
		)
	}
	acctstore, err := acct_resolver.NewAcctstore(*ACCTSTORE)
	if err != nil {
		logging.Errorp("Could not connect to acctstore",
			logging.Pair{"error", err},
		)
	} else {
		// Create the metrics we are interested in
		accounting.RegisterMetrics(acctstore)
		// Make metrics available
		acctstore.MetricReporter().Start(1, 1)
	}

	// Start the completed requests log
	accounting.RequestsInit(*COMPLETED_THRESHOLD, *COMPLETED_LIMIT)

	channel := make(server.RequestChannel, *REQUEST_CAP)
	plusChannel := make(server.RequestChannel, *REQUEST_CAP)

	sys, err := system.NewDatastore(datastore)
	if err != nil {
		logging.Errorp(err.Error())
		os.Exit(1)
	}

	server, err := server.NewServer(datastore, sys, configstore, acctstore, *NAMESPACE,
		*READONLY, channel, plusChannel, *SERVICERS, *PLUS_SERVICERS,
		*MAX_PARALLELISM, *TIMEOUT, *SIGNATURE, *METRICS, *ENTERPRISE, *GCPERCENT)
	if err != nil {
		logging.Errorp(err.Error())
		os.Exit(1)
	}

	datastore_package.SetSystemstore(server.Systemstore())

	server.SetCpuProfile(*CPU_PROFILE)
	server.SetKeepAlive(*KEEP_ALIVE_LENGTH)
	server.SetMemProfile(*MEM_PROFILE)
	server.SetPipelineCap(*PIPELINE_CAP)
	server.SetPipelineBatch(*PIPELINE_BATCH)
	server.SetRequestSizeCap(*REQUEST_SIZE_CAP)
	server.SetScanCap(*SCAN_CAP)

	if server.Enterprise() && os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if !server.Enterprise() {
		var numCPU int
		if os.Getenv("GOMAXPROCS") == "" {
			numCPU = runtime.NumCPU()
		} else {
			numCPU = runtime.GOMAXPROCS(0)
		}

		// Use at most 4 cpus in non-enterprise mode
		runtime.GOMAXPROCS(util.MinInt(numCPU, 4))
	}

	go server.Serve()
	go server.PlusServe()

	logging.Infop("cbq-engine started",
		logging.Pair{"version", util.VERSION},
		logging.Pair{"datastore", *DATASTORE},
		logging.Pair{"max-concurrency", runtime.GOMAXPROCS(0)},
		logging.Pair{"loglevel", logging.LogLevel().String()},
		logging.Pair{"servicers", server.Servicers()},
		logging.Pair{"plus-servicers", server.PlusServicers()},
		logging.Pair{"gc-percent", server.GcPercent()},
		logging.Pair{"pipeline-cap", server.PipelineCap()},
		logging.Pair{"pipeline-batch", *PIPELINE_BATCH},
		logging.Pair{"request-cap", *REQUEST_CAP},
		logging.Pair{"request-size-cap", server.RequestSizeCap()},
		logging.Pair{"timeout", server.Timeout()},
	)

	// Create http endpoint
	endpoint := http.NewServiceEndpoint(server, *STATIC_PATH, *METRICS,
		*HTTP_ADDR, *HTTPS_ADDR, *CERT_FILE, *KEY_FILE, *SSL_MINIMUM_PROTOCOL)
	er := endpoint.Listen()
	if er != nil {
		logging.Errorp("cbq-engine exiting with error",
			logging.Pair{"error", er},
			logging.Pair{"HTTP_ADDR", *HTTP_ADDR},
		)
		os.Exit(1)
	}
	if server.Enterprise() && *CERT_FILE != "" && *KEY_FILE != "" {
		er := endpoint.ListenTLS()
		if er != nil {
			logging.Errorp("cbq-engine exiting with error",
				logging.Pair{"error", er},
				logging.Pair{"HTTPS_ADDR", *HTTPS_ADDR},
			)
			os.Exit(1)
		}
	}
	signalCatcher(server, endpoint)
}

// signalCatcher blocks until a signal is recieved and then takes appropriate action
func signalCatcher(server *server.Server, endpoint *http.HttpEndpoint) {
	sig_chan := make(chan os.Signal, 4)
	signal.Notify(sig_chan, os.Interrupt, syscall.SIGTERM)

	var s os.Signal
	select {
	case s = <-sig_chan:
	}
	if server.CpuProfile() != "" {
		logging.Infop("Stopping CPU profile")
		pprof.StopCPUProfile()
	}
	if server.MemProfile() != "" {
		f, err := os.Create(server.MemProfile())
		if err != nil {
			logging.Errorp("Cannot create memory profile file", logging.Pair{"error", err})
		} else {

			logging.Infop("Writing  Memory profile")
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}
	if s == os.Interrupt {
		// Interrupt (ctrl-C) => Immediate (ungraceful) exit
		logging.Infop("Shutting down immediately")
		os.Exit(0)
	}
	logging.Infop("Attempting graceful exit")
	// Stop accepting new requests
	err := endpoint.Close()
	if err != nil {
		logging.Errorp("error closing http listener", logging.Pair{"err", err})
	}
	err = endpoint.CloseTLS()
	if err != nil {
		logging.Errorp("error closing https listener", logging.Pair{"err", err})
	}
}
