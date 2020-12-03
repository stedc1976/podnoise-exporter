package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logRowCountMetricMap = map[string]float64{}

	logRowCountPrometheus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "logrowcount",
		Help: "log row count metric",
	},
		[]string{"namespace", "pod_name", "container_log"},
	)
)

// Type Params stores parameters.
type Params struct {
	Path  *string
	UseWg bool
	Wg    *sync.WaitGroup
}

type Metric struct {
	Namespace    string `json:"namespace"`
	PodName      string `json:"pod_name"`
	ContainerLog string `json:"container_log"`
	RowCount     int64  `json:"row_count"`
}

type Output struct {
	Metrics []Metric `json:""`
	Job     string   `json:""`
}

func (o *Output) RunJob(p *Params) {
	if p.UseWg {
		defer p.Wg.Done()
	}
	o.RunExec(p.Path)
}

func (o *Output) RunExec(path *string) {

	out, err := exec.Command(*path).Output()
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(out, &o.Metrics)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(logRowCountPrometheus)
	log.Println("Registered internal metrics")
}

func main() {
	addr := flag.String("web.listen-address", ":9300", "Address on which to expose metrics")
	interval := flag.Int("interval", 10, "Interval for metrics collection in seconds")
	pathName := flag.String("pathname", "./scripts/job.sh", "pathname bash script")
	debug := flag.Bool("debug", true, "Debug log true/false")
	flag.Parse()

	http.Handle("/metrics", promhttp.Handler())
	go Run(int(*interval), *pathName, *debug)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func Run(interval int, pathName string, debug bool) {
	_, err := ioutil.ReadFile(pathName)
	if err != nil {
		log.Fatal(err)
	}

	// This is executed only if the method didn't return an error
	for {
		var wg sync.WaitGroup
		wg.Add(1)
		o := Output{}
		o.Job = strings.Split(pathName, ".")[0]
		p := Params{UseWg: true, Wg: &wg, Path: &pathName}
		go o.RunJob(&p)
		wg.Wait()

		for _, val := range o.Metrics {

			key := val.Namespace + "-" + val.PodName + "-" + val.ContainerLog
			value := val.RowCount

			if debug == true {
				log.Println("New metric read from "+pathName+":", key, value)
			}

			updateLogRowCountMetricMap(key, value, debug)

			prometheusLabels := map[string]string{}

			prometheusLabels["namespace"] = val.Namespace
			prometheusLabels["pod_name"] = val.PodName
			prometheusLabels["container_log"] = val.ContainerLog

			logRowCountPrometheus.With(prometheus.Labels(prometheusLabels)).Set(logRowCountMetricMap[key])

			if debug == true {
				log.Println("New metric saved in memory for Prometheus:", prometheusLabels, logRowCountMetricMap[key])
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func updateLogRowCountMetricMap(key string, lastValue int64, debug bool) {
	currentValue := logRowCountMetricMap[key]
	nextValue := float64(lastValue)

	if currentValue > 0 {
		nextValue = (currentValue + float64(lastValue)) / 2.0
	}

	nextValueRoundToNearest := math.Round(nextValue*100) / 100

	logRowCountMetricMap[key] = nextValueRoundToNearest

	if debug == true {
		log.Println("New metric saved in memory:", key, nextValueRoundToNearest)
	}
}

func printLogRowCountMetricMav() {
	log.Println("Print current content logRowCountMetricMap")
	keys := make([]string, 0)
	for k := range logRowCountMetricMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, logRowCountMetricMap[k])
	}
}
