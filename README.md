cs575
=====

Project for CS575, Software Design

Apache ZooKeeper
----------------
`maestro` uses Apache ZooKeeper for its shared configuration.

To run `maestro`, you must:

1. download and install ZooKeeper: https://zookeeper.apache.org
2. start an instance of the server using `zkServer` tool

Building
========

Building the Agent
------------------
1. download and install go: https://golang.org/doc/install
2. from the agent directory, run `export GOPATH=$PWD`
3. run `go get github.com/samuel/go-zookeeper` to acquire a required library
4. compile the agent: `go build github.com/jetblack87/maestro/agent`

Building config loader
----------------------
1. follow steps 1-3 above for building agent
2. compile the loader: `go build github.com/jetblack87/maestro/zkload`

Building the server
-------------------
1. follow steps 1-3 above for building agent
2. compile the server: `go build github.com/jetblack87/maestro/server`

Running
=======

Loading config
--------------
Before you can run `maestro`, you must first load a configuration into ZooKeeper.

The configuration defines the following things:
1. the 'domain' - a collection of agent and process configurations
2. the 'agents' configurations
3. the 'processes' that will be started/stopped/monitored by the agents

The file `maestro_data.json` is a sample configuration.

1. create a configuration file or use `maestro_data.json`
2. run the loader, pointing to the config file with the `-file` argument: `zkload -file maestro_data.json`

If ZooKeeper is running on a host other than localhost:2181, you must supply the `-zookeeper` argument to point to that host:
`zkload -file maestro_data.json -zookeeper host1:2181,host2:2181`

If you want to dump out the current configuration from ZooKeeper, run with the `-dump` flag:
`zkload -dump`


**NOTE:** to see the full usage, run `zkload -help`

Running the Agent
-----------------
Once your configuration has been loaded, you can now start the agent process.

The agent executable has two required arguments:
1. `-domain` specifies the name of the domain in which this agent is configured
2. `-agent` specifies the name of this agent

The following command will start agent for the domain 'd01' and the agent name 'a01' (matching the configuration in maestro_data.json):
`agent -name a01 -domain d01`

Like the loader, agent assumes that the ZooKeeper server is running on localhost:2181. If you want to run against one or more remote hosts, specify them using the `-zookeeper` argument.

**NOTE:** to see the full usage, run `agent -help`

Running the Server
------------------
Once your configuration has been loaded, you can now start the server process.

The server can be used to query and update the domain configuration.

Running the `server` executable will start a webserver listening on port 8080 by default. If a custom port is needed, the `-port` option can be supplied:
`server -port 9090`

The server also accepts the `-zookeeper` argument to point to an alternate ZooKeeper server.

### GET requests

To query for all of the domains in the configuration, perform a GET request on the following URL:
`http://<host>:<port>/domains`

To query for a specific domain, perform a GET request on the following URL:
`http://<host>:<port>/domains/<domain_key>`

Where `<domain_key>` is the key of the domain for which you are querrying.

Similarly, you can query for a specific process by performing a GET request on the following URL:
`http://<host>:<port>/processes/<process_key>`

Where `<process_key>` is the key of the process for which you are querrying.


### PATCH requests

To update a process, perform a PATCH request against the URL:
`http://<host>:<port>/processes/<process_key>`

With a body containing the process fields that you wish to update:
`{"AdminState":"on"}`

The above request will update the process to change the "AdminState" of the process to "on".
