package mdp

type MarkovDecisionProcess struct {
	States []string
	currentState int
    actions [][]float32
}

func (mdp *MarkovDecisionProcess) NextState()  {
	return
}

func New(states []string, currentState int, actions [][]float32) *MarkovDecisionProcess {
	return &MarkovDecisionProcess{
		States: states,
		currentState: currentState,
		actions: actions,
	}
}


