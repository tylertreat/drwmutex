package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonhoo/drwmutex"
)

const (
	SYNC int = iota
	DRWM     = iota
	ATMV     = iota
	END      = iota
)

type L interface {
	Lock()
	Unlock()
	RLocker() sync.Locker
}

type valueLocker struct {
	a *AtomicValue
}

func (v *valueLocker) Lock() {
	_ = v.a.v.Load()
}

func (v *valueLocker) Unlock() {
}

type AtomicValue struct {
	v  atomic.Value
	mu sync.Mutex
}

func (a *AtomicValue) Lock() {
	a.mu.Lock()
	_ = a.v.Load()
}

func (a *AtomicValue) Unlock() {
	a.v.Store(1)
	a.mu.Unlock()
}

func (a *AtomicValue) RLocker() sync.Locker {
	return &valueLocker{a}
}

func main() {
	cpuprofile := flag.Bool("cpuprofile", false, "enable CPU profiling")
	locks := flag.Uint64("i", 10000, "Number of iterations to perform")
	write := flag.Float64("p", 0.0001, "Probability of write locks")
	wwork := flag.Int("w", 1, "Amount of work for each writer")
	rwork := flag.Int("r", 100, "Amount of work for each reader")
	readers := flag.Int("n", runtime.GOMAXPROCS(0), "Total number of readers")
	checkcpu := flag.Uint64("c", 100, "Update CPU estimate every n iterations")
	flag.Parse()

	readers_per_core := *readers / runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	var mx L

	for l := 0; l < END; l++ {
		var o *os.File
		if *cpuprofile {
			if o != nil {
				pprof.StopCPUProfile()
				o.Close()
			}

			o, _ := os.Create(fmt.Sprintf("rw%d.out", l))
			pprof.StartCPUProfile(o)
		}

		switch l {
		case SYNC:
			mx = new(sync.RWMutex)
		case DRWM:
			mx = drwmutex.New()
		case ATMV:
			mx = new(AtomicValue)
		}

		start := time.Now()
		for n := 0; n < runtime.GOMAXPROCS(0); n++ {
			for r := 0; r < readers_per_core; r++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					rmx := mx.RLocker()
					r := rand.New(rand.NewSource(rand.Int63()))
					for n := uint64(0); n < *locks; n++ {
						if l != SYNC && *checkcpu != 0 && n%*checkcpu == 0 {
							rmx = mx.RLocker()
						}
						if r.Float64() < *write {
							mx.Lock()
							x := 0
							for i := 0; i < *wwork; i++ {
								x++
							}
							_ = x
							mx.Unlock()
						} else {
							rmx.Lock()
							x := 0
							for i := 0; i < *rwork; i++ {
								x++
							}
							_ = x
							rmx.Unlock()
						}
					}
				}()
			}
		}
		wg.Wait()
		end := time.Now()

		t := end.Sub(start)
		fmt.Println(fmt.Sprintf("mx%d", l+1), runtime.GOMAXPROCS(0), *readers, *locks, *write, *wwork, *rwork, *checkcpu, t.Seconds(), t)
	}
}
