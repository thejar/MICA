package main

const (
	memSeqSize       = 10000
	dynamicTableSize = memSeqSize * memSeqSize
	numSeeds         = 100
)

type memory struct {
	table    []int
	ref, org []byte
	seeds    [][2]int
}

func newMemory() *memory {
	return &memory{
		table: make([]int, memSeqSize*memSeqSize),
		ref:   make([]byte, memSeqSize),
		org:   make([]byte, memSeqSize),
		seeds: make([][2]int, 0, numSeeds),
	}
}