package utils

import "testing"

// BenchmarkGenerate генерации
func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := GenerateShortKey()
		if key == "" {
			b.Fatalf("Ошибка при генерации ID")
		}
	}
}
