package cmd

import (
	pb "github.com/openfaas/faasd/proto/agent"
	"log"
	"sync/atomic"
	"time"
)

func loadBalancer(RequestURI string, exteraPath string, sReq []byte, sReqHash string) (*pb.TaskResponse, error, time.Time) {
	var agentId uint32

	if UseTAHC {
		t1 := time.Now()
		value, found := TAHCCache.Get(sReqHash)
		if found {
			agentId = value.(uint32)
			if workerCluster.CheckAgentLoad(int(agentId)){
				mutexAgent.Lock()
				cacheHit++
				mutexAgent.Unlock()
				duration := time.Since(t1)
				log.Printf("UseTAHC sendToAgent due to Cache cacheHit: %v, RequestURI :%s, duration: %v  \n",
					cacheHit, RequestURI, duration.Microseconds())
				endTime := time.Now()
				res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
				return res, err, endTime
			}
			atomic.AddUint64(&loadMiss, 1)
		}
		duration := time.Since(t1)
		log.Printf("loadBalancer duration: %v \n", duration.Microseconds())
	}

	agentId = uint32(workerCluster.SelectAgent())
	if UseTAHC {
		TAHCCache.Add(sReqHash, agentId)
		mutexAgent.Lock()
		cacheMiss++
		mutexAgent.Unlock()
	}
	log.Printf("sendToAgent loadMiss: %v, cacheMiss: %v,  RequestURI :%s", loadMiss, cacheMiss, RequestURI)
	endTime := time.Now()
	res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
	return res, err, endTime
}
