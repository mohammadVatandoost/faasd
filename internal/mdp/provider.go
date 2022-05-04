package mdp

import (
	"sync"
)

type StateUpdate func(int, int) (uint, uint, uint)

type MarkovDecisionProcess struct {
	States              []string
	currentState        int
	actions             [][]float32
	FunctionName        string
	uniqueInputCounter  map[string]int
	inputCounter        int
	totalInputEachStep  []int
	uniqueInputEachStep []int
	notifyStateUpdate   StateUpdate
	nFoC                uint
	nTAHC               uint
	nNoCache            uint
	addFunctionLock     sync.Mutex
	addAVGResponseTime  sync.Mutex
	avgResponsesCounter int
	sumResponsesTime    int64
	totalAVGResponses   int
}

//func (mdp *MarkovDecisionProcess) SetActionState(state int, action []float32)  {
//	mdp.actions[state] = action
//}

func (mdp *MarkovDecisionProcess) AddFunctionInput(rHash string) {
	mdp.addFunctionLock.Lock()
	mdp.uniqueInputCounter[rHash] = mdp.uniqueInputCounter[rHash] + 1
	mdp.inputCounter = mdp.inputCounter + 1
	mdp.addFunctionLock.Unlock()
	if UpdateStateUnirary {
		mdp.UpdateStates()
	}

}

func (mdp *MarkovDecisionProcess) AddAVGResponseTime(avgResponseTime int64) {
	mdp.addAVGResponseTime.Lock()
	defer mdp.addAVGResponseTime.Unlock()
	mdp.avgResponsesCounter = mdp.avgResponsesCounter + 1
	mdp.sumResponsesTime = mdp.sumResponsesTime + avgResponseTime
}

func New(states []string, currentState int, actions [][]float32, fn StateUpdate, nFoC uint,
	nTAHC uint, nNoCache uint, functionName string) *MarkovDecisionProcess {
	return &MarkovDecisionProcess{
		States:              states,
		currentState:        currentState,
		actions:             actions,
		uniqueInputCounter:  make(map[string]int),
		inputCounter:        0,
		notifyStateUpdate:   fn,
		nFoC:                nFoC,
		nTAHC:               nTAHC,
		nNoCache:            nNoCache,
		FunctionName:        functionName,
		avgResponsesCounter: 0,
		sumResponsesTime:    0,
	}
}
