package rpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"neo3-squirrel/config"
	"neo3-squirrel/log"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/counter"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	nodes      map[string]int
	bestHeight counter.SafeCounter
	nLock      sync.RWMutex
)

type nodeInfo struct {
	url    string
	height int
}

// GetStatus prints rpc host with its current best height.
func GetStatus() {
	nLock.RLock()
	defer nLock.RUnlock()

	for host, height := range nodes {
		if height < 0 {
			log.Warnf(color.BRedf("%s: Server unavailable", host))
		} else {
			log.Infof("%s: %d\n", host, height)
		}
	}
}

// GetBestHeight Get the highest best height from all connected fullnodes.
func GetBestHeight() int {
	return bestHeight.Get()
}

func refreshNodesStatus() int {
	nodeInfos := getHeights()

	nLock.Lock()

	bestIndex := -1
	for _, blockIndex := range nodeInfos {
		if bestIndex < blockIndex {
			bestIndex = blockIndex
		}
	}

	if len(nodes) == 0 && bestIndex == -1 {
		log.Error(color.BRed("All fullnodes unavailable"))
		os.Exit(1)
	}

	nodes = nodeInfos

	// TODO: If bestIndex == 0
	bestHeight.Set(bestIndex)

	nLock.Unlock()

	return bestIndex
}

// TraceBestHeight traces the best block height of listed rpc nodes.
func TraceBestHeight() {
	// TODO: mail alert
	// defer mail.AlertIfErr()

	log.Info("Checking all fullnodes...")

	refreshNodesStatus()
	GetStatus()

	go func() {
		for {
			refreshNodesStatus()

			time.Sleep(2 * time.Second)
		}
	}()
}

func pickNode(minHeight int) (string, bool) {
	if minHeight < 0 {
		err := fmt.Errorf("parameter invalid, minHeight cannot lower than 0, current value=%d", minHeight)
		log.Panic(err)
	}

	nLock.RLock()
	defer nLock.RUnlock()

	// Suppose all nodes are qualified.
	candidates := []string{}

	for url, height := range nodes {
		if height >= minHeight {
			// Increase the possibility to select local nodes.
			if strings.Contains(url, "127.0.0.1") ||
				strings.Contains(url, "localhost") {
				candidates = append(candidates, url)
			}

			candidates = append(candidates, url)
		}
	}

	l := len(candidates)
	if l == 0 {
		return "", false
	}

	return candidates[rand.Intn(l)], true
}

// getHeights gets best height of all fullnodes
// and returns best height from these nodes.
func getHeights() map[string]int {
	rpcs := config.GetRPCs()
	c := make(chan nodeInfo, len(rpcs))

	for _, url := range rpcs {
		go func(url string, c chan<- nodeInfo) {
			height, _ := getHeightFrom(url)
			c <- nodeInfo{
				url:    url,
				height: height,
			}
		}(url, c)
	}

	result := make(map[string]int)

	for range rpcs {
		s := <-c
		result[s.url] = s.height
	}

	close(c)

	return result
}

func getHeightFrom(url string) (int, error) {
	params := []interface{}{}
	args := generateRequestBody("getblockcount", params)
	resp := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.SetBody([]byte(args))
	req.SetRequestURI(url)

	respData := BlockCountRespponse{}
	err := client.Do(req, resp)
	if err != nil {
		log.Error(err)
		return -1, err
	}

	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		log.Error(err)
		return -1, err
	}

	return respData.Result - 1, nil
}

func nodeUnavailable(url string) {
	nLock.RLock()
	defer nLock.RUnlock()

	// Incase node changed(e.g., reloaded dut to config file change).
	if _, ok := nodes[url]; ok {
		nodes[url] = -1
	}
}
