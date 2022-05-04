package cmd

import (
	"net/http"
)

func loadBalancer(RequestURI string, exteraPath string, r *http.Request, sReqHash string) ([]byte, bool, error) {
	var agentId uint32
	if UseMDPCache {
		return mdpLoadBalance(RequestURI, sReqHash, exteraPath, r)
	}
	sReq, err := captureRequestData(r)
	if err != nil {
		return nil, false, err
	}
	cacheHitted := false
	if UseTAHC {
		agentId, cacheHitted = tahcLoadBalance(RequestURI, sReqHash)
	} else {
		agentId = uint32(workerCluster.SelectAgent())
	}

	//log.Printf("sendToAgent loadMiss: %v, cacheMiss: %v,  RequestURI :%s, cacheHit: %v",
	//	loadMiss, cacheMiss, RequestURI, cacheHit)
	res, err := workerCluster.SendToAgent(int(agentId), RequestURI, exteraPath, sReq, true)
	if err != nil {
		return nil, false, err
	}
	return res.Response, cacheHitted, nil
}
