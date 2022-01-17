package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"sort"
	"sync"
)

type ColorsByName map[string]*color.Color

type WatchingContainers struct {
	mutex        sync.Mutex
	containers   []*ContainerInfo
	LogPrefixLen int
	prefixColors ColorsByName
}

func NewWatchingContainers() *WatchingContainers {
	w := &WatchingContainers{
		containers:   make([]*ContainerInfo, 0),
		prefixColors: make(ColorsByName),
	}
	w.prefixColors["blue"] = color.New(color.FgBlue)
	w.prefixColors["hiblue"] = color.New(color.FgHiBlue)
	w.prefixColors["green"] = color.New(color.FgGreen)
	w.prefixColors["higreen"] = color.New(color.FgHiGreen)
	w.prefixColors["red"] = color.New(color.FgRed)
	w.prefixColors["hired"] = color.New(color.FgHiRed)
	w.prefixColors["magenta"] = color.New(color.FgMagenta)
	w.prefixColors["himagenta"] = color.New(color.FgHiMagenta)
	w.prefixColors["yellow"] = color.New(color.FgYellow)
	w.prefixColors["hiyellow"] = color.New(color.FgHiYellow)
	w.prefixColors["cyan"] = color.New(color.FgCyan)
	w.prefixColors["hicyan"] = color.New(color.FgHiCyan)
	return w
}

func (w *WatchingContainers) addContainer(c *ContainerInfo) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	c.LogPrefixColor = w.getNextColor()
	w.containers = append(w.containers, c)
	w.updatePrefixes()
}

func (w *WatchingContainers) removeContainer(c *ContainerInfo) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	idx := -1
	for i := range w.containers {
		if w.containers[i].ID == c.ID {
			idx = i
			break
		}
	}
	if idx == len(w.containers)-1 {
		w.containers = w.containers[:idx]
	} else if idx != -1 {
		w.containers = append(w.containers[:idx], w.containers[idx+1:]...)
	}
	w.updatePrefixes()
}

func (w *WatchingContainers) updatePrefixes() {
	// assuming mutex is locked
	prefixLen := 0
	for i := range w.containers {
		c := w.containers[i]
		use_container_number := c.DockerComposeContainerNumber > 1
		for j := range w.containers {
			if j == i {
				continue
			}
			cj := w.containers[j]
			if c.DockerComposeProject == cj.DockerComposeProject && c.DockerComposeService == cj.DockerComposeService {
				use_container_number = true
			}
		}
		c.applyLogPrefix(use_container_number)
		if len(c.LogPrefix) > prefixLen {
			prefixLen = len(c.LogPrefix)
		}
	}
	w.LogPrefixLen = prefixLen
}

func (w *WatchingContainers) getNextColor() string {
	// assuming mutex is locked
	usageByColor := make(map[string]int)
	for color := range w.prefixColors {
		usageByColor[color] = 0
	}
	for i := range w.containers {
		usageByColor[w.containers[i].LogPrefixColor]++
	}
	pairs := make(StringIntPairList, len(usageByColor))
	i := 0
	for k, v := range usageByColor {
		pairs[i] = StringIntPair{k, v}
		i++
	}
	sort.Sort(pairs)
	return pairs[0].Key
}

// read container logs line by line and output with colorized prefix
func (w *WatchingContainers) watchOutput(container *ContainerInfo, out io.Reader) {
	scanner := bufio.NewScanner(NewDockerLogFilter(out))
	const maxCapacity = 1024 * 256
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	if prefixColor, prefixColorValid := w.prefixColors[container.LogPrefixColor]; prefixColorValid {
		prefixColorized := prefixColor.SprintFunc()
		for scanner.Scan() {
			prefix := fmt.Sprintf("%-*s:", w.LogPrefixLen, container.LogPrefix)
			fmt.Printf("%s %s\n", prefixColorized(prefix), scanner.Text())
		}
	} else {
		panic(fmt.Sprintf("Unknown prefix color '%s'", container.LogPrefixColor))
	}
}
