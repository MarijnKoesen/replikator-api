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

func listReplikators(w http.ResponseWriter, r *http.Request) {
	output := execute("-l -o json")

	fmt.Fprintf(w, output)
}

func createReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	log.Printf("Creating replikator: %s", name)
	fmt.Printf("Creating2 replikator: %s", name)

	output := execute("-o json --create " + name)

	fmt.Fprintf(w, output)
}

func getReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	output := execute("-o json --get-status " + name)

	fmt.Fprintf(w, output)
}

func deleteReplikator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	log.Printf("Deleting replikator: %s", name)
	fmt.Printf("Deleting2 replikator: %s", name)

	output := execute("-o json --delete " + name)

	fmt.Fprintf(w, output)
}

func startApiServer() {
	r := mux.NewRouter()

	r.HandleFunc("/replikators", listReplikators).Methods("GET")

	r.HandleFunc("/replikator/{name}", createReplikator).Methods("PUT")
	r.HandleFunc("/replikator/{name}", getReplikator).Methods("GET")
	r.HandleFunc("/replikator/{name}", deleteReplikator).Methods("DELETE")

	log.Printf("Listening on '%s', using replikator executable '%s'\n", *listenAddress, *replikatorPath)

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

		log.Printf("Shutting down, signal '%s' received...", sig)
		os.Exit(1)
	}()
}
