package main

import (
	"flag"
	"fmt"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var zookeeper *string = flag.String("zookeeper", "localhost:2181", "The ZooKeeper connection string (defaults to 'localhost:2182').")
var agentName *string = flag.String("name", "", "REQUIRED: The name of the agent.")

func main() {
	fmt.Println("Hello, world")

	flag.Parse() // Scan the arguments list

	if *versionFlag {
		fmt.Println("Version:", APP_VERSION)
	}

	if *agentName == "" {
		panic("-agent is required")
	}
}
