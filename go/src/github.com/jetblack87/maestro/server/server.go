package main

import (
	"encoding/json"
	"flag"
	"github.com/jetblack87/maestro/data"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"io/ioutil"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var zookeeper *string = flag.String("zookeeper", "localhost:2181", "The ZooKeeper connection string (defaults to 'localhost:2182').")
var port *int = flag.Int("port", 8080, "Port on which to listen.")
var logfilePath *string = flag.String("logfile", "stdout", "The path to the logfile.")

var zkdao data.ZkDAO

const PRETTY_PRINT_PARAM = "pretty"

func main() {
	flag.Parse() // Scan the arguments list

	// Setup logging
	if *logfilePath != "stdout" {
		f, err := os.OpenFile(*logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if *versionFlag {
		log.Println("Version:", APP_VERSION)
		os.Exit(0)
	}

	zkdao, err := data.NewZkDAO(strings.Split(*zookeeper, ","))
	if err != nil {
		panic(err)
	}

	// Setup handlers
	dh := domainHandler{zkdao: zkdao}
	http.Handle("/domains/", dh)
	ph := processesHandler{zkdao: zkdao}
	http.Handle("/processes/", ph)

	log.Fatal(http.ListenAndServe(":"+strconv.FormatInt(int64(*port), 10), nil))
}

type domainHandler struct{ zkdao *data.ZkDAO }

func (dh domainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP '%s' request for url '%s'", r.Method, r.URL)
	domainKeyRegexp := regexp.MustCompile("/domains/(.*)")
	domainKey := string(domainKeyRegexp.FindSubmatch([]byte(r.URL.Path))[1])
	switch r.Method {
	case "GET":
		dh.getDomains(domainKey, w, r)
	case "PATCH":
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed: " + r.Method))
	}
}

func (dh domainHandler) getDomains(domainKey string, w http.ResponseWriter, r *http.Request) {
	var domains []data.Domain
	var err error
	if domainKey == "" {
		domains, err = dh.zkdao.LoadDomains(data.PathToKey("/maestro"), true)
	} else {
		var domain data.Domain
		domain, err = dh.zkdao.LoadDomain(domainKey, true)
		domains = []data.Domain{domain}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errMsg := "Error occurred retrieving domain"
		w.Write([]byte(errMsg))
		log.Println(errMsg + "\n" + err.Error())
	} else {
		var responseJson []byte
		if r.URL.Query().Get(PRETTY_PRINT_PARAM) == "true" {
			responseJson, err = json.MarshalIndent(domains, "", "   ")
		} else {
			responseJson, err = json.Marshal(domains)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Error occurred retrieving domain"
			w.Write([]byte(errMsg))
			log.Println(errMsg + "\n" + err.Error())
		} else {
			w.Write(responseJson)
		}
	}
}

type processesHandler struct{ zkdao *data.ZkDAO }

func (ph processesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP '%s' request for url '%s'", r.Method, r.URL)
	processesKeyRegexp := regexp.MustCompile("/processes/(.*)")
	processKey := string(processesKeyRegexp.FindSubmatch([]byte(r.URL.Path))[1])
	switch r.Method {
	case "GET":
		ph.getProcess(processKey, w, r)
	case "PATCH":
	    requestJson, err := ioutil.ReadAll(r.Body)
	    if err != nil {
			w.Write([]byte("Bad request: " + err.Error()))
			w.WriteHeader(http.StatusBadRequest)   	
	    } else {
			ph.updateProcess(processKey, requestJson, w, r) 	    	
	    }
	default:
		w.Write([]byte("Method not allowed: " + r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ph processesHandler) getProcess(processKey string, w http.ResponseWriter, r *http.Request) {
	if processKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := "Process key is required"
		w.Write([]byte(errMsg))
		log.Println(errMsg)
	} else {
		process, err := ph.zkdao.LoadProcess(processKey, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := "Error occurred retrieving process"
			w.Write([]byte(errMsg))
			log.Println(errMsg + "\n" + err.Error())
		} else {
			var responseJson []byte
			if r.URL.Query().Get(PRETTY_PRINT_PARAM) == "true" {
				responseJson, err = json.MarshalIndent(process, "", "   ")
			} else {
				responseJson, err = json.Marshal(process)
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errMsg := "Error occurred retrieving process"
				w.Write([]byte(errMsg))
				log.Println(errMsg + "\n" + err.Error())
			} else {
				w.Write(responseJson)
			}
		}
	}
}

func (ph processesHandler) updateProcess(processKey string, requestJson []byte, w http.ResponseWriter, r *http.Request) {
	log.Printf("Process request for process '%s' with body:\n%s\n", processKey, string(requestJson)) 
	if processKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := "Process key is required"
		w.Write([]byte(errMsg))
		log.Println(errMsg)
	} else {
		var process data.Process
		err := json.Unmarshal(requestJson, &process)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errMsg := "Malformed request:\n" + err.Error()
			w.Write([]byte(errMsg))
			log.Println(errMsg)
		} else {
			err = ph.zkdao.UpdateProcess(processKey, process, true)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errMsg := "Internal server error while trying to update process"
				w.Write([]byte(errMsg))
				log.Println(errMsg + "\n" + err.Error())
			}
			ph.getProcess(processKey, w, r)
		}
	}
}
