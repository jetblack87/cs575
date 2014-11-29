package data

import (
	"github.com/samuel/go-zookeeper/zk"
	"path"
	"time"
	"strconv"
	"log"
	"errors"
	"bytes"
	"encoding/base64" 
)

type ZkDAO struct {
	client *zk.Conn
}

// #### CONSTRUCTOR ####

func NewZkDAO(zookeeper []string) (*ZkDAO, error) {
	client, _, err := zk.Connect(zookeeper, time.Second)
	zkdao := new(ZkDAO)
	zkdao.client = client
	return zkdao, err
}


// #### PUBLIC METHODS ####

func (zkdao *ZkDAO) LoadDomains(key string, recursive bool) ([]Domain, error) {
	nodepath := KeyToPath(key)
	var domains[] Domain
	domainNodes,_,_ := zkdao.client.Children(nodepath)
	for _,domainNode := range domainNodes {
		domain, err := zkdao.LoadDomain(PathToKey(nodepath + "/" + domainNode), recursive)
		if err != nil { return domains, err }
		domains = append(domains, domain) 
	}
	return domains, nil
}

func (zkdao *ZkDAO) LoadDomain(key string, recursive bool) (Domain, error) {
	nodepath := KeyToPath(key)
	var domain Domain
	domain.Key  = key
	domain.Name = path.Base(nodepath)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if exists {
		config, err := zkdao.LoadStaticConfig(PathToKey(nodepath + "/config"), recursive)
		if err != nil { return domain, err }
		domain.Config = config
		runtime, err := zkdao.LoadRuntimeConfig(PathToKey(nodepath + "/runtime"), recursive)
		if err != nil { return domain, err }
		domain.Runtime = runtime
	} else {
        log.Println("Domain node does not exist: " + string(key))
	}
	return domain, nil
}

func (zkdao *ZkDAO) LoadStaticConfig(key string, recursive bool) (StaticConfig, error) {
	nodepath := KeyToPath(key)
	var config StaticConfig
	exists,_,_ := zkdao.client.Exists(nodepath)
	if exists {
		agentsNode,_,_ := zkdao.client.Children(nodepath + "/agents")
		for _,agentNode := range agentsNode {
			agent, err := zkdao.LoadAgent(PathToKey(nodepath + "/agents/" + agentNode), recursive)
			if err != nil { return config, err }
			config.Agents = append(config.Agents, agent) 
		}
		
		processesNode,_,_ := zkdao.client.Children(nodepath + "/processes")
		for _,processNode := range processesNode {
			process, err := zkdao.LoadProcess(PathToKey(nodepath + "/processes/" + processNode), recursive)
			if err != nil { return config, err }
			config.Processes = append(config.Processes, process) 
		}
	} else {
        log.Println("Static config node does not exist: " + nodepath)
	}
	return config, nil
}

func (zkdao *ZkDAO) LoadRuntimeConfig(key string, recursive bool) (RuntimeConfig, error) {
	nodepath := KeyToPath(key)
	var runtime RuntimeConfig
	exists,_,_ := zkdao.client.Exists(nodepath)
	if exists {
		agentsNode,_,_ := zkdao.client.Children(nodepath + "/agents")
		for _,agentNode := range agentsNode {
			agent, err := zkdao.LoadAgent(PathToKey(nodepath + "/agents/" + agentNode), recursive)
			if err != nil { return runtime, err }
			runtime.Agents = append(runtime.Agents, agent) 
		}
	} else {
        log.Println("Runtime config node does not exist: " + nodepath)
	}
	return runtime, nil
}

func (zkdao *ZkDAO) LoadAgent(key string, recursive bool) (Agent, error) {
	nodepath := KeyToPath(key)
	var agent Agent
	agent.Key  = key
	agent.Name = path.Base(nodepath)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if exists {
		processesNode,_,_ := zkdao.client.Children(nodepath + "/processes")
		for _,processNode := range processesNode {
			process, err := zkdao.LoadProcess(PathToKey(nodepath + "/processes/" + processNode), recursive)
			if err != nil { return agent, err }
			agent.Processes = append(agent.Processes, process) 
		}

		exists,_,_ = zkdao.client.Exists(nodepath + "/eph")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/eph")
			if err == nil {
				agent.Eph = string(data)
			}
		}
	} else {
        log.Println("Agent node does not exist: " + nodepath)
	}
	return agent, nil
}

func (zkdao *ZkDAO) LoadProcess(key string, recursive bool) (Process, error) {
	nodepath := KeyToPath(key)
	process := Process {Pid : -1}
	process.Key  = key
	process.Name = path.Base(nodepath)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if exists {
		exists,_,_ = zkdao.client.Exists(nodepath + "/command")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/command")
			if err == nil {
				process.Command = string(data)
			}
			log.Println("Reading command " + process.Command)
		}
		exists,_,_ = zkdao.client.Exists(nodepath + "/arguments")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/arguments")
			if err == nil {
				process.Arguments = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(nodepath + "/process_class")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/process_class")
			if err == nil {
				process.ProcessClass = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(nodepath + "/admin_state")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/admin_state")
			if err == nil {
				process.AdminState = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(nodepath + "/oper_state")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/oper_state")
			if err == nil {
				process.OperState = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(nodepath + "/pid")
		if exists { 
			data,_,err := zkdao.client.Get(nodepath + "/pid")
			if err == nil {
				tempPid, err := strconv.ParseInt(string(data), 10, 32)
				if err != nil {
					log.Printf("Failed to parse pid '%s':\n%s", nodepath + "/pid", err)
				} else {
					process.Pid = int(tempPid)
				}
			}
		}
	} else {
        log.Println("Process node does not exist: " + nodepath)
	}
	return process, nil
}

func (zkdao *ZkDAO) UpdateDomain(key string, domain Domain, recursive bool) error {
	nodepath := KeyToPath(key)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if !exists {
		_, err := zkdao.createWithParents(nodepath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		err := zkdao.UpdateStaticConfig(PathToKey(nodepath + "/config"), domain.Config, recursive)
		if err != nil { return err }
		err = zkdao.UpdateRuntimeConfig(PathToKey(nodepath + "/runtime"), domain.Runtime, recursive)
		if err != nil { return err }
	}
	return nil
}

func (zkdao *ZkDAO) UpdateStaticConfig(key string, config StaticConfig, recursive bool) error {
	nodepath := KeyToPath(key)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if !exists {
		_, err := zkdao.createWithParents(nodepath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		for _, value := range config.Agents {
			err := zkdao.UpdateAgent(PathToKey(nodepath + "/agents/" + value.Name), value, recursive)
			if err != nil { return err }
		}
		for _, value := range config.Processes {
			err := zkdao.UpdateProcess(PathToKey(nodepath + "/processes/" + value.Name), value, recursive)
			if err != nil { return err }
		}
	}
	return nil
}

func (zkdao *ZkDAO) UpdateRuntimeConfig(key string, runtime RuntimeConfig, recursive bool) error {
	return nil
}

func (zkdao *ZkDAO) UpdateAgent(key string, agent Agent, recursive bool) error {
	nodepath := KeyToPath(key)
	log.Println("Updating agent: " + nodepath)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if !exists {
		_, err := zkdao.createWithParents(nodepath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		for _, value := range agent.Processes {
			err := zkdao.UpdateProcess(PathToKey(nodepath + "/processes/" + value.Name), value, recursive)
			if err != nil { return err }
		}
	}
	return nil
}

func (zkdao *ZkDAO) UpdateProcess(key string, process Process, recursive bool) error {
	nodepath := KeyToPath(key)
	log.Println("Updating process: " + nodepath)
	exists,_,_ := zkdao.client.Exists(nodepath)
	if !exists {
		_, err := zkdao.createWithParents(nodepath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Command != "" {
		_, err := zkdao.createOrSet(nodepath + "/command", []byte(process.Command), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Arguments != "" {
		_, err := zkdao.createOrSet(nodepath + "/arguments", []byte(process.Arguments), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.ProcessClass != "" {
		_, err := zkdao.createOrSet(nodepath + "/process_class", []byte(process.ProcessClass), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.AdminState != "" {
		_, err := zkdao.createOrSet(nodepath + "/admin_state", []byte(process.AdminState), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.OperState != "" {
		_, err := zkdao.createOrSet(nodepath + "/oper_state", []byte(process.OperState), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Pid != -1 {
		_, err := zkdao.createOrSet(nodepath + "/pid", []byte(strconv.FormatInt(int64(process.Pid), 10)), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	return nil
}

func (zkdao *ZkDAO) Watch(path string, watchChannel chan<- zk.Event) (error) {
	log.Println("Adding watch: " + path)
	exists,_,eventChan,err := zkdao.client.ExistsW(path)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("The path '" + path + "' does not exist" )
	}
	go func(to chan<- zk.Event, from <-chan zk.Event) {
		e := <-from
		to <-e
		zkdao.Watch(path, to)
	}(watchChannel, eventChan)
	return nil
}

func (zkdao *ZkDAO) GetValue(path string) ([]byte, error) {
	data,_,err := zkdao.client.Get(path)
	return data,err;
}
func (zkdao *ZkDAO) SetValue(path string, data []byte) (error) {
	exists,stat,err := zkdao.client.Exists(path)
	if err != nil {
		return err
	}
	if exists {
		_, err = zkdao.client.Set(path, data, stat.Version)
		if err != nil {
			return err
		}
	}
	return nil
}

func (zkdao *ZkDAO) RemoveRecursive(path string) (error) {
	exists,stat,err := zkdao.client.Exists(path)
	if exists && err != nil {
		return err
	}
	if exists {
		children,_,_ := zkdao.client.Children(path)
		for _,child := range children {
			err := zkdao.RemoveRecursive(path + "/" + child)
			if err != nil {
				return err
			}
		}
		return zkdao.client.Delete(path, stat.Version)
	}
	return nil
}

// Creates an ephemeral node for the given path
func (zkdao ZkDAO) CreateEphemeral(path string, data []byte) (string, error) {
	return zkdao.client.Create(path,data,zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
}

// Converts a key to a ZK path
func KeyToPath(key string) (string) {
	path, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return ""
	} else {
		return string(path)
	}
}

// Converts a ZK path to a key
func PathToKey(path string) (string) {
	return base64.StdEncoding.EncodeToString([]byte(path))
}

// #### PRIVATE METHODS ####

func (zkdao *ZkDAO) createOrSet(nodepath string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	exists,stat,_ := zkdao.client.Exists(nodepath)
	if exists {
		oldData,_,_ := zkdao.client.Get(nodepath)
		if bytes.Compare(oldData, data) != 0 {
			_, err2 := zkdao.client.Set(nodepath, data, stat.Version)
			return "", err2
		}
	} else {
		return zkdao.client.Create(nodepath, data, flags, acl)
	}
	return "", nil
}

func (zkdao *ZkDAO) createWithParents(nodepath string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	parent := path.Dir(nodepath)
	if parent == "." {
		return "", nil
	}
	exists, _, err := zkdao.client.Exists(parent)
	if err != nil {
		return "", err
	}
	if !exists {
		s, err := zkdao.createWithParents(parent, data, flags, acl)
		if err != nil {
			return s, err
		}
	}
	return zkdao.client.Create(nodepath, data, flags, acl)
}
