package persist

import (
	"sync"

	"github.com/ymz-ncnk/circbrk-go"
)

type partitionState struct {
	cb *circbrk.CircuitBreaker
	mu sync.Mutex
}
