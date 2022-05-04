package cmd

import (
	"github.com/openfaas/faasd/internal/lru"
	"log"
	"sync/atomic"
	"time"
)

const (
	UseTAHC       = false
	TAHCCacheSize = 1 * 1024 * 1024
)

var TAHCCache *lru.Cache

func tahcLoadBalance(RequestURI string, sReqHash string) (uint32, bool) {
	var agentId uint32
	t1 := time.Now()
	value, found := TAHCCache.Get(sReqHash)
	if found {
		agentId = value.(uint32)
		if workerCluster.CheckAgentLoad(int(agentId)) {
			mutexAgent.Lock()
			cacheHit++
			mutexAgent.Unlock()
			duration := time.Since(t1)
			log.Printf("UseTAHC sendToAgent due to Cache cacheHit: %v, RequestURI :%s, duration: %v  \n",
				cacheHit, RequestURI, duration.Microseconds())
			return agentId, true
		}
		atomic.AddUint64(&loadMiss, 1)
	}
	agentId = uint32(workerCluster.SelectAgent())
	TAHCCache.AddUint32(sReqHash, agentId)
	mutexAgent.Lock()
	cacheMiss++
	mutexAgent.Unlock()
	return agentId, false
}
