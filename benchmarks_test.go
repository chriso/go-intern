package intern

import (
	"fmt"
	"testing"
)

func benchmarkIntern(b *testing.B, str string) {
	repo := NewRepository()
	repo.Intern(str)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.Intern(str)
	}
}

func BenchmarkInternSmall(b *testing.B) {
	benchmarkIntern(b, "foobar")
}

func BenchmarkInternLarge(b *testing.B) {
	benchmarkIntern(b, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
}

func benchmarkLookup(b *testing.B, str string, exists bool) {
	repo := NewRepository()
	if exists {
		repo.Intern(str)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.Lookup(str)
	}
}

func BenchmarkLookupStringThatExists(b *testing.B) {
	benchmarkLookup(b, "foobar", true)
}

func BenchmarkLookupStringThatDoesntExist(b *testing.B) {
	benchmarkLookup(b, "foobar", false)
}

func benchmarkLookupID(b *testing.B, str string, exists bool) {
	repo := NewRepository()
	if exists {
		repo.Intern(str)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.LookupID(1)
	}
}

func BenchmarkLookupIDThatExists(b *testing.B) {
	benchmarkLookupID(b, "foobar", true)
}

func BenchmarkLookupIDThatDoesntExist(b *testing.B) {
	benchmarkLookupID(b, "foobar", false)
}

func BenchmarkOptimize1k(b *testing.B) {
	repo := NewRepository()
	for i := 1; i <= 1000; i++ {
		str := fmt.Sprintf("%d", i)
		repo.Intern(str)
	}
	freq := NewFrequency()
	freq.AddAll(repo)
	repo.Optimize(freq) // only the first call requires a sort of id/counts

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.Optimize(freq)
	}
}
