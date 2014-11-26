package data

import (
	"github.com/samuel/go-zookeeper/zk"
	"path"
	"time"
	"strconv"
	"log"
	"errors"
	"bytes"
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
	var domains[] Domain
	domainNodes,_,_ := zkdao.client.Children(key)
	for _,domainNode := range domainNodes {
		domain, err := zkdao.LoadDomain(key + "/" + domainNode, recursive)
		if err != nil { return domains, err }
		domains = append(domains, domain) 
	}
	return domains, nil
}

func (zkdao *ZkDAO) LoadDomain(key string, recursive bool) (Domain, error) {
	var domain Domain
	domain.Key  = key
	domain.Name = path.Base(key)
	exists,_,_ := zkdao.client.Exists(key)
	if exists {
		config, err := zkdao.LoadStaticConfig(key + "/config", recursive)
		if err != nil { return domain, err }
		domain.Config = config
	} else {
        log.Println("Domain node does not exist: " + key)
	}
	return domain, nil
}

func (zkdao *ZkDAO) LoadStaticConfig(key string, recursive bool) (StaticConfig, error) {
	var config StaticConfig
	exists,_,_ := zkdao.client.Exists(key)
	if exists {
		agentsNode,_,_ := zkdao.client.Children(key + "/agents")
		for _,agentNode := range agentsNode {
			agent, err := zkdao.LoadAgent(key + "/agents/" + agentNode, recursive)
			if err != nil { return config, err }
			config.Agents = append(config.Agents, agent) 
		}
		
		processesNode,_,_ := zkdao.client.Children(key + "/processes")
		for _,processNode := range processesNode {
			process, err := zkdao.LoadProcess(key + "/processes/" + processNode, recursive)
			if err != nil { return config, err }
			config.Processes = append(config.Processes, process) 
		}
	} else {
        log.Println("Static Config node does not exist: " + key)
	}
	return config, nil
}

func (zkdao *ZkDAO) LoadAgent(key string, recursive bool) (Agent, error) {
	var agent Agent
	agent.Key  = key
	agent.Name = path.Base(key)
	exists,_,_ := zkdao.client.Exists(key)
	if exists {
		processesNode,_,_ := zkdao.client.Children(key + "/processes")
		for _,processNode := range processesNode {
			process, err := zkdao.LoadProcess(key + "/processes/" + processNode, recursive)
			if err != nil { return agent, err }
			agent.Processes = append(agent.Processes, process) 
		}
	} else {
        log.Println("Agent node does not exist: " + key)
	}
	return agent, nil
}

func (zkdao *ZkDAO) LoadProcess(key string, recursive bool) (Process, error) {
	process := Process {Pid : -1}
	process.Key  = key
	process.Name = path.Base(key)
	exists,_,_ := zkdao.client.Exists(key)
	if exists {
		exists,_,_ = zkdao.client.Exists(key + "/command")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/command")
			if err == nil {
				process.Command = string(data)
			}
			log.Println("Reading command " + process.Command)
		}
		exists,_,_ = zkdao.client.Exists(key + "/arguments")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/arguments")
			if err == nil {
				process.Arguments = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(key + "/process_class")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/process_class")
			if err == nil {
				process.ProcessClass = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(key + "/admin_state")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/admin_state")
			if err == nil {
				process.AdminState = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(key + "/oper_state")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/oper_state")
			if err == nil {
				process.OperState = string(data)
			}
		}
		exists,_,_ = zkdao.client.Exists(key + "/pid")
		if exists { 
			data,_,err := zkdao.client.Get(key + "/pid")
			if err == nil {
				tempPid, err := strconv.ParseInt(string(data), 10, 32)
				if err != nil {
					log.Printf("Failed to parse pid '%s':\n%s", key + "/pid", err)
				} else {
					process.Pid = int(tempPid)
				}
			}
		}
	} else {
        log.Println("Process node does not exist: " + key)
	}
	return process, nil
}

func (zkdao *ZkDAO) UpdateDomain(key string, domain Domain, recursive bool) error {
	exists,_,_ := zkdao.client.Exists(key)
	if !exists {
		_, err := zkdao.createWithParents(key, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		err := zkdao.UpdateStaticConfig(key + "/config", domain.Config, recursive)
		if err != nil { return err }
		err = zkdao.UpdateRuntimeConfig(key + "/runtime", domain.Runtime, recursive)
		if err != nil { return err }
	}
	return nil
}

func (zkdao *ZkDAO) UpdateStaticConfig(key string, config StaticConfig, recursive bool) error {
	exists,_,_ := zkdao.client.Exists(key)
	if !exists {
		_, err := zkdao.createWithParents(key, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		for _, value := range config.Agents {
			err := zkdao.UpdateAgent(key + "/agents/" + value.Name, value, recursive)
			if err != nil { return err }
		}
		for _, value := range config.Processes {
			err := zkdao.UpdateProcess(key + "/processes/" + value.Name, value, recursive)
			if err != nil { return err }
		}
	}
	return nil
}

func (zkdao *ZkDAO) UpdateRuntimeConfig(key string, runtime RuntimeConfig, recursive bool) error {
	return nil
}

func (zkdao *ZkDAO) UpdateAgent(key string, agent Agent, recursive bool) error {
	log.Println("Updating agent: " + key)
	exists,_,_ := zkdao.client.Exists(key)
	if !exists {
		_, err := zkdao.createWithParents(key, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if recursive {
		for _, value := range agent.Processes {
			err := zkdao.UpdateProcess(key + "/processes/" + value.Name, value, recursive)
			if err != nil { return err }
		}
	}
	return nil
}

func (zkdao *ZkDAO) UpdateProcess(key string, process Process, recursive bool) error {
	log.Println("Updating process: " + key)
	exists,_,_ := zkdao.client.Exists(key)
	if !exists {
		_, err := zkdao.createWithParents(key, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Command != "" {
		_, err := zkdao.createOrSet(key + "/command", []byte(process.Command), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Arguments != "" {
		_, err := zkdao.createOrSet(key + "/arguments", []byte(process.Arguments), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.ProcessClass != "" {
		_, err := zkdao.createOrSet(key + "/process_class", []byte(process.ProcessClass), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.AdminState != "" {
		_, err := zkdao.createOrSet(key + "/admin_state", []byte(process.AdminState), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.OperState != "" {
		_, err := zkdao.createOrSet(key + "/oper_state", []byte(process.OperState), 0, zk.WorldACL(zk.PermAll))
		if err != nil { return err }
	}
	if process.Pid != -1 {
		_, err := zkdao.createOrSet(key + "/pid", []byte(strconv.FormatInt(int64(process.Pid), 10)), 0, zk.WorldACL(zk.PermAll))
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
