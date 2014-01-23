package helpers

import (
	h "k.prv/rpimon/helpers"
	"testing"
)

func BenchmarkReadFromFileLastLines(b *testing.B) {
	for i := 0; i < b.N; i++ {
		res, err := h.ReadFromFileLastLines("files.go", 10)
		if res == "" || err != nil {
			b.Errorf("wrong resut %v", err)
		}
	}
}
