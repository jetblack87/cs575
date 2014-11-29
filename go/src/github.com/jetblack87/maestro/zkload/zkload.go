package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jetblack87/maestro/data"
	"io/ioutil"
	"strings"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var zookeeper *string = flag.String("zookeeper", "localhost:2181", "The ZooKeeper connection string (defaults to 'localhost:2182').")
var filename *string = flag.String("file", "maestro_data.json", "Supply the file to load.")
var dump *bool = flag.Bool("dump", false, "Dumps the zookeeper config.")

func main() {
	flag.Parse() // Scan the arguments list

	if *versionFlag {
		fmt.Println("Version:", APP_VERSION)
	}

	if *dump {
		DumpFile(*zookeeper)	
	} else {
		jsonData, err := ioutil.ReadFile(*filename)
		if err != nil {
			panic(err)
		}
		LoadFile(*zookeeper, jsonData)
	}
}

func LoadFile(zookeeper string, jsonData []byte) {
	var domain data.Domain
	err := json.Unmarshal(jsonData, &domain)
	if err != nil {
		panic(err)
	}
	zkdao, err := data.NewZkDAO(strings.Split(zookeeper, ","))
	if err != nil {
		panic(err)
	}
	err = zkdao.UpdateDomain(data.PathToKey("/maestro/" + domain.Name), domain, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("Completed load successfully")
}

func DumpFile(zookeeper string) {
	zkdao, err := data.NewZkDAO(strings.Split(zookeeper, ","))
	if err != nil {
		panic(err)
	}
	domains, err := zkdao.LoadDomains(data.PathToKey("/maestro"), true)
		if err != nil {
		panic(err)
	}
	json, err := json.MarshalIndent(domains,""," ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}
