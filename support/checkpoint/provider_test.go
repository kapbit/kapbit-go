package checkpoint_test

import (
	"math/rand"
	"testing"
	"time"

	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func TestNewCheckpointProvider(t *testing.T) {
	provider := chptsup.NewProvider(1)
	asserterror.Equal(t, provider.Checkpoint(), chptsup.Tuple{-1, -1})
}

func TestOneChptSize(t *testing.T) {
	provider := chptsup.NewProvider(1)
	chpt, ok := provider.Add(chptsup.Tuple{0, 0})
	asserterror.Equal(t, ok, true)
	asserterror.Equal(t, chpt, chptsup.Tuple{0, 0})

	chpt, ok = provider.Add(chptsup.Tuple{1, 4})
	asserterror.Equal(t, ok, true)
	asserterror.Equal(t, chpt, chptsup.Tuple{1, 4})

	chpt, ok = provider.Add(chptsup.Tuple{3, 7})
	asserterror.Equal(t, ok, false)
	asserterror.Equal(t, chpt, chptsup.Tuple{0, 0})

	chpt, ok = provider.Add(chptsup.Tuple{2, 5})
	asserterror.Equal(t, ok, true)
	asserterror.Equal(t, chpt, chptsup.Tuple{3, 7})
}

// func TestAddBeforeChpt(t *testing.T) {
// 	provider := chptman.NewCheckpointProvider(1)
// 	provider.Add(chptman.Tuple{0, 0})
// 	provider.Add(chptman.Tuple{1, 1})
// 	provider.Add(chptman.Tuple{2, 2})
// 	asserterror.Equal(provider.StringCompact(), "", t)
// }

// go test -fuzz=FuzzCheckpointProvider -fuzztime=30s

// --- Fuzz test ---
func FuzzCheckpointProvider(f *testing.F) {
	// Seed corpus: tiny, small, typical, and stress batches
	// Edge / tiny
	f.Add(1, 1, int64(42))
	f.Add(10, 5, int64(43))

	// Typical workloads
	f.Add(50, 10, int64(44))
	f.Add(100, 20, int64(45))

	// Stress / large
	f.Add(500, 50, int64(time.Now().UnixNano()))
	f.Add(1000, 100, int64(time.Now().UnixNano()))

	f.Fuzz(func(t *testing.T, n, batchSize int, seed int64) {
		// Sanitize fuzz inputs
		if n <= 0 || n > 10_000 {
			t.Skip()
		}
		if batchSize <= 0 {
			batchSize = 1
		}
		if batchSize > n {
			batchSize = n / 2 // keep it within a sensible range
			if batchSize == 0 {
				batchSize = 1
			}
		}

		// Create provider and test data
		provider := chptsup.NewProvider(batchSize)
		testData := RandomTestData(n, seed)

		// Add all numbers in random order
		for _, num := range testData.nums {
			// provider.Add(int64(num), int64(num))
			provider.Add(chptsup.Tuple{int64(num), int64(num)})
		}

		// Assertions
		if ProviderShouldBeEmpty(provider, testData) {
			asserterror.Equal(t, provider.First(), nil)
			asserterror.Equal(t, provider.Last(), nil)
			return
		}
		asserterror.Equal(t, (provider.Checkpoint().V + provider.First().Len()),
			testData.lastIndex)
		asserterror.Equal(t, provider.First(), provider.Last())
		// asserterror.Equal(provider.First().next, nil, t)
		// asserterror.Equal(provider.Last().prev, nil, t)
	})
}

// --- Helper functions ---
func RandomTestData(n int, seed int64) TestData {
	if n <= 0 {
		panic("n < 0")
	}
	d := TestData{
		nums:      make([]int, n),
		lastIndex: int64(n - 1),
	}
	for i := range d.nums {
		d.nums[i] = i
	}
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(n, func(i, j int) {
		d.nums[i], d.nums[j] = d.nums[j], d.nums[i]
	})
	return d
}

type TestData struct {
	nums      []int
	lastIndex int64
}

func ProviderShouldBeEmpty(provider *chptsup.Provider,
	testData TestData,
) bool {
	return provider.Checkpoint() == chptsup.Tuple{
		testData.lastIndex,
		testData.lastIndex,
	}
}
