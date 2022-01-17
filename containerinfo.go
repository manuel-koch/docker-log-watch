package main

import "fmt"

type ContainerInfo struct {
	ID                           string
	Name                         string
	DockerComposeProject         string
	DockerComposeService         string
	DockerComposeContainerNumber int
	LogPrefix                    string
	LogPrefixColor               string
}

func (c *ContainerInfo) applyLogPrefix(use_container_number bool) {
	if len(c.DockerComposeService) > 0 && use_container_number && c.DockerComposeContainerNumber > 0 {
		c.LogPrefix = fmt.Sprintf("%s-%d", c.DockerComposeService, c.DockerComposeContainerNumber)
	} else if len(c.DockerComposeService) > 0 {
		c.LogPrefix = c.DockerComposeService
	} else if len(c.Name) > 0 {
		c.LogPrefix = c.Name
	} else {
		c.LogPrefix = fmt.Sprintf("%8.8s", c.ID)
	}
}
