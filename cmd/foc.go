package cmd

import (
	"github.com/openfaas/faasd/internal/multilru"
	"log"
	"net/http"
)

const (
	UseFoCCache              = false
	FoCCacheSize             = 1000*1024*1024
)

var focCache *multilru.Cache

func checkFoCCache(hashReq string, r *http.Request) (*http.Response, error) {
	response, found := focCache.Get(hashReq)
	if found {
		res, err := unserializeReq(response.([]byte), r)
		if err != nil {
			log.Println("Mohammad unserialize res: ", err.Error())
			return nil, err
		}
		return res, nil
	}
	return nil, nil
}
