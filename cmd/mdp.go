package cmd

import (
	"fmt"
	"github.com/openfaas/faasd/internal/lru"
	"github.com/openfaas/faasd/internal/mdp"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	MUFoCCacheSize  = 1 * 1024 * 1024
	MUTAHCCacheSize = 1 * 1024 * 1024
	UseMDPCache     = true
)

var functionsMDP = make(map[string]*mdp.MarkovDecisionProcess)
var multiLRU *lru.MultiCache
var mdpLock sync.Mutex
var nFoC uint
var nTAHC uint
var nNoCache uint

//ToDo: complete cache resize

func MDPStateUpdate(lastState int, currentState int) (uint, uint, uint) {
	mdpLock.Lock()
	defer mdpLock.Unlock()
	switch lastState {
	case 0:
		nFoC = nFoC - 1
		break
	case 1:
		nTAHC = nTAHC - 1
		break
	case 2:
		nNoCache = nNoCache - 1
		break
	}

	switch currentState {
	case 0:
		nFoC = nFoC + 1
		break
	case 1:
		nTAHC = nTAHC + 1
		break
	case 2:
		nNoCache = nNoCache + 1
		break
	}

	fmt.Printf("MDPStateUpdate nFoC: %v, nTAHC: %v, nNoCache: %v \n", nFoC, nTAHC, nNoCache)
	return nFoC, nTAHC, nNoCache
}

func mdpLoadBalance(RequestURI string, sReqHash string, exteraPath string, r *http.Request) ([]byte, error) {
	var agentId uint32
	mdpLock.Lock()
	markovDecisionProcess, ok := functionsMDP[RequestURI]
	if !ok {
		fmt.Printf("mdpLoadBalance new function name: %v \n", RequestURI)
		nFoC = nFoC + 1
		markovDecisionProcess = mdp.New([]string{"FoC", "TAHC", "NoCache"},
			0, [][]float32{
				{0.5, 0.5, 0},
				{0.3333, 0.3333, 0.3333},
				{0, 0.5, 0.5},
			}, MDPStateUpdate, nFoC, nTAHC, nNoCache, RequestURI)
		functionsMDP[RequestURI] = markovDecisionProcess
	}
	mdpLock.Unlock()

	markovDecisionProcess.AddFunctionInput(sReqHash)
	state := markovDecisionProcess.CurrentState()
	fmt.Printf("mdpLoadBalance function name: %v, state: %v \n", RequestURI, state)
	if state == 0 { // FoC
		value, found := multiLRU.Get(lru.FoCCache, sReqHash)
		if found {
			fmt.Printf("mdpLoadBalance FOC is found function name: %v \n", RequestURI)
			return value.([]byte), nil
		}
	} else if state == 1 {
		value, found := multiLRU.Get(lru.TAHCCache, sReqHash)
		if found {
			t1 := time.Now()
			agentId = value.(uint32)
			if workerCluster.CheckAgentLoad(int(agentId)) {
				sReq, err := captureRequestData(r)
				if err != nil {
					return nil, err
				}
				mutexAgent.Lock()
				cacheHit++
				mutexAgent.Unlock()
				duration := time.Since(t1)
				log.Printf("UseMDPCache sendToAgent due to Cache cacheHit: %v, RequestURI :%s, duration: %v  \n",
					cacheHit, RequestURI, duration.Microseconds())
				res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
				return res.Response, nil
			}
			atomic.AddUint64(&loadMiss, 1)
		}
	}
	sReq, err := captureRequestData(r)
	if err != nil {
		return nil, err
	}
	agentId = uint32(workerCluster.SelectAgent())
	res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
	if err != nil {
		return nil, err
	}
	if state == 0 {
		multiLRU.Add(lru.FoCCache, sReqHash, res.Response)
	} else if state == 1 {
		multiLRU.Add(lru.TAHCCache, sReqHash, agentId)
	}
	return res.Response, nil
}
