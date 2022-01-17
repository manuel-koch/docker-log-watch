package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
)

func main() {
	sigs := make(chan os.Signal, 1)
	newContainers := make(chan ContainerInfo, 1)
	done := make(chan bool, 1)
	watchingContainers := NewWatchingContainers()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println("Received ", sig, " signal")
		done <- true
	}()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// abort if docker server is not available
	ping, err := cli.Ping(ctx)
	if err != nil || len(ping.APIVersion) == 0 {
		fmt.Println("Unable to ping docker server, aborting...")
		os.Exit(1)
	}

	// check if docker server is still alive, abort when ping fails
	go func() {
		for true {
			ping, err := cli.Ping(ctx)
			if err != nil || len(ping.APIVersion) == 0 {
				fmt.Println("Unable to ping docker server, aborting...")
				done <- true
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// listen to docker events related to starting containers
	eventOptions := types.EventsOptions{}
	events, _ := cli.Events(ctx, eventOptions)
	go func() {
		for event := range events {
			//fmt.Printf("%s %s %s\n", event.Type, event.Status, event.Action)
			if event.Type == "container" && event.Action == "start" {
				container_number := 1
				if i, err := strconv.Atoi(event.Actor.Attributes["com.docker.compose.container-number"]); err == nil {
					container_number = i
				}
				newContainer := ContainerInfo{
					ID:                           event.Actor.ID,
					Name:                         event.Actor.Attributes["name"],
					DockerComposeProject:         event.Actor.Attributes["com.docker.compose.project"],
					DockerComposeService:         event.Actor.Attributes["com.docker.compose.service"],
					DockerComposeContainerNumber: container_number,
				}
				newContainers <- newContainer
			}
		}
	}()

	go func() {
		for newContainer := range newContainers {
			// local copy of newContainer struct
			container := newContainer
			bold := color.New(color.Bold)
			watchingContainers.addContainer(&container)
			bold.Printf("Following container %s...\n", container.LogPrefix)
			go func(container *ContainerInfo) {
				options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true, Tail: "0"}
				out, err := cli.ContainerLogs(ctx, container.ID, options)
				if err != nil {
					panic(err)
				}
				watchingContainers.watchOutput(container, out)
				// stopped watching, container has stopped
				bold.Printf("Stopped following container %s\n", container.LogPrefix)
				watchingContainers.removeContainer(container)
				if len(watchingContainers.containers) == 0 {
					bold.Println("No more containers to follow")
				}
			}(&container)
		}
	}()

	// get currently running containers too
	listOptions := types.ContainerListOptions{}
	containers, err := cli.ContainerList(ctx, listOptions)
	if err != nil {
		panic(err)
	}
	for i := range containers {
		container := containers[i]
		container_number := 1
		if i, err := strconv.Atoi(container.Labels["com.docker.compose.container-number"]); err == nil {
			container_number = i
		}
		newContainer := ContainerInfo{
			ID:                           container.ID,
			Name:                         container.Labels["name"],
			DockerComposeProject:         container.Labels["com.docker.compose.project"],
			DockerComposeService:         container.Labels["com.docker.compose.service"],
			DockerComposeContainerNumber: container_number,
		}
		newContainers <- newContainer
	}

	// block until we receive the "done" via channel
	<-done
	fmt.Println("exiting")
}
