package drwmutex

import "runtime"

func map_cpus() (cpus map[uint64]int) {
	cpus = make(map[uint64]int)
	numCPU := runtime.NumCPU()
	i := 0

	for {
		c := cpu()
		if _, ok := cpus[c]; !ok {
			cpus[c] = i
			i++
			if i == numCPU {
				break
			}
		}
	}

	return
}
