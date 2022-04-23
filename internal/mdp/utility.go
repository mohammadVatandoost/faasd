package mdp

import "math/rand"

func sample(pdf []float32) int {
	r := rand.Float32()
	cdf := make([]float32, len(pdf))
	cdf[0] = pdf[0]
	for i := 1; i < len(pdf); i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}
	bucket := 0
	for r > cdf[bucket] {
		bucket++
	}
	return bucket
}
