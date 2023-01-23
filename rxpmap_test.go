package rxpmap_test

import (
	"context"
	"testing"
	"time"

	"github.com/Instantan/rxpmap"
	"github.com/stretchr/testify/assert"
)

func TestRxpmap(t *testing.T) {

	m, err := rxpmap.NewMemory()
	assert.NoError(t, err)

	v1Input := []byte("Value1")
	v2Input := []byte("Value2")

	assert.NoError(t, m.Write("Test1", v1Input))
	assert.NoError(t, m.Write("Test2", v2Input))

	v1, ok := m.Get("Test1")
	assert.True(t, ok)
	assert.Equal(t, v1Input, v1)

	v2, ok := m.Get("Test2")
	assert.True(t, ok)
	assert.Equal(t, v2Input, v2)

	t.Logf("Map: %v", m)
}

func TestRxpmapListen(t *testing.T) {

	m, err := rxpmap.NewMemory()
	assert.NoError(t, err)

	receiver := make(chan map[string][]byte)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		m.Listen(ctx, "data", receiver)
		close(receiver)
	}()

	go func() {
		for _, v := range []string{"v1", "v2", "v3"} {
			time.Sleep(time.Millisecond * 500)
			m.Write("data", []byte(v))
		}
		time.Sleep(time.Millisecond)
		cancel()
	}()

	for v := range receiver {
		t.Logf("Received: %v", v)
	}
}
