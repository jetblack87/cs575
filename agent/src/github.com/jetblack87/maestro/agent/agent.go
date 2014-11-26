package main

import (
	"encoding/json"
	"flag"
	"github.com/jetblack87/maestro/agent/data"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"path"
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"log"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var zookeeper *string = flag.String("zookeeper", "localhost:2181", "The ZooKeeper connection string (defaults to 'localhost:2182').")
var agentName *string = flag.String("name", "", "REQUIRED: The name of the agent.")
var domainName *string = flag.String("domain", "", "REQUIRED: The name of the domain in which this agent lives.")
var agentConfig *string = flag.String("agentConfig", "", "Supply a json file that contains specific configuration for this agent.")
var processesConfig *string = flag.String("processesConfig", "", "Supply a json file that contains specific configuration any processes.")
var logfilePath *string = flag.String("logfile", "stdout", "The path to the logfile.")

var zkdao data.ZkDAO
var request *processStartRequest


const MAX_START_RETRIES = 3 

func main() {

	flag.Parse() // Scan the arguments list

	// Setup signal channel
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)


    // Setup logging
	if *logfilePath != "stdout" {
		f, err := os.OpenFile(*logfilePath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
		    log.Printf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Check the parameters
	if *versionFlag {
		log.Println("Version:", APP_VERSION)
		os.Exit(0)
	}
	if *agentName == "" {
		panic("-agent is required")
	}
	if *domainName == "" {
		panic("-domain is required")
	}
	
	log.Printf("maestro agent starting for domain '%s' and agent '%s'\n", *domainName, *domainName)

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

	log.Println("Loading the agent configuration")
	agent, err := zkdao.LoadAgent("/maestro/"+*domainName+"/config/agents/"+*agentName, true)
	if err != nil {
		panic(err)
	}

	// Create channel used for watching ZK nodes
	watchChannel := make(chan zk.Event, 1)

	// Remove old runtime config for this agent
	err = zkdao.RemoveRecursive("/maestro/"+*domainName+"/runtime/agents/"+agent.Name)

	// Add processes to the runtime configuration, adding watches to admin_state
	for key := range agent.Processes {
		log.Println("Loading processes from config: " + agent.Processes[key].ProcessClass)
		agent.Processes[key], err = zkdao.LoadProcess(agent.Processes[key].ProcessClass, true)
		if err != nil {
			panic(err)
		}
		agent.Processes[key].Key = "/maestro/"+*domainName+"/runtime/agents/"+agent.Name+"/processes/"+agent.Processes[key].Name
		if agent.Processes[key].AdminState == "" {
			// Default to on
			log.Println("Defaulting admin_state to 'on'")
			agent.Processes[key].AdminState = "on"
		}
		agent.Processes[key].OperState = "off"
	}

	log.Println("Adding agent to runtime configuration")
	if err != nil {
		log.Printf("Failed to remove agent runtime configuration")
		panic(err)
	}
	err = zkdao.UpdateAgent("/maestro/"+*domainName+"/runtime/agents/"+agent.Name, agent, true)
	if err != nil {
		panic(err)
	}

	// Add watches for all admin_state nodes
	for key := range agent.Processes {
		adminStatePath := agent.Processes[key].Key+"/admin_state"
		log.Println("Adding watch to node: " + adminStatePath)
		err := zkdao.Watch(adminStatePath, watchChannel)
		if err != nil {
			log.Println("Failed to add watch to process node:\n" + err.Error())
		}
	}
	
	log.Println("Starting process monitoring")

	// Create out request (including channels)
	request = &processStartRequest{
		processes : agent.Processes,
		commandChan : make(chan *command, 1),
		resultChan : make(chan *result, 1)}

	go startAndMonitorProcesses(request)

	log.Println("Process monitoring started, waiting on channels")
	for {
		select {
			case w := <-watchChannel:
			if w.Type.String() == "EventNodeDataChanged" {
				adminState, err := zkdao.GetValue(w.Path)
				if err != nil {
					log.Printf("Error getting data for path '%s': %s\n", w.Path, err.Error())
				} else {
					process, err2 := zkdao.LoadProcess(path.Dir(w.Path), true)
					if err2 != nil {
						log.Printf("Error loading process '%s': %s\n", w.Path, err2)
					} else {
						request.commandChan <- &command{process : process, adminState : string(adminState)}
					}
				}
			}
			case r := <-request.resultChan:
			if r.err != nil {
				log.Printf("An error occured running a process:%s\n", r.err.Error())
				// Failed to start, turn off
				r.process.OperState = "off"
				r.process.AdminState = "off"
				zkdao.UpdateProcess(r.process.Key, r.process, false)
			} else {
				var p data.Process
	    		// Update the admin_state and pid in ZK
				if r.process.Key != "" {
					p = r.process
				} else {
					p,err = zkdao.LoadProcess(r.key, false)
					if err != nil {
						log.Printf("Failed to load process: %s\n", err.Error())	
					} else {
						p.OperState = r.operState
					}
				}
				log.Printf("Process '%s' oper_state = '%s'\n", p.Key, p.OperState)
	    		zkdao.UpdateProcess(r.process.Key, r.process, false)
	    		
	    		// Touch the admin_state node to get process turned back on
	    		if p.AdminState == "on" && p.OperState == "off" {
	    			zkdao.SetValue(p.Key + "/admin_state", []byte(p.AdminState))
	    		}
			}
			case <-signalChannel: // FIXME this doesn't seem to work (at least not on Windows)
			log.Println("Received signal")
			a,err := zkdao.LoadAgent("/maestro/"+*domainName+"/runtime/agents/"+agent.Name,true)
			if err != nil {
				log.Println("Error retrieving agent")
			} else {
				for _, process := range a.Processes {
					request.commandChan <- &command{process : process, adminState : "off"}
				}
			}
			os.Exit(0)
		}
	}
}

func loadAgentConfig(agentConfig, agentName, domainName string) error {
	// Load the file if it was supplied
	if agentConfig != "" {
		log.Println("Loading agent config: " + agentConfig)
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
		log.Println("Loading processes config: " + processesConfig)
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

func startAndMonitorProcesses (startRequest *processStartRequest) {
	// Mapping of process key (string) to command
	processMap := make(map[string]*exec.Cmd)
	// Start all of the processes
	for key := range startRequest.processes {
		if startRequest.processes[key].AdminState == "on" {
        	log.Println("Starting process: " + startRequest.processes[key].Key)
			cmd, err := startProcess(startRequest.processes[key])
			if err != nil {
				log.Printf("Error starting process:\n%s\n", err.Error())
				startRequest.resultChan <- &result{process : startRequest.processes[key], err : err}
			} else {
				processMap[startRequest.processes[key].Key] = cmd
				// Send the result back
				startRequest.processes[key].OperState = "on"
				startRequest.processes[key].Pid = cmd.Process.Pid
				startRequest.resultChan <- &result{process : startRequest.processes[key]}
			}
		}
	}
	
	// Monitor the processes
	log.Println("Monitoring command channel and processes")
	for {
		select {
			case c := <-startRequest.commandChan:
		    	switch c.adminState {
		    	case "off":
		    		log.Println("Killing process: " + c.process.Key)
		    		if processMap[c.process.Key] != nil {
				        processMap[c.process.Key].Process.Kill()
				        delete(processMap, c.process.Key)
				    } else {
				    	log.Println("Process is already stopped: " + c.process.Key)
				    }
		    	case "on":
		        	if processMap[c.process.Key] == nil {
			        	log.Println("Starting process: " + c.process.Key)
		        		cmd, err := startProcess(c.process)
						if err != nil {
							log.Printf("Error starting process:\n%s", err.Error())
							c.process.OperState = "off"
							startRequest.resultChan <- &result{process : c.process,
															   err : err}
						} else {
							processMap[c.process.Key] = cmd
							// Send the result back
							c.process.OperState = "on"
							startRequest.resultChan <- &result{process : c.process}
						}
		        	} else {
		        		log.Println("Process is already running: " + c.process.Key)
		        	}
		    	}
			default:
			   // Check running processes
			   log.Println("Checking processes")
			   for key,process := range processMap {
			   	if process.ProcessState != nil && process.ProcessState.Exited() {
			   		startRequest.resultChan <- &result{key : key,
			   										   operState : "off",
			   										   success : process.ProcessState.Success()}
                    delete(processMap, key)
			   	}
			   }
			   time.Sleep(5 * time.Second)		   
		}
	}
}

func startProcess (process data.Process) (*exec.Cmd, error) {
	log.Println("Process command: " + process.Command)
	var cmd *exec.Cmd
	if process.Arguments != "" {
		cmd = exec.Command(process.Command, process.Arguments)
	} else {
		cmd = exec.Command(process.Command)
	}
	success := false
	var err error
	for i:=0; i<MAX_START_RETRIES && !success; i++ {
		log.Println("Attempting to start: " + process.Name)
		err = cmd.Start()
		// Create new thread to wait on this process in order to reap it
		go cmd.Wait()
		if err == nil {
			success = true
		}
        if !success {
	        time.Sleep(5 * time.Second)		   
        }
	}
	return cmd, err
}

// Private structures for communication

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
	key string
	operState string
	process data.Process
	success bool
	err error
}
