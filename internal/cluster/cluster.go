package cluster

import (
	"context"
	pb "github.com/openfaas/faasd/proto/agent"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"sync"
	"time"
)

const MaxClientLoad = 8

type Cluster struct {
	agents []Agent
	mutex  sync.Mutex
}

func (c *Cluster) AddAgent(agent Agent) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.agents = append(c.agents, agent)
}

func (c *Cluster) AddAgentLoad(Id int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.agents[Id].Loads++
}

func (c *Cluster) CheckAgentLoad(Id int) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.agents[Id].Loads < MaxClientLoad {
		c.agents[Id].Loads++
		return true
	}
	return false
}

func (c *Cluster) decreaseAgentLoad(Id int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.agents[Id].Loads--
}

func (c *Cluster) SelectAgent() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	size := len(c.agents)
	for i := 0; i < size; i++ {
		agentID := rand.Int31n(int32(size))
		if c.agents[agentID].Loads < MaxClientLoad {
			c.agents[agentID].Loads++
			return int(agentID)
		}
	}
	return int(rand.Int31n(int32(size)))
}

func (c *Cluster) SendToAgent(Id int, RequestURI string, exteraPath string, sReq []byte,
	cacheHit bool) (*pb.TaskResponse, error) {
	c.mutex.Lock()
	address := c.agents[Id].Address
	c.mutex.Unlock()
	defer c.decreaseAgentLoad(Id)
	//fmt.Printf("Cluster SendToAgent ID: %v, RequestURI: %v, exteraPath: %v \n", Id, RequestURI, exteraPath)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	caller := pb.NewTasksRequestClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	r, err := caller.TaskAssign(ctx, &pb.TaskRequest{FunctionName: RequestURI,
		ExteraPath: exteraPath, SerializeReq: sReq,
		TimeNanoSecond: time.Now().UnixNano(), CacheHit: cacheHit})
	if err != nil {
		log.Printf("could not TaskAssign: %v", err.Error())
		return nil, err
	}
	//fmt.Printf("Cluster SendToAgent ID: %v, RequestURI: %v, exteraPath: %v, response: %v \n",
	//	Id, RequestURI, exteraPath, string(r.Response))
	return r, nil

}
