package mdp

import "fmt"

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
}

const (
	WindowSize     = 15
	NumberOfWindow = 3
)

func (mdp *MarkovDecisionProcess) CurrentState() int {
	return mdp.currentState
}

func (mdp *MarkovDecisionProcess) nextState() {
	nextState := sample(mdp.actions[mdp.currentState])
	if mdp.currentState != nextState {
		mdp.nFoC, mdp.nTAHC, mdp.nNoCache = mdp.notifyStateUpdate(mdp.currentState, nextState)
	}
	fmt.Printf("mdp function name: %v, nextState: %v, currentState: %v, actions: %v \n",
		mdp.FunctionName, nextState, mdp.currentState, mdp.actions)
	mdp.currentState = nextState
}

//func (mdp *MarkovDecisionProcess) SetActionState(state int, action []float32)  {
//	mdp.actions[state] = action
//}

func (mdp *MarkovDecisionProcess) AddFunctionInput(rHash string) {
	mdp.uniqueInputCounter[rHash] = mdp.uniqueInputCounter[rHash] + 1
	mdp.inputCounter = mdp.inputCounter + 1
	if mdp.inputCounter == WindowSize {
		mdp.totalInputEachStep = append(mdp.totalInputEachStep, mdp.inputCounter)
		mdp.uniqueInputEachStep = append(mdp.uniqueInputEachStep, len(mdp.uniqueInputCounter))
		mdp.updateActionsProbability()
		mdp.nextState()
		mdp.inputCounter = 0
		mdp.uniqueInputCounter = make(map[string]int)
	}

	if len(mdp.totalInputEachStep) > NumberOfWindow {
		mdp.totalInputEachStep = mdp.totalInputEachStep[1:]
		mdp.uniqueInputEachStep = mdp.uniqueInputEachStep[1:]
	}
}

func New(states []string, currentState int, actions [][]float32, fn StateUpdate, nFoC uint,
	nTAHC uint, nNoCache uint, functionName string) *MarkovDecisionProcess {
	return &MarkovDecisionProcess{
		States:             states,
		currentState:       currentState,
		actions:            actions,
		uniqueInputCounter: make(map[string]int),
		inputCounter:       0,
		notifyStateUpdate:  fn,
		nFoC:               nFoC,
		nTAHC:              nTAHC,
		nNoCache:           nNoCache,
		FunctionName:       functionName,
	}
}
