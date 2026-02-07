package rvmat

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "multi.rvmat"))
	if err != nil {
		b.Fatalf("read: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Parse(data, nil); err != nil {
			b.Fatalf("parse: %v", err)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	m, err := DecodeFile(filepath.Join("testdata", "multi.rvmat"), nil)
	if err != nil {
		b.Fatalf("parse: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Format(m, nil); err != nil {
			b.Fatalf("format: %v", err)
		}
	}
}
