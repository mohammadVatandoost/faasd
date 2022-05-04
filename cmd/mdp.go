package cmd

import (
	"fmt"
	"github.com/openfaas/faasd/internal/lru"
	"github.com/openfaas/faasd/internal/mdp"
	"net/http"
	"sync"
	"sync/atomic"
)

const (
	HASHSIZE        = 40
	NODEIDSIZE      = 4
	MUFoCCacheSize  = 128 * 1024
	MUTAHCCacheSize = (HASHSIZE + NODEIDSIZE + 1) * 5 // 1 is for not equality
	UseMDPCache     = false
	TotalWindowSize = 50
)

var functionsMDP = make(map[string]*mdp.MarkovDecisionProcess)
var multiLRU *lru.MultiCache
var mdpLock sync.RWMutex
var functionCounter sync.Mutex
var mdpReqMetricLock sync.Mutex
var nFoC uint
var nTAHC uint
var nNoCache uint
var mdpRequestCounter uint64
var epochCounter uint

//ToDo: complete cache resize

func MDPStateUpdate(lastState int, currentState int) (uint, uint, uint) {
	functionCounter.Lock()
	defer functionCounter.Unlock()
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

func updateMDPsState() {
	mdpLock.RLock()
	defer mdpLock.RUnlock()
	epochCounter++

	totalAVGResponseTime := int64(0)
	for _, markovDecisionProcess := range functionsMDP {
		totalAVGResponseTime = totalAVGResponseTime + markovDecisionProcess.GetAVGResponseTime()
	}
	avgResponseTime := int(totalAVGResponseTime / int64(len(functionsMDP)))
	for _, markovDecisionProcess := range functionsMDP {
		markovDecisionProcess.SetTotalAVGResponseTime(avgResponseTime)
		markovDecisionProcess.UpdateStates()
	}
	fmt.Printf("******** updateMDPsState epochCounter: %v, avgResponseTime: %v *********** \n",
		epochCounter, avgResponseTime)
}

func mdpLoadBalance(RequestURI string, sReqHash string, exteraPath string, r *http.Request) ([]byte, bool, error) {
	var agentId uint32
	mdpLock.RLock()
	markovDecisionProcess, ok := functionsMDP[RequestURI]
	mdpLock.RUnlock()
	if !ok {
		fmt.Printf("mdpLoadBalance new function name: %v \n", RequestURI)
		nFoC = nFoC + 1
		markovDecisionProcess = mdp.New([]string{"FoC", "TAHC", "NoCache"},
			0, [][]float32{
				{0.5, 0.5, 0},
				{0.3333, 0.3333, 0.3333},
				{0, 0.5, 0.5},
			}, MDPStateUpdate, nFoC, nTAHC, nNoCache, RequestURI)
		mdpLock.Lock()
		functionsMDP[RequestURI] = markovDecisionProcess
		mdpLock.Unlock()
	}
	if !mdp.UpdateStateUnirary {
		mdpReqMetricLock.Lock()
		mdpRequestCounter = mdpRequestCounter + 1
		if mdpRequestCounter%TotalWindowSize == 0 {
			go updateMDPsState()
		}
		mdpReqMetricLock.Unlock()
	}

	markovDecisionProcess.AddFunctionInput(sReqHash)
	state := markovDecisionProcess.CurrentState()
	//fmt.Printf("mdpLoadBalance function name: %v, state: %v \n", RequestURI, state)
	if state == 0 { // FoC
		value, found := multiLRU.Get(lru.FoCCache, sReqHash)
		if found {
			//fmt.Printf("mdpLoadBalance FOC is found function name: %v \n", RequestURI)
			return value.([]byte), true, nil
		}
	} else if state == 1 {
		value, found := multiLRU.Get(lru.TAHCCache, sReqHash)
		if found {
			//t1 := time.Now()
			agentId = value.(uint32)
			if workerCluster.CheckAgentLoad(int(agentId)) {
				sReq, err := captureRequestData(r)
				if err != nil {
					return nil, false, err
				}
				mutexAgent.Lock()
				cacheHit++
				mutexAgent.Unlock()
				//duration := time.Since(t1)
				//log.Printf("UseMDPCache sendToAgent due to Cache cacheHit: %v, RequestURI :%s, duration: %v  \n",
				//	cacheHit, RequestURI, duration.Microseconds())
				res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
				return res.Response, true, nil
			}
			atomic.AddUint64(&loadMiss, 1)
		}
	}
	sReq, err := captureRequestData(r)
	if err != nil {
		return nil, false, err
	}
	agentId = uint32(workerCluster.SelectAgent())
	res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
	if err != nil {
		return nil, false, err
	}
	if state == 0 {
		multiLRU.AddByteArray(lru.FoCCache, sReqHash, res.Response)
	} else if state == 1 {
		multiLRU.AddUint32(lru.TAHCCache, sReqHash, agentId)
	}
	return res.Response, false, nil
}
