package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jetblack87/maestro/agent/data"
	"io/ioutil"
	"os"
	"strings"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var zookeeper *string = flag.String("zookeeper", "localhost:2181", "The ZooKeeper connection string (defaults to 'localhost:2182').")
var agentName *string = flag.String("name", "", "REQUIRED: The name of the agent.")
var domainName *string = flag.String("domain", "", "REQUIRED: The name of the domain in which this agent lives.")
var agentConfig *string = flag.String("agentConfig", "", "Supply a json file that contains specific configuration for this agent.")
var processesConfig *string = flag.String("processesConfig", "", "Supply a json file that contains specific configuration any processes.")

var zkdao data.ZkDAO

func main() {
	fmt.Println("Hello, world")

	flag.Parse() // Scan the arguments list

	if *versionFlag {
		fmt.Println("Version:", APP_VERSION)
		os.Exit(0)
	}
	if *agentName == "" {
		panic("-agent is required")
	}
	if *domainName == "" {
		panic("-domain is required")
	}

	zkdao, err := data.NewZkDAO(strings.Split(*zookeeper, ","))
	if err != nil {
		panic(err)
	}

	// Load config into ZK
	err = loadAgentConfig(*agentConfig, *agentName, *domainName)
	if err != nil {
		panic(err)
	}

	// Load processes data into ZK
	err = loadProcessesConfig(*processesConfig, *domainName)
	if err != nil {
		panic(err)
	}

	agent, err := zkdao.LoadAgent("/maestro/"+*domainName+"/config/agents/"+*agentName, true)
	if err != nil {
		panic(err)
	}
	
	for _, process := range agent.Processes {
		fmt.Println("process: " + process.Name)
		fmt.Println("process_class: " + process.ProcessClass)
	}
}

func loadAgentConfig(agentConfig, agentName, domainName string) error {
	// Load the file if it was supplied
	if agentConfig != "" {
		jsonData, err := ioutil.ReadFile(agentConfig)
		if err != nil {
			return err
		}
		var agent data.Agent
		err = json.Unmarshal(jsonData, &agent)
		if err != nil {
			return err
		}
		err = zkdao.UpdateAgent(
			"/maestro/"+domainName+"/config/agents/"+agentName,
			agent, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadProcessesConfig(processesConfig, domainName string) error {
	// Load the file if it was supplied
	if processesConfig != "" {
		jsonData, err := ioutil.ReadFile(processesConfig)
		if err != nil {
			return err
		}
		var processes []data.Process
		err = json.Unmarshal(jsonData, &processes)
		if err != nil {
			return err
		}
		for _, process := range processes {
			err = zkdao.UpdateProcess(
				"/maestro/"+domainName+"/config/processes/"+process.Name,
				process, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func watch(path string) {
}
