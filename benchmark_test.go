package rvmat

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "fixtures", "multi.rvmat"))
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

func BenchmarkFormat(b *testing.B) {
	m, err := DecodeFile(filepath.Join("testdata", "fixtures", "multi.rvmat"), nil)
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

func BenchmarkParseValidateCorpus(b *testing.B) {
	paths := benchmarkCorpusPaths(b)
	if len(paths) == 0 {
		b.Fatalf("benchmark corpus is empty")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			m, err := DecodeFile(path, nil)
			if err != nil {
				b.Fatalf("parse %s: %v", path, err)
			}

			_ = Validate(m, &ValidateOptions{
				DisableFileCheck: true,
			})
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	opts := GenerateOptions{
		BaseMaterial: BaseMaterialSteel,
		Finish:       FinishGloss,
		Condition:    ConditionWorn,
		UseTexGen:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Generate(opts); err != nil {
			b.Fatalf("generate material: %v", err)
		}
	}
}

func BenchmarkGenerateSet(b *testing.B) {
	tmp := b.TempDir()
	mustWriteBenchTexture(b, filepath.Join(tmp, "bench_co.paa"))
	mustWriteBenchTexture(b, filepath.Join(tmp, "bench_nohq.paa"))
	mustWriteBenchTexture(b, filepath.Join(tmp, "bench_as.paa"))
	mustWriteBenchTexture(b, filepath.Join(tmp, "bench_smdi.paa"))

	opts := GenerateSetOptions{
		OutputPath: filepath.Join(tmp, "bench"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GenerateSet(opts); err != nil {
			b.Fatalf("generate rvmat: %v", err)
		}
	}
}

// benchmarkCorpusPaths collects parseable local corpus files for benchmark.
func benchmarkCorpusPaths(b *testing.B) []string {
	b.Helper()

	roots := []string{filepath.Join("testdata")}
	paths := make([]string, 0, 256)

	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".rvmat" {
				return nil
			}

			paths = append(paths, path)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			b.Fatalf("scan corpus %s: %v", root, err)
		}
	}

	if len(paths) == 0 {
		paths = append(paths, filepath.Join("testdata", "fixtures", "multi.rvmat"))
	}

	return paths
}

// mustWriteBenchTexture creates a small texture placeholder file for benchmark.
func mustWriteBenchTexture(b *testing.B, path string) {
	b.Helper()

	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		b.Fatalf("write bench texture %s: %v", path, err)
	}
}
