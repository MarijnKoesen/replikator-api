package main

import (
	"bytes"
	"fmt"
	"github.com/droundy/goopt"
	"github.com/gorilla/mux"
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

	log.Printf("Creating replikator [%s]", name)

	output := executeWithFormat("--output json --create %s", name)

	fmt.Fprint(w, output)
}

func createReplikatorFromReplica(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	fromReplica := vars["fromReplica"]

	log.Printf("Creating replikator [%s] from replica [%s]", name, fromReplica)

	output := executeWithFormat("--output json --create %s --from-replica %s", name, fromReplica)

	fmt.Fprint(w, output)
}

func stopReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	log.Printf("Stopping replikator [%s]", name)

	output := executeWithFormat("--output json --stop %s", name)

	fmt.Fprint(w, output)
}

func startReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	log.Printf("Starting replikator [%s]", name)

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

	log.Printf("Deleting replikator [%s]", name)

	output := executeWithFormat("--output json --delete %s", name)

	fmt.Fprint(w, output)
}

func startApiServer() {
	r := mux.NewRouter()

	r.HandleFunc("/replikators", listReplikators).Methods("GET")

	r.HandleFunc("/replikator/{name}", createReplikatorFromReplica).Methods("PUT").Queries("fromReplica", "{fromReplica}")
	r.HandleFunc("/replikator/{name}", createReplikator).Methods("PUT")
	r.HandleFunc("/replikator/{name}/stop", stopReplikator).Methods("PUT")
	r.HandleFunc("/replikator/{name}/start", startReplikator).Methods("PUT")
	r.HandleFunc("/replikator/{name}", getReplikator).Methods("GET")
	r.HandleFunc("/replikator/{name}", deleteReplikator).Methods("DELETE")

	log.Printf("Listening on [%s], using replikator executable [%s]\n", *listenAddress, *replikatorPath)

	err := http.ListenAndServe(*listenAddress, r)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	goopt.Description = func() string {
		return "Restfull Replikator API server that allows you to list, create, delete fetch replikators"
	}
	goopt.Version = "0.1.0"
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
