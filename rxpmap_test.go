package rxpmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRxpmap(t *testing.T) {

	m, err := NewMemory()
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
