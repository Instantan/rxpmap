package rxpmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKey(t *testing.T) {
	raw := "test.place.bla.test"
	expected := []string{
		"test",
		"place",
		"bla",
		"test",
	}
	parsed := parseKey(raw)
	assert.Equal(t, expected, parsed)
	t.Logf("Parsed Key: %v", expected)
}

func BenchmarkParseKey(b *testing.B) {
	parseKey("test.place.bla.test")
}

func TestParseQuery(t *testing.T) {
	raw := "test.place.[ mywildcar?, mywildcar2 ].test"
	expected := [][]string{
		{"test"},
		{"place"},
		{"mywildcar?", "mywildcar2"},
		{"test"},
	}
	parsed, err := parseQuery(raw)
	assert.NoError(t, err)
	assert.Equal(t, expected, parsed.data)
	t.Logf("Parsed Query: %v", expected)
}

func TestMatches(t *testing.T) {
	q, _ := parseQuery("test.place.[ mywildcar?, mywildcar2 ].test")

	assert.False(t, q.Matches(parseKey("test.place.no.test")))
	assert.True(t, q.Matches(parseKey("test.place.mywildcart.test")))
}
