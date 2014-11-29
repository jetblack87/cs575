package data

type Domain struct {
    Name string
    Key string
    Runtime RuntimeConfig
    Config StaticConfig
}

type RuntimeConfig struct {
	Agents[] Agent
}

type StaticConfig struct {
	Agents[] Agent
	Processes[] Process
}

type Agent struct {
	Name string
	Key string
	AgentClass string
	OS string
	Eph string
	Processes[] Process
}

type Process struct {
	Name string
	Key string
	Command string
	Arguments string
	ProcessClass string
	AdminState string
	OperState string
	Pid int
}