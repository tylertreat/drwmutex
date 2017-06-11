package drwmutex

import "testing"

func BenchmarkMapCPUs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = map_cpus()
	}
}
