package cmd

import (
	"github.com/openfaas/faasd/internal/lru"
	"log"
	"net/http"
)

const (
	UseFoCCache  = false
	FoCCacheSize = 1000 * 1024 * 1024
)

var focCache *lru.Cache

func checkFoCCache(cache *lru.Cache, hashReq string, r *http.Request) (*http.Response, error) {
	response, found := cache.Get(hashReq)
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
