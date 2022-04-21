package mdp

type MarkovDecisionProcess struct {
	States []string
	currentState int
    actions [][]float32
}

func (mdp *MarkovDecisionProcess) NextState() int {
	mdp.currentState = sample(mdp.actions[mdp.currentState])
	return mdp.currentState
}

func (mdp *MarkovDecisionProcess) SetActionState(state int, action []float32)  {
	mdp.actions[state] = action
}

func New(states []string, currentState int, actions [][]float32) *MarkovDecisionProcess {
	return &MarkovDecisionProcess{
		States: states,
		currentState: currentState,
		actions: actions,
	}
}


