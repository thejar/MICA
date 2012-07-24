package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/BurntSushi/cablastp"
)

var ignoredResidues = []byte{'J', 'O', 'U'}

var (
	flagGoMaxProcs         = runtime.NumCPU()
	flagMinMatchLen        = 25
	flagMatchKmerSize      = 3
	flagGappedWindowSize   = 25
	flagUngappedWindowSize = 10
	flagSeqIdThreshold     = 50
	flagSeedSize           = 3

	flagCpuProfile = ""
	flagMemProfile = ""
	flagMemStats   = ""
)

func init() {
	log.SetFlags(0)

	flag.IntVar(&flagGoMaxProcs, "p", flagGoMaxProcs,
		"The maximum number of CPUs that can be executing simultaneously.")
	flag.IntVar(&flagMinMatchLen, "min-match-len", flagMinMatchLen,
		"The minimum size of a match.")
	flag.IntVar(&flagMatchKmerSize, "match-kmer-size", flagMatchKmerSize,
		"The size of kmer fragments to match in ungapped extension.")
	flag.IntVar(&flagGappedWindowSize, "gapped-window-size",
		flagGappedWindowSize, "The size of the gapped match window.")
	flag.IntVar(&flagUngappedWindowSize, "ungapped-window-size",
		flagUngappedWindowSize, "The size of the ungapped match window.")
	flag.IntVar(&flagSeqIdThreshold, "seq-id-threshold", flagSeqIdThreshold,
		"The sequence identity threshold of a match. (An integer in the "+
			"inclusive range from 0 to 100.)")
	flag.IntVar(&flagSeedSize, "seed-size", flagSeedSize,
		"The size of a seed.")

	flag.StringVar(&flagCpuProfile, "cpuprofile", flagCpuProfile,
		"When set, a CPU profile will be written to the file specified.")
	flag.StringVar(&flagMemProfile, "memprofile", flagMemProfile,
		"When set, a memory profile will be written to the file specified.")
	flag.StringVar(&flagMemStats, "memstats", flagMemStats,
		"When set, memory statistics will be written to the file specified.")
}

func main() {
	if flag.NArg() < 2 {
		flag.Usage()
	}

	if len(flagCpuProfile) > 0 {
		f, err := os.Create(flagCpuProfile)
		if err != nil {
			fatalf("%s\n", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	sigChan := make(chan os.Signal, 2)
	go func() {
		<-sigChan
		println("\n\nStopping CPU profile...")
		pprof.StopCPUProfile()
		os.Exit(0)
	}()
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	orgSeqId := 0
	DB, err := cablastp.NewDB(flag.Arg(0), flagSeedSize, false, true)
	if err != nil {
		fatalf("%s\n", err)
	}

	// Create the compression workers.
	wg := &sync.WaitGroup{}
	jobs := make(chan compressJob, 200)
	for i := 0; i < max(1, runtime.GOMAXPROCS(0)-1); i++ {
		wg.Add(1)
		go compressWorker(DB, jobs, wg)
	}

	start := time.Now()
	for _, arg := range flag.Args()[1:] {
		seqChan, err := cablastp.ReadOriginalSeqs(arg, ignoredResidues)
		if err != nil {
			log.Fatal(err)
		}
		for readSeq := range seqChan {
			if readSeq.Err != nil {
				log.Fatal(err)
			}

			jobs <- compressJob{
				orgSeqId: orgSeqId,
				orgSeq:   readSeq.Seq,
			}
			orgSeqId++

			if orgSeqId%1000 == 0 {
				kmers, locs := DB.CoarseDB.Seeds.Size()
				secElapsed := time.Since(start).Seconds()

				seqsPerSec := float64(orgSeqId) / float64(secElapsed)
				kmersPerSec := float64(kmers) / float64(secElapsed)
				locsPerSec := float64(locs) / float64(secElapsed)

				fmt.Printf("%d sequences compressed (%0.4f seqs/sec) "+
					":: %d kmers with %d total locations "+
					"(%0.4f kmers/sec, %0.4f locs/sec)\n",
					orgSeqId, seqsPerSec, kmers, locs, kmersPerSec, locsPerSec)

				if len(flagMemProfile) > 0 {
					writeMemProfile(fmt.Sprintf("%s.%d",
						flagMemProfile, orgSeqId))
				}
				if len(flagMemStats) > 0 {
					writeMemStats(fmt.Sprintf("%s.%d", flagMemStats, orgSeqId))
				}
			}
		}
	}
	close(jobs)
	wg.Wait()

	if len(flagMemProfile) > 0 {
		writeMemProfile(fmt.Sprintf("%s.%d",
			flagMemProfile, orgSeqId))
	}

	if err := DB.CoarseDB.Save(); err != nil {
		fatalf("Could not save coarse database: %s\n", err)
	}

	DB.Close()

	fmt.Println("")
}

func errorf(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
}

func fatalf(format string, v ...interface{}) {
	errorf(format, v...)
	os.Exit(1)
}

func writeMemProfile(name string) {
	f, err := os.Create(name)
	if err != nil {
		fatalf("%s\n", err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func init() {
	flag.Usage = usage
	flag.Parse()

	runtime.GOMAXPROCS(flagGoMaxProcs)
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [flags] "+
			"database-directory "+
			"fasta-file [fasta-file ...]\n",
		path.Base(os.Args[0]))
	flag.PrintDefaults()
	os.Exit(1)
}

func writeMemStats(name string) {
	f, err := os.Create(name)
	if err != nil {
		fatalf("%s\n", err)
	}

	kb := uint64(1024)
	mb := kb * 1024
	// gb := mb * 1024 

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	fmt.Fprintf(f,
		`Alloc: %d MB
TotalAlloc: %d MB
Sys: %d MB
Lookups: %d
Mallocs: %d
Frees: %d

HeapAlloc: %d MB
HeapSys: %d MB
HeapIdle: %d MB
HeapInuse: %d MB
HeapReleased: %d B
HeapObjects: %d

StackInuse: %d
StackSys: %d
MSpanInuse: %d
MSpanSys: %d
MCacheInuse: %d
MCacheSys: %d
BuckHashSys: %d

NextGC: %d
LastGC: %d
PauseTotalNs: %d s
PauseNs: %d
NumGC: %d
`,
		ms.Alloc/mb, ms.TotalAlloc/mb,
		ms.Sys/mb, ms.Lookups, ms.Mallocs,
		ms.Frees, ms.HeapAlloc/mb, ms.HeapSys/mb,
		ms.HeapIdle/mb,
		ms.HeapInuse/mb, ms.HeapReleased, ms.HeapObjects,
		ms.StackInuse, ms.StackSys, ms.MSpanInuse, ms.MSpanSys,
		ms.MCacheInuse, ms.MCacheSys, ms.BuckHashSys,
		ms.NextGC, ms.LastGC, ms.PauseTotalNs/1000000000,
		ms.PauseNs, ms.NumGC)

	f.Close()
}
