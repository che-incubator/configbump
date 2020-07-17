package configmaps

import "testing"

func TestKachny(t *testing.T) {
	m := make(map[string][]string)

	v := m["a"]
	if v == nil {
		v = make([]string, 1)
		v[0] = "b"
		m["a"] = v
	}

	println(m["a"][0])
}
