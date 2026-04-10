package support

import (
	"errors"

	"github.com/kapbit/kapbit-go/codes"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/ymz-ncnk/circbrk-go"
)

var ErrCantResolvePartition = errors.New("can't resolve partition")

type PartitionResolver struct {
	serviceName string
	cbs         []*circbrk.CircuitBreaker
}

func NewPartitionResolver(serviceName string, cbs []*circbrk.CircuitBreaker) (
	resolver PartitionResolver, err error,
) {
	if len(cbs) <= 0 {
		err = errors.New("partition count must be positive")
		return
	}
	resolver = PartitionResolver{serviceName: serviceName, cbs: cbs}
	return
}

func MustPartitionResolver(serviceName string, cbs []*circbrk.CircuitBreaker) (
	resolver PartitionResolver,
) {
	if len(cbs) <= 0 {
		panic("partition count must be positive")
	}
	return PartitionResolver{serviceName: serviceName, cbs: cbs}
}

func (r PartitionResolver) Resolve(wid wfl.ID) (n int, err error) {
	l := len(r.cbs)
	if l == 1 {
		if r.cbs[0].Allow() {
			return 0, nil
		}
		return 0, codes.NewCircuitBreakerOpenError(r.serviceName)
	}

	// Double Hashing
	var (
		hashValue = r.hashID(string(wid))
		start     = int(hashValue % uint32(l))
		step      = int((hashValue>>16)%uint32(l-1)) + 1
	)
	for i := range l {
		idx := (start + i*step) % l
		if r.cbs[idx].Allow() {
			return idx, nil
		}
	}
	return 0, codes.NewCircuitBreakerOpenError(r.serviceName)
}

// Use a manual inline FNV-1a to avoid the heap allocation of fnv.New32a()
func (r PartitionResolver) hashID(s string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= 16777619
	}
	return hash
}
