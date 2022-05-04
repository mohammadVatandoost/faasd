package mdp

import "fmt"

func (mdp *MarkovDecisionProcess) UpdateStates() {
	mdp.addFunctionLock.Lock()
	defer mdp.addFunctionLock.Unlock()
	mdp.addAVGResponseTime.Lock()
	defer mdp.addAVGResponseTime.Unlock()
	if UpdateStateUnirary {
		if KeepHistoryOfWindow {
			if mdp.inputCounter == WindowSize {
				mdp.totalInputEachStep = append(mdp.totalInputEachStep, mdp.inputCounter)
				mdp.uniqueInputEachStep = append(mdp.uniqueInputEachStep, len(mdp.uniqueInputCounter))
				mdp.updateActionsProbability()
				mdp.nextState()
				mdp.inputCounter = 0
				mdp.avgResponsesCounter = 0
				mdp.sumResponsesTime = 0
				mdp.uniqueInputCounter = make(map[string]int)
			}

			if len(mdp.totalInputEachStep) > NumberOfWindow {
				mdp.totalInputEachStep = mdp.totalInputEachStep[1:]
				mdp.uniqueInputEachStep = mdp.uniqueInputEachStep[1:]
			}
		} else {
			if mdp.inputCounter == WindowSize {
				mdp.updateActionsProbability()
				mdp.nextState()
				mdp.inputCounter = 0
				mdp.avgResponsesCounter = 0
				mdp.sumResponsesTime = 0
				mdp.uniqueInputCounter = make(map[string]int)
			}
		}
	} else {
		mdp.updateActionsProbability()
		mdp.nextState()
		mdp.inputCounter = 0
		mdp.avgResponsesCounter = 0
		mdp.sumResponsesTime = 0
		mdp.uniqueInputCounter = make(map[string]int)
	}

}

func (mdp *MarkovDecisionProcess) CurrentState() int {
	mdp.addFunctionLock.Lock()
	defer mdp.addFunctionLock.Unlock()
	return mdp.currentState
}

func (mdp *MarkovDecisionProcess) GetAVGResponseTime() int64 {
	mdp.addAVGResponseTime.Lock()
	defer mdp.addAVGResponseTime.Unlock()
	return mdp.sumResponsesTime / int64(mdp.avgResponsesCounter)
}

func (mdp *MarkovDecisionProcess) SetTotalAVGResponseTime(totalAVGResponses int) {
	mdp.addAVGResponseTime.Lock()
	defer mdp.addAVGResponseTime.Unlock()
	mdp.totalAVGResponses = totalAVGResponses
}

func (mdp *MarkovDecisionProcess) nextState() {
	//mdp.addFunctionLock.Lock()
	//defer mdp.addFunctionLock.Unlock()
	nextState := sample(mdp.actions[mdp.currentState])
	if mdp.currentState != nextState {
		mdp.nFoC, mdp.nTAHC, mdp.nNoCache = mdp.notifyStateUpdate(mdp.currentState, nextState)
	}
	fmt.Printf("mdp function name: %v, nextState: %v, currentState: %v, actions: %v \n",
		mdp.FunctionName, nextState, mdp.currentState, mdp.actions)
	mdp.currentState = nextState
}
