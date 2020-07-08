package main

import "github.com/smartping/smartping/src/g"

var consulEndpoint string = "10.93.10.66:80"

func main() {
	pipe := make(chan g.AgentInfos, 1)
	go func() {
		g.AutoDiscoveryHostAgent(consulEndpoint, pipe)
	}()
	g.GenerateHostAgentConfig(pipe, g.DefaultHandler)
}
