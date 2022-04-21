package cmd

import (
	"bytes"
	"fmt"
	"github.com/openfaas/faasd/internal/cluster"
	"github.com/openfaas/faasd/internal/lru"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/httputil"
	"github.com/openfaas/faas-provider/types"
	pb "github.com/openfaas/faasd/proto/agent"
)

type BaseURLResolver interface {
	Resolve(functionName string) (url.URL, error)
}

const (
	//address     = "localhost:50051"
	//defaultName = "world"
	defaultContentType    = "text/plain"
	MaxCacheItem          = 5
	MaxAgentFunctionCache = 5

	UseLoadBalancerCache  = false
	batchTime             = 50
	FileCaching           = false
	BatchChecking         = false
	StoreMetric           = true

)

type Agent struct {
	Id      uint
	Address string
	Loads   uint
}

type CacheCheckingReq struct {
	sReqHash   string
	resultChan chan pb.TaskResponse
	agentID    uint32
}


//var Cache *cache.Cache
//var Cache *lru.Cache
//var CacheAgent *lru.Cache

//var mutex sync.Mutex
var mutexAgent sync.Mutex
var cacheHit uint
//var resultCacheHit uint
//var batchCacheHit uint
var cacheMiss uint
var loadMiss uint64
var totalTime int64
var hashRequests = make(chan CacheCheckingReq, 100)

var workerCluster *cluster.Cluster


// var hashRequestsResult = make(chan CacheChecking, 100)

func initHandler() {
	log.Printf("UseFoCCache: %v, FunctionCachingSize: %v, UseLoadBalancerCache: %v, FileCaching: %v, BatchChecking: %v, batchTime: %v, UseTAHC: %v, TAHCCacheSize: %v, FoCCacheSize: %v",
		UseFoCCache, MaxCacheItem, UseLoadBalancerCache, FileCaching, BatchChecking, batchTime, UseTAHC, TAHCCacheSize, FoCCacheSize)

	cacheHit = 0
	cacheMiss = 0
	loadMiss = 0
	//batchCacheHit = 0
	//resultCacheHit = 0


	focCache  = lru.New(FoCCacheSize)
	TAHCCache = lru.New(TAHCCacheSize)

	IPAddress := "192.168.2.9"
	//localAddress := "127.0.0.1"
	workerCluster = cluster.NewCluster()
	workerCluster.AddAgent(cluster.Agent{Id: 0, Address: IPAddress+":50061", Loads: 0})
	workerCluster.AddAgent(cluster.Agent{Id: 1, Address: IPAddress+":50061", Loads: 0})
	workerCluster.AddAgent(cluster.Agent{Id: 3, Address: IPAddress+":50061", Loads: 0})
	workerCluster.AddAgent(cluster.Agent{Id: 4, Address: IPAddress+":50061", Loads: 0})
	//workerCluster.AddAgent(cluster.Agent{Id: 0, Address: IPAddress+":50061", Loads: 0})
	//workerCluster.AddAgent(cluster.Agent{Id: 0, Address: IPAddress+":50061", Loads: 0})
	//workerCluster.AddAgent(cluster.Agent{Id: 0, Address: IPAddress+":50061", Loads: 0})
	//workerCluster.AddAgent(cluster.Agent{Id: 0, Address: IPAddress+":50061", Loads: 0})

	//if BatchChecking {
	//	go checkAllNodesCache()
	//}
	//
	//if StoreMetric {
	//	go storeMetric()
	//}
}

func NewHandlerFunc(config types.FaaSConfig, resolver BaseURLResolver) http.HandlerFunc {
	log.Println("Mohammad NewHandlerFunc")
	if resolver == nil {
		panic("NewHandlerFunc: empty proxy handler resolver, cannot be nil")
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:

			initialTime := time.Now()

			pathVars := mux.Vars(r)
			functionName := pathVars["name"]
			if functionName == "" {
				httputil.Errorf(w, http.StatusBadRequest, "missing function name")
				return
			}

			exteraPath := pathVars["params"]

			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("read request bodey error :", err.Error())
			}
			log.Println("Mohammad RequestURI: ", r.RequestURI, ", inputs:", string(bodyBytes))
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			//********* check in batch caching
			var checkInNodes string
			checkInNodes = hash(append([]byte(functionName), bodyBytes...))
			if BatchChecking {
				if FileCaching {
					checkInNodes = string(bodyBytes)
				}
			}

			//*********** cache  ******************
			if UseFoCCache {
				res, err := checkFoCCache(checkInNodes, r)
				if err != nil {
					log.Println("Mohammad unserialize res: ", err.Error())
					httputil.Errorf(w, http.StatusInternalServerError, "Can't unserialize res: %s.", functionName)
					return
				}
				if res != nil {
					clientHeader := w.Header()
					copyHeaders(clientHeader, &res.Header)
					w.Header().Set("Content-Type", getContentType(r.Header, res.Header))
					w.WriteHeader(res.StatusCode)
					io.Copy(w, res.Body)
					return
				}
			}

			sReq, err := captureRequestData(r)
			if err != nil {
				httputil.Errorf(w, http.StatusInternalServerError, "Can't captureRequestData for: %s.", functionName)
				return
			}

			//proxy.ProxyRequest(w, r, proxyClient, resolver)
			// mctx := opentracing.ContextWithSpan(context.Background(), span)
			agentRes, err, decesionTime := loadBalancer(functionName, exteraPath, sReq, checkInNodes)
			if err != nil {
				httputil.Errorf(w, http.StatusInternalServerError, "Can't reach service for: %s.", functionName)
				return
			}
			resTime := time.Since(initialTime).Microseconds()
			atomic.AddInt64(&totalTime, resTime)
			fmt.Printf("Function Result acheived, RequestURI: %v, decesionTime: %v us, totalTime: %v  \n",
				functionName, decesionTime.Sub(initialTime).Microseconds(), totalTime)

			if UseFoCCache {
				focCache.Add(checkInNodes, agentRes.Response)
			}
			//resSize := len(agentRes.Response)
			res, err := unserializeReq(agentRes.Response, r)
			if err != nil {
				log.Println("Mohammad unserialize res: ", err.Error())
				httputil.Errorf(w, http.StatusInternalServerError, "Can't unserialize res: %s.", functionName)
				return
			}

			clientHeader := w.Header()
			copyHeaders(clientHeader, &res.Header)
			w.Header().Set("Content-Type", getContentType(r.Header, res.Header))

			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}



