package mdp

func (mdp *MarkovDecisionProcess) updateActionsProbability()  {
	sumTotalInput := 0
	sumUniqueInput := 0
	for i:=0; i< len(mdp.totalInputEachStep); i++ {
		sumTotalInput = sumTotalInput + mdp.totalInputEachStep[i]
		sumUniqueInput = sumUniqueInput + mdp.uniqueInputEachStep[i]
	}
	uniqueProbability := float32(sumUniqueInput)/float32(sumTotalInput)
	notUniqueProbability := 1 - uniqueProbability
	// FoC
	mdp.actions[0][0] = notUniqueProbability
	mdp.actions[0][1] = uniqueProbability
	mdp.actions[0][2] = 0
	// TAHC
	mdp.actions[2][0] = 0
	mdp.actions[2][1] = notUniqueProbability
	mdp.actions[2][2] = uniqueProbability
	// No Ca
	mdp.actions[2][0] = 0
	mdp.actions[2][1] = notUniqueProbability
	mdp.actions[2][2] = uniqueProbability
}
