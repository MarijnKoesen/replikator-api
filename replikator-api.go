package main

import (
	"bytes"
	"fmt"
	"github.com/droundy/goopt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var mutex sync.Mutex

var listenAddress = goopt.String([]string{"-l", "--listen"}, ":8080", "listen address")
var replikatorPath = goopt.String([]string{"-r", "--replikator"}, "sudo replikator-ctl", "Path to replikator-ctl")

func execute(parameters string) string {
	args := strings.Fields(*replikatorPath + " " + parameters)
	cmd := exec.Command(args[0], args[1:]...)

	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut

	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	// replikator-ctl can only be run in a single thread, so use a mutex to make sure we never
	// execute the script from multiple threads when we get multiple api connections
	mutex.Lock()
	err := cmd.Run()
	mutex.Unlock()

	if err != nil {
		return stdErr.String()
	}

	return stdOut.String()
}

func executeWithFormat(format string, arguments ...interface{}) string {
	return execute(fmt.Sprintf(format, arguments...))
}

func listReplikators(w http.ResponseWriter, r *http.Request) {
	output := execute("--output json --list")

	fmt.Fprint(w, output)
}

func createReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := executeWithFormat("--output json --create %s", name)

	fmt.Fprint(w, output)
}

func createReplikatorFromReplica(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	fromReplica := vars["fromReplica"]

	output := executeWithFormat("--output json --create %s --from-replica %s", name, fromReplica)

	fmt.Fprint(w, output)
}

func stopReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := executeWithFormat("--output json --stop %s", name)

	fmt.Fprint(w, output)
}

func startReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := executeWithFormat("--output json --run %s", name)

	fmt.Fprint(w, output)
}

func getReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := executeWithFormat("--output json --get-status %s", name)

	fmt.Fprint(w, output)
}

func deleteReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := executeWithFormat("--output json --delete %s", name)

	fmt.Fprint(w, output)
}

func wrapHandler(handler http.HandlerFunc) http.Handler {
	return promhttp.InstrumentHandlerCounter(httpRequestsTotal, http.HandlerFunc(handler))
}

func startApiServer() {
	registerMetrics()

	router := mux.NewRouter()
	router.Handle("/replikators", wrapHandler(listReplikators)).Methods(http.MethodGet)
	router.Handle("/replikator/{name}", wrapHandler(createReplikatorFromReplica)).Methods(http.MethodPut).Queries("fromReplica", "{fromReplica}")
	router.Handle("/replikator/{name}", wrapHandler(createReplikator)).Methods(http.MethodPut)
	router.Handle("/replikator/{name}/stop", wrapHandler(stopReplikator)).Methods(http.MethodPut)
	router.Handle("/replikator/{name}/start", wrapHandler(startReplikator)).Methods(http.MethodPut)
	router.Handle("/replikator/{name}", wrapHandler(getReplikator)).Methods(http.MethodGet)
	router.Handle("/replikator/{name}", wrapHandler(deleteReplikator)).Methods(http.MethodDelete)
	router.Handle("/metrics", getMetrics()).Methods(http.MethodGet)

	router.Use(loggingMiddleware)
	router.Use(jsonHeaderMiddleware)

	log.Printf("Listening on [%s], using replikator executable [%s]\n", *listenAddress, *replikatorPath)

	err := http.ListenAndServe(*listenAddress, router)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		os.Exit(1)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.Method, r.RequestURI, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

func jsonHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(w, r)
	})
}

func main() {
	goopt.Description = func() string {
		return "Restfull Replikator API server that allows you to list, create, delete fetch replikators"
	}
	goopt.Version = "0.2.0"
	goopt.Summary = "Restfull Replikator API server"
	goopt.Parse(nil)

	setupSignalHandler()
	startApiServer()
}

func setupSignalHandler() {
	// setup signal catching
	stopSignals := make(chan os.Signal, 1)

	// catch only stop signals
	signal.Notify(stopSignals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// method invoked upon seeing any of the stopSignals
	go func() {
		sig := <-stopSignals

		log.Printf("Shutting down, signal [%s] received...", sig)
		os.Exit(1)
	}()
}
