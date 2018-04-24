package cors

import (
	"strings"
	"testing"

	"github.com/gotuxu/assert"
)

func TestWildcard(t *testing.T) {
	w := wildcard{"foo", "bar"}
	assert.True(t, w.match("foobar"))
	assert.True(t, w.match("foobazbar"))
	assert.False(t, w.match("foobaz"))
}

func TestConvert(t *testing.T) {
	s := convert([]string{"A", "b", "C"}, strings.ToLower)
	e := []string{"a", "b", "c"}
	assert.DeepEqual(t, s, e)
}

func TestParseHeaderList(t *testing.T) {
	h := parseHeaderList("header, second-header, THIRD-HEADER, Numb3r3d-H34d3r")
	e := []string{"Header", "Second-Header", "Third-Header", "Numb3r3d-H34d3r"}
	assert.DeepEqual(t, h, e)
}

func TestParseHeaderListEmpty(t *testing.T) {
	h1 := parseHeaderList("")
	assert.Empty(t, h1)

	h2 := parseHeaderList(", ")
	assert.Empty(t, h2)
}

func BenchmarkParseHeaderList(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parseHeaderList("header, second-header, THIRD-HEADER")
	}
}
