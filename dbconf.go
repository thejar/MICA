package mica

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type DBConf struct {
	MinMatchLen         int
	MatchKmerSize       int
	GappedWindowSize    int
	UngappedWindowSize  int
	ExtSeqIdThreshold   int
	MatchSeqIdThreshold int
	MatchExtend         int
	MapSeedSize         int
	ExtSeedSize         int
	LowComplexity       int
	SeedLowComplexity   int
	SavePlain           bool
	ReadOnly            bool
	SaveCompressed      bool
	BlastMakeBlastDB    string
	Dmnd                string
	BlastDBSize         uint64
}
var(
	 DefaultQueryDBConf = &DBConf{
		MinMatchLen:         40,
		MatchKmerSize:       4,
		GappedWindowSize:    25,
		UngappedWindowSize:  10,
		ExtSeqIdThreshold:   60,
		MatchSeqIdThreshold: 70,
		MatchExtend:         30,
		MapSeedSize:         6,
		ExtSeedSize:         0,
		LowComplexity:       10,
		SeedLowComplexity:   6,
		SavePlain:           false,
		ReadOnly:            true,
		SaveCompressed:      false,
		BlastMakeBlastDB:    "makeblastdb",
		Dmnd:                "diamond",
		BlastDBSize:         0,
	}

	 DefaultDBConf = &DBConf{
		MinMatchLen:         40,
		MatchKmerSize:       4,
		GappedWindowSize:    25,
		UngappedWindowSize:  10,
		ExtSeqIdThreshold:   60,
		MatchSeqIdThreshold: 70,
		MatchExtend:         30,
		MapSeedSize:         6,
		ExtSeedSize:         0,
		LowComplexity:       10,
		SeedLowComplexity:   6,
		SavePlain:           false,
		ReadOnly:            true,
		SaveCompressed:      false,
		BlastMakeBlastDB:    "makeblastdb",
		Dmnd:                "diamond",
		BlastDBSize:         0,
	}
)





func (conf *DBConf) DeepCopy() *DBConf {
	copied := DBConf{

		MinMatchLen:         conf.MinMatchLen,
		MatchKmerSize:       conf.MatchKmerSize,
		GappedWindowSize:    conf.GappedWindowSize,
		UngappedWindowSize:  conf.UngappedWindowSize,
		ExtSeqIdThreshold:   conf.ExtSeqIdThreshold,
		MatchSeqIdThreshold: conf.MatchSeqIdThreshold,
		MatchExtend:         conf.MatchExtend,
		MapSeedSize:         conf.MapSeedSize,
		ExtSeedSize:         conf.ExtSeedSize,
		LowComplexity:       conf.LowComplexity,
		SeedLowComplexity:   conf.SeedLowComplexity,
		SavePlain:           conf.SavePlain,
		ReadOnly:            conf.ReadOnly,
		SaveCompressed:      conf.SaveCompressed,
		BlastMakeBlastDB:    conf.BlastMakeBlastDB,
		Dmnd:                conf.Dmnd,
		BlastDBSize:         conf.BlastDBSize,
	}
	return &copied
}

func LoadDBConf(r io.Reader) (conf *DBConf, err error) {
	defer func() {
		if perr := recover(); perr != nil {
			err = perr.(error)
		}
	}()
	conf = DefaultDBConf
	csvReader := csv.NewReader(r)
	csvReader.Comma = ':'
	csvReader.Comment = '#'
	csvReader.FieldsPerRecord = 2
	csvReader.TrailingComma = false
	csvReader.TrimLeadingSpace = true

	lines, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		atoi := func() int {
			var i64 int64
			var err error
			if i64, err = strconv.ParseInt(line[1], 10, 32); err != nil {
				panic(err)
			}
			return int(i64)
		}
		atoui := func() uint64 {
			var ui64 uint64
			var err error
			if ui64, err = strconv.ParseUint(line[1], 10, 64); err != nil {
				panic(err)
			}
			return uint64(ui64)
		}
		switch line[0] {
		case "MinMatchLen":
			conf.MinMatchLen = atoi()
		case "MatchKmerSize":
			conf.MatchKmerSize = atoi()
		case "GappedWindowSize":
			conf.GappedWindowSize = atoi()
		case "UngappedWindowSize":
			conf.UngappedWindowSize = atoi()
		case "ExtSeqIdThreshold":
			conf.ExtSeqIdThreshold = atoi()
		case "MatchSeqIdThreshold":
			conf.MatchSeqIdThreshold = atoi()
		case "MatchExtend":
			conf.MatchExtend = atoi()
		case "MapSeedSize":
			conf.MapSeedSize = atoi()
		case "ExtSeedSize":
			conf.ExtSeedSize = atoi()
		case "LowComplexity":
			conf.LowComplexity = atoi()
		case "SeedLowComplexity":
			conf.SeedLowComplexity = atoi()
		case "SavePlain":
			if strings.TrimSpace(line[1]) == "1" {
				conf.SavePlain = true
			} else {
				conf.SavePlain = false
			}
		case "ReadOnly":
			if strings.TrimSpace(line[1]) == "1" {
				conf.ReadOnly = true
			} else {
				conf.ReadOnly = false
			}
		case "SaveCompressed":
			if strings.TrimSpace(line[1]) == "1" {
				conf.SaveCompressed = true
			} else {
				conf.SaveCompressed = false
			}
		case "BlastMakeBlastDB":
			conf.BlastMakeBlastDB = strings.TrimSpace(line[1])
		case "Dmnd":
			conf.Dmnd = strings.TrimSpace(line[1])
		case "BlastDBSize":
			conf.BlastDBSize = atoui()
		default:
			return nil, fmt.Errorf("Invalid DBConf flag: %s", line[0])
		}
	}

	return conf, nil
}

func (flagConf *DBConf) FlagMerge(fileConf *DBConf) (*DBConf, error) {
	only := make(map[string]bool, 0)
	flag.Visit(func(f *flag.Flag) { only[f.Name] = true })

	if only["map-seed-size"] {
		return flagConf, fmt.Errorf("The map seed size cannot be changed for " +
			"an existing database.")
	}
	if only["read-only"] {
		return flagConf, fmt.Errorf("The read-only setting cannot be changed " +
			"for an existing database.")
	}

	if !only["min-match-len"] {
		flagConf.MinMatchLen = fileConf.MinMatchLen
	}
	if !only["match-kmer-size"] {
		flagConf.MatchKmerSize = fileConf.MatchKmerSize
	}
	if !only["gapped-window-size"] {
		flagConf.GappedWindowSize = fileConf.GappedWindowSize
	}
	if !only["ungapped-window-size"] {
		flagConf.UngappedWindowSize = fileConf.UngappedWindowSize
	}
	if !only["ext-seq-id-threshold"] {
		flagConf.ExtSeqIdThreshold = fileConf.ExtSeqIdThreshold
	}
	if !only["match-seq-id-threshold"] {
		flagConf.MatchSeqIdThreshold = fileConf.MatchSeqIdThreshold
	}
	if !only["match-extend"] {
		flagConf.MatchExtend = fileConf.MatchExtend
	}
	if !only["ext-seed-size"] {
		flagConf.ExtSeedSize = fileConf.ExtSeedSize
	}
	if !only["low-complexity"] {
		flagConf.LowComplexity = fileConf.LowComplexity
	}
	if !only["seed-low-complexity"] {
		flagConf.SeedLowComplexity = fileConf.SeedLowComplexity
	}
	if !only["plain"] {
		flagConf.SavePlain = fileConf.SavePlain
	}
	if !only["compressed"] {
		flagConf.SaveCompressed = fileConf.SaveCompressed
	}
	if !only["read-only"] {
		flagConf.ReadOnly = fileConf.ReadOnly
	}
	if !only["makeblastdb"] {
		flagConf.BlastMakeBlastDB = fileConf.BlastMakeBlastDB
	}
	if !only["diamond"] {
		flagConf.Dmnd = fileConf.Dmnd
	}
	if !only["dbsize"] {
		flagConf.BlastDBSize = fileConf.BlastDBSize
	}
	return flagConf, nil
}

func (dbConf DBConf) Write(w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ':'
	csvWriter.UseCRLF = false

	s := func(i int) string {
		return fmt.Sprintf("%d", i)
	}
	su := func(i uint64) string {
		return fmt.Sprintf("%d", i)
	}
	bs := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}
	records := [][]string{
		{"MinMatchLen", s(dbConf.MinMatchLen)},
		{"MatchKmerSize", s(dbConf.MatchKmerSize)},
		{"GappedWindowSize", s(dbConf.GappedWindowSize)},
		{"UngappedWindowSize", s(dbConf.UngappedWindowSize)},
		{"ExtSeqIdThreshold", s(dbConf.ExtSeqIdThreshold)},
		{"MatchSeqIdThreshold", s(dbConf.MatchSeqIdThreshold)},
		{"MatchExtend", s(dbConf.MatchExtend)},
		{"MapSeedSize", s(dbConf.MapSeedSize)},
		{"ExtSeedSize", s(dbConf.ExtSeedSize)},
		{"LowComplexity", s(dbConf.LowComplexity)},
		{"SeedLowComplexity", s(dbConf.SeedLowComplexity)},
		{"SavePlain", bs(dbConf.SavePlain)},
		{"SaveCompressed", bs(dbConf.SaveCompressed)},
		{"ReadOnly", bs(dbConf.ReadOnly)},
		{"BlastMakeBlastDB", dbConf.BlastMakeBlastDB},
		{"Dmnd", dbConf.Dmnd},
		{"BlastDBSize", su(dbConf.BlastDBSize)},
	}
	if err := csvWriter.WriteAll(records); err != nil {
		return err
	}
	return nil
}
