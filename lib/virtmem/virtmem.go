package virtmem

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

type VirtmemPlugin struct {
	Prefix string
}

type memMetrics struct {
	commitLimit float64
	commitAs    float64
}

func (v VirtmemPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(v.MetricKeyPrefix())
	return map[string]mp.Graphs{
		"": {
			Label: labelPrefix,
			Unit:  mp.UnitBytes,
			Metrics: []mp.Metrics{
				{Name: "CommittedAS", Label: "Committed_AS", Diff: false, Scale: 1024},
				{Name: "CommitLimit", Label: "CommitLimit", Diff: false, Scale: 1024},
			},
		},
	}
}

func (v VirtmemPlugin) MetricKeyPrefix() string {
	if v.Prefix == "" {
		v.Prefix = "virtmem"
	}
	return v.Prefix
}

func (mm *memMetrics) getAndParseVirtMemMetrics() error {
	fp, err := os.Open("/proc/meminfo")
	if err != nil {
		return err
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		t := scanner.Text()
		switch strings.Fields(t)[0] {
		case "CommitLimit:":
			f, err := strconv.ParseFloat(strings.Fields(t)[1], 64)
			if err != nil {
				return err
			}
			mm.commitLimit = f
		case "Committed_AS:":
			f, err := strconv.ParseFloat(strings.Fields(t)[1], 64)
			if err != nil {
				return err
			}
			mm.commitAs = f
		}
	}
	return nil
}

func (v VirtmemPlugin) FetchMetrics() (map[string]float64, error) {
	mm := &memMetrics{}
	m := make(map[string]float64)
	err := mm.getAndParseVirtMemMetrics()
	if err != nil {
		return m, err
	}
	m["CommitLimit"] = mm.commitLimit
	m["CommittedAS"] = mm.commitAs
	return m, nil
}

func Do() {
	optPrefix := flag.String("metric-key-prefix", "VirtualMemory", "Metric key prefix")
	flag.Parse()
	v := VirtmemPlugin{
		Prefix: *optPrefix,
	}
	plugin := mp.NewMackerelPlugin(v)
	plugin.Run()
}
