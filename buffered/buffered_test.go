package buffered

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestMultiCache_Get(t *testing.T) {
	mc := NewMulti[int, int]()
	val, err := mc.Get(7)
	assert.NoError(t, err)
	assert.Equal(t, 0, val)

	mc.Set(7, 7)
	val, err = mc.Get(7)
	assert.NoError(t, err)
	assert.Equal(t, 7, val)
}

func TestMultiCache_Set(t *testing.T) {
	const (
		numConsumers = 10
		numKeys      = 5
	)

	rand.Seed(time.Now().Unix())

	mc := NewMulti[int, int]()
	for i := 0; i < numKeys; i++ {
		mc.Set(i, i)
	}

	var (
		wg          sync.WaitGroup
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		hitChange   bool
	)
	defer cancel()

	wg.Add(numConsumers)
	for i := 0; i < numConsumers; i++ {
		i := i
		go func() {
			defer wg.Done()
			start := time.Now()
			for {
				select {
				case <-ctx.Done():
					t.Error("Should not have hit context cancellation")
					return
				default:
					key := rand.Int() % numKeys
					val, err := mc.Get(key)
					assert.NoError(t, err)
					if val != key {
						t.Logf("Goroutine %d got %d instead of %d after %s", i, val, key, time.Since(start))
						hitChange = true
						return
					}
				}
			}
		}()
	}

	time.Sleep(time.Second)
	key := rand.Int() % numKeys
	start := time.Now()
	mc.Set(key, -1)
	t.Logf("Spent %s setting", time.Since(start))
	wg.Wait()
	assert.True(t, hitChange, "Consumers should have seen a change dispatched after a new value was set")
}
