package dogfood

const DefaultMinConfidence = 0.75

type Router struct {
	MinConfidence float64
}

func (r Router) minConfidence() float64 {
	if r.MinConfidence <= 0 || r.MinConfidence > 1 {
		return DefaultMinConfidence
	}
	return r.MinConfidence
}
