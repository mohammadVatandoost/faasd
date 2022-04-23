package cmd

import (
	pb "github.com/openfaas/faasd/proto/agent"
	"log"
	"time"
)

func loadBalancer(RequestURI string, exteraPath string, sReq []byte, sReqHash string) (*pb.TaskResponse, error, time.Time) {
	var agentId uint32
	if UseTAHC {
		agentId = tahcLoadBalance(RequestURI, sReqHash)
	}

	log.Printf("sendToAgent loadMiss: %v, cacheMiss: %v,  RequestURI :%s, cacheHit: %v",
		loadMiss, cacheMiss, RequestURI, cacheHit)
	endTime := time.Now()
	res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
	return res, err, endTime
}
