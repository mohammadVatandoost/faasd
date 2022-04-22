package mdp

type MarkovDecisionProcess struct {
	States []string
	currentState int
    actions [][]float32
	FunctionName string
	uniqueInputCounter map[string]int
	inputCounter int
	totalInputEachStep []int
	uniqueInputEachStep []int
}

const (
	WindowSize = 10
	NumberOfWindow = 3
)

func (mdp *MarkovDecisionProcess) NextState() int {
	mdp.currentState = sample(mdp.actions[mdp.currentState])
	return mdp.currentState
}

//func (mdp *MarkovDecisionProcess) SetActionState(state int, action []float32)  {
//	mdp.actions[state] = action
//}

func (mdp *MarkovDecisionProcess) AddFunctionInput(rHash string)  {
	mdp.uniqueInputCounter[rHash] = mdp.uniqueInputCounter[rHash]+1
	mdp.inputCounter = mdp.inputCounter + 1
	if mdp.inputCounter == WindowSize {
		mdp.totalInputEachStep = append(mdp.totalInputEachStep, mdp.inputCounter)
		mdp.uniqueInputEachStep = append(mdp.uniqueInputEachStep, len(mdp.uniqueInputCounter))
	}
	mdp.inputCounter = 0
	mdp.uniqueInputCounter = make(map[string]int)
	if len(mdp.totalInputEachStep) > NumberOfWindow {
		mdp.totalInputEachStep = mdp.totalInputEachStep[1:]
		mdp.uniqueInputEachStep = mdp.uniqueInputEachStep[1:]
	}
}



func New(states []string, currentState int, actions [][]float32) *MarkovDecisionProcess {
	return &MarkovDecisionProcess{
		States: states,
		currentState: currentState,
		actions: actions,
		uniqueInputCounter: make(map[string]int),
		inputCounter: 0,
	}
}


