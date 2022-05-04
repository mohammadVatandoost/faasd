package cmd

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

type Metric struct {
	FunctionName string
	ResponseTime int64
	InputSize    int
	ResultSize   int
	CacheHit     bool
}

var metricDataChan = make(chan *Metric, 1000)

func storeMetric() {
	csvfile, err := os.Create("benchmark_typical.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvfile)
	csvwriter.Write([]string{"Function Name", "ResponseTime", "Input Size", "Result Size", "CacheHit"})
	csvwriter.Flush()
	for {
		m := <-metricDataChan
		tmp := "False"
		if m.CacheHit {
			tmp = "True"
		}
		csvwriter.Write([]string{m.FunctionName,
			strconv.Itoa(int(m.ResponseTime)),
			strconv.Itoa(m.InputSize), strconv.Itoa(m.ResultSize), tmp})
		csvwriter.Flush()
	}
	csvfile.Close()
}
