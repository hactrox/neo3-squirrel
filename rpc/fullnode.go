package rpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"neo3-squirrel/config"
	"neo3-squirrel/util/color"
	"neo3-squirrel/util/counter"
	"neo3-squirrel/util/log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	// map[host]height -> map[string]int
	nodeHeight sync.Map
	reqLock    sync.RWMutex

	bestHeight counter.SafeCounter
)

type nodeInfo struct {
	url    string
	height int
}

func getNodes() map[string]int {
	mp := make(map[string]int)

	nodeHeight.Range(func(key interface{}, value interface{}) bool {
		host := key.(string)
		height := value.(int)

		mp[host] = height

		return true
	})
	return mp
}

func updateNodes(mp map[string]int) {
	for host, height := range mp {
		nodeHeight.Store(host, height)
	}
}

// GetStatus prints rpc host with its current best height.
func GetStatus() {
	for host, height := range getNodes() {
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

func refreshNodesHeight() int {
	nodeInfos := getHeights()

	bestIndex := -1
	for _, blockIndex := range nodeInfos {
		if bestIndex < blockIndex {
			bestIndex = blockIndex
		}
	}

	updateNodes(nodeInfos)
	bestHeight.Set(bestIndex)

	return bestIndex
}

// TraceBestHeight traces the best block height of listed rpc nodes.
func TraceBestHeight() {
	// TODO: mail alert
	// defer mail.AlertIfErr()

	rpcs := config.GetRPCs()
	newodeHeight := make(map[string]int)
	for _, rpc := range rpcs {
		newodeHeight[rpc] = 0
	}

	if len(newodeHeight) == 0 {
		log.Panic("fullnode rpc address must be set before use")
	}

	updateNodes(newodeHeight)

	log.Info("Checking all fullnodes...")

	refreshNodesHeight()
	GetStatus()

	if GetBestHeight() == -1 {
		log.Error(color.BRed("All fullnodes unavailable"))
		os.Exit(1)
	}

	go func() {
		for {
			bestHeight := refreshNodesHeight()
			if bestHeight == -1 {
				reqLock.Lock()
				hinted := false
				for {
					bestHeight = refreshNodesHeight()
					if bestHeight > -1 {
						reqLock.Unlock()
						break
					}

					if !hinted {
						msg := color.BYellow("All fullnodes down. All sync tasks paused and waiting for any nodes up")
						log.Warn(msg)
					}
					hinted = true

					time.Sleep(1 * time.Second)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func selectNode(minHeight int) (string, bool) {
	if minHeight < 0 {
		err := fmt.Errorf("minHeight cannot lower than 0, current value=%d", minHeight)
		log.Panic(err)
	}

	// Suppose all nodes are qualified.
	candidates := []string{}

	for url, height := range getNodes() {
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

// getHeights gets best height of all fullnodes.
func getHeights() map[string]int {
	nodes := getNodes()
	if len(nodes) == 0 {
		return nil
	}

	c := make(chan nodeInfo, len(nodes))

	for url := range nodes {
		go func(url string, c chan<- nodeInfo) {
			height, _ := getHeightFrom(url)
			c <- nodeInfo{
				url:    url,
				height: height,
			}
		}(url, c)
	}

	result := make(map[string]int)

	for range nodes {
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
		// log.Debug(err)
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
	updateNodes(map[string]int{url: -1})
}

func setRPCforTest(rpcAddr string) {
	nodes := map[string]int{rpcAddr: 0}
	updateNodes(nodes)
	refreshNodesHeight()
}
