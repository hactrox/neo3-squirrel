package rpc

import (
	"encoding/json"
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
	// map[host(string)]height(int)
	nodeHeights  sync.Map
	allNodesDown bool

	bestHeight counter.SafeCounter
)

type nodeInfo struct {
	url    string
	height int
}

func getNodes() map[string]int {
	mp := make(map[string]int)

	nodeHeights.Range(func(key interface{}, value interface{}) bool {
		host := key.(string)
		height := value.(int)

		mp[host] = height

		return true
	})
	return mp
}

func updateNodes(mp map[string]int) {
	for host, newHeight := range mp {
		oldHeight, ok := nodeHeights.Load(host)
		if ok {
			if oldHeight == -1 && newHeight > 0 {
				log.Info(color.BGreenf("Upstream fullnode %s now available(height=%d)", host, newHeight))
			}

			if oldHeight.(int) > -1 && newHeight == -1 {
				log.Info(color.BPurplef("Upstream fullnode %s unavailable", host))
			}
		}

		nodeHeights.Store(host, newHeight)
	}
}

// GetStatus prints rpc host with its current best height.
func GetStatus() {
	for host, height := range getNodes() {
		if height < 0 {
			log.Warnf(color.Redf("%s: failed to get block height, rpc unavailable", host))
		} else {
			log.Infof("%s: %d\n", host, height)
		}
	}
}

// AllFullnodesDown tells if all fullnode down.
func AllFullnodesDown() bool {
	return allNodesDown
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
	rpcs := config.GetRPCs()
	newodeHeight := make(map[string]int)
	for _, rpc := range rpcs {
		newodeHeight[rpc] = -2
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
				allNodesDown = true
				reqLock.Lock()
				hinted := false
				for {
					bestHeight = refreshNodesHeight()
					if bestHeight > -1 {
						reqLock.Unlock()
						allNodesDown = false
						log.Info(color.BGreenf("Alive fullnode detected, continue sync tasks."))
						break
					}

					if !hinted {
						log.Warn(color.BYellow("All upstream fullnodes down."))
						log.Warn(color.BYellow("All tasks paused."))
						log.Warn(color.BYellow("Waiting for any nodes wake up."))
						hinted = true
					}

					time.Sleep(1 * time.Second)
				}
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func selectNode(minHeight uint) (string, bool, int) {
	// Suppose all nodes are qualified.
	candidates := []string{}
	currBestHeight := -1

	for url, height := range getNodes() {
		if height > currBestHeight {
			currBestHeight = height
		}

		if height >= int(minHeight) {
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
		return "", false, currBestHeight
	}

	return candidates[rand.Intn(l)], true, currBestHeight
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

	respData := BlockCountResponse{}
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
