package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os/exec"
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
		[]string{"namespace", "pod_name", "container_name"},
	)
)

// Params represents the parameters required of RunJob.
type Params struct {
	Path  *string
	UseWg bool
	Wg    *sync.WaitGroup
}

// Metric represents the structure of a metric measured.
type Metric struct {
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
	RowCount      int64  `json:"row_count"`
}

// Output represents the bash script execution output.
type Output struct {
	Metrics []Metric `json:""`
}

// init
func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(logRowCountPrometheus)
	log.Println("Registered internal metrics")
}

// main
func main() {
	addr := flag.String("web.listen-address", ":9300", "Address on which to expose metrics")
	interval := flag.Int("interval", 10, "Interval for metrics collection in seconds")
	pathName := flag.String("pathname", "./scripts/job.sh", "pathname bash script")
	debug := flag.Bool("debug", true, "Debug log true/false")
	flag.Parse()
	// serve the HTTP request
	http.Handle("/metrics", promhttp.Handler())
	// invoke the function Run in a goroutine,
	go Run(int(*interval), *pathName, *debug)
	// start a web server listening on a specific port for Prometheus scraping
	// create a goroutine for every HTTP request and run it against a Handler.
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
Run function schedule a job every "interval" seconds.
This job run the bash script defined in the param "pathname" that collects metrics.
Metrics will be store in a memory map and in a memory structure for Prometheus scraping.
If "debug" is true, the job print log messages to stdout (default false).
If it encounters any errors, it will panic.
*/
func Run(interval int, pathName string, debug bool) {
	_, err := ioutil.ReadFile(pathName)
	if err != nil {
		log.Fatal(err)
	}

	// this is executed only if the method didn't return an error
	for {
		var wg sync.WaitGroup
		wg.Add(1)
		o := Output{}
		p := Params{UseWg: true, Wg: &wg, Path: &pathName}

		// run the job in a goroutine
		go o.RunJob(&p)
		wg.Wait()

		for _, val := range o.Metrics {

			key := val.Namespace + "-" + val.PodName + "-" + val.ContainerName
			value := val.RowCount

			if debug {
				log.Println("New metric read from "+pathName+":", key, value)
			}

			// store metric into a memory map (global variable)
			UpdateLogRowCountMetricMap(key, value, debug)

			prometheusLabels := map[string]string{}

			prometheusLabels["namespace"] = val.Namespace
			prometheusLabels["pod_name"] = val.PodName
			prometheusLabels["container_name"] = val.ContainerName

			// store metric into a memory structure for Prometheus (global variable).
			logRowCountPrometheus.With(prometheus.Labels(prometheusLabels)).Set(logRowCountMetricMap[key])

			if debug {
				log.Println("New metric saved in memory for Prometheus:", prometheusLabels, logRowCountMetricMap[key])
			}
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

/*
RunJob implements a job that run the bash script defined in the param.
The job execution output will be saved into the variable "o"
*/
func (o *Output) RunJob(p *Params) {
	if p.UseWg {
		defer p.Wg.Done()
	}
	// run the bash script
	o.RunExec(p.Path)
}

/* RunExec run the bash script defined in the param that collects metrics.
Metrics will be saved into the variable "o"
If RunExec encounters any errors, it will panic.
*/
func (o *Output) RunExec(pathname *string) {

	out, err := exec.Command(*pathname).Output()
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(out, &o.Metrics)
	if err != nil {
		log.Fatal(err)
	}
}

/*
UpdateLogRowCountMetricMap store the metric defined in the param into a memory map (global variable).
If the previos values of this metric is greater then 0, the average between the previous value
and the current one will be store into the the memory map.
*/
func UpdateLogRowCountMetricMap(key string, lastValue int64, debug bool) {
	currentValue := logRowCountMetricMap[key]
	nextValue := float64(lastValue)

	if currentValue > 0 {
		nextValue = (currentValue + float64(lastValue)) / 2.0
	}

	nextValueRoundToNearest := math.Round(nextValue*100) / 100

	logRowCountMetricMap[key] = nextValueRoundToNearest

	if debug {
		log.Println("New metric saved in memory:", key, nextValueRoundToNearest)
	}
}
