package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jetblack87/maestro/agent/data"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"regexp"
	"path"
	"github.com/samuel/go-zookeeper/zk"
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
var request *processStartRequest

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

	watchChannel := make(chan zk.Event)

	for key := range agent.Processes {
		agent.Processes[key], err = zkdao.LoadProcess(agent.Processes[key].ProcessClass, true)
		if err != nil {
			panic(err)
		}
		agent.Processes[key].Key = "/maestro/"+*domainName+"/runtime/agents/"+agent.Name+"/processes/"+agent.Processes[key].Name
		agent.Processes[key].AdminState = "on"
		agent.Processes[key].OperState = "off"
		
		adminStatePath := agent.Processes[key].Key+"/admin_state"
		fmt.Println("Adding watch to node: " + adminStatePath)
		err := zkdao.Watch(adminStatePath, watchChannel)
		if err != nil {
			fmt.Println("Failed to add watch to process node:\n" + err.Error())
		}
	}

	fmt.Println("Adding agent to runtime configuration")
	err = zkdao.UpdateAgent("/maestro/"+*domainName+"/runtime/agents/"+agent.Name, agent, true)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Starting process monitoring")
	
	request = &processStartRequest{
		processes : agent.Processes,
		commandChan : make(chan *command, len(agent.Processes)),
		resultChan : make(chan *result, len(agent.Processes))}

	go startAndMonitorProcesses(request)

	fmt.Println("looping over channels")
	for {
		select {
			case r := <-request.resultChan:
			fmt.Println("request")
			if r.err != nil {
				fmt.Errorf("An error occured running a process:%s\n", r.err.Error())
			} else {
	    		fmt.Printf("Started process with with: %d\n", r.pid)
			}
			case w := <-watchChannel:
			fmt.Println("watch")
			if w.Type.String() == "EventNodeDataChanged" {
				adminState, err := zkdao.GetValue(w.Path)
				if err != nil {
					fmt.Errorf("Error getting data for path '%s': %s\n", w.Path, err.Error())				
				} else {
					process, err2 := zkdao.LoadProcess(path.Dir(w.Path), true)
					if err2 != nil {
						fmt.Errorf("Error loading process '%s': %s\n", w.Path, err2)
					} else {
						request.commandChan <- &command{process : process, adminState : string(adminState)}
					}
				}
			}
		}
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

func watch(nodepath string, data []byte) {
	fmt.Println("watched node changed")
	var adminStateRegexp = regexp.MustCompile(`admin_state$`)
	if adminStateRegexp.MatchString(nodepath) {
		processPath := path.Dir(nodepath)
		adminState := string(data)
		if request != nil {
			process, err2 := zkdao.LoadProcess(processPath, true)
			if err2 != nil {
				fmt.Errorf("Error loading process '%s': %s\n", processPath, err2)
			} else {
				request.commandChan <- &command{process : process, adminState : adminState}
			}
		}
	}
}

func startAndMonitorProcesses (startRequest *processStartRequest) {
	// Mapping of process key (string) to command
	processMap := make(map[string]*exec.Cmd)
	// Start all of the processes
	for key := range startRequest.processes {
		cmd, err := startProcess(startRequest.processes[key])
		if err != nil {
			fmt.Errorf("Error starting process:\n%s", err.Error())
			startRequest.resultChan <- &result{err : err}
		} else {
			processMap[startRequest.processes[key].Key] = cmd
			// Send the result back
			startRequest.resultChan <- &result{pid : cmd.Process.Pid}
		}
	}
	
	for c := range startRequest.commandChan {
    	switch c.adminState {
    	case "off":
    		fmt.Println("Killing process: " + c.process.Key)
    		if processMap[c.process.Key] != nil {
		        processMap[c.process.Key].Process.Kill()
		        processMap[c.process.Key] = nil
		    } else {
		    	fmt.Println("Process is already stopped: " + c.process.Key)
		    }
    	case "on":
        	if processMap[c.process.Key] == nil {
	        	fmt.Println("Starting process: " + c.process.Key)
        		cmd, err := startProcess(c.process)
				if err != nil {
					fmt.Errorf("Error starting process:\n%s", err.Error())
					startRequest.resultChan <- &result{err : err}
				} else {
					processMap[c.process.Key] = cmd
					// Send the result back
					startRequest.resultChan <- &result{pid : cmd.Process.Pid}
				}
        	} else {
        		fmt.Println("Process is already running: " + c.process.Key)
        	}
    	}
	}
}

func startProcess (process data.Process) (*exec.Cmd, error) {
	fmt.Println("Process command: " + process.Command)
	var cmd *exec.Cmd
	if process.Arguments != "" {
		cmd = exec.Command(process.Command, process.Arguments)
	} else {
		cmd = exec.Command(process.Command)
	}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

type processStartRequest struct {
	processes[] data.Process
	commandChan chan *command
	resultChan chan *result
}

type command struct {
	process data.Process
	adminState string
}

type result struct {
	rc int
	pid int
	err error
}
