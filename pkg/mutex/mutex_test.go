package mutex

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutexLock_Success(t *testing.T) {
	m := New()

	m.Lock()
	assert.True(t, m.IsLocked())
}

func TestMutexTryLock_Failed(t *testing.T) {
	m := New()

	ctx, cancel := context.WithCancel(context.Background())
	m.Lock()

	go cancel()
	res := m.TryLock(ctx)
	assert.False(t, res)
}

func TestMutextTryLock_Multiple(t *testing.T) {
	m := New()
	i := 0

	var wg sync.WaitGroup
	wg.Add(2)

	m.Lock()

	inc := func() {
		defer wg.Done()
		m.Lock()
		defer m.Unlock()

		i = i + 1
	}

	go inc()
	go inc()

	assert.Zero(t, i)

	m.Unlock()
	wg.Wait()

	assert.False(t, m.IsLocked())

	assert.Equal(t, 2, i)
}
