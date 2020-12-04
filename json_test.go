package main

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

func TestParseScript2(t *testing.T) {

	out := []byte(`[{"namespace": "psd2","pod_name": "3ds-card-12-7spjp","container_name": "3ds-card-4038631ff5792525eee47ff9c0cb945da44cd05f9d48d943fb010d747655b970","row_count": 222}, {"namespace": "crmu-dev","pod_name": "redis-3-t4qlh","container_name": "redis-60faea9b6ec32975946402567018784ab4783e6370c10c7934916f62a80c5157","row_count": 8}]`)

	o := Output{}

	err := json.Unmarshal(out, &o.Metrics)
	if err != nil {
		fmt.Println(err)
	}

	var rowcountfloat float64

	for _, val := range o.Metrics {
		fmt.Println("metrica: "+val.Namespace, val.PodName, val.ContainerName, val.RowCount)
		rowcountfloat = float64(val.RowCount) / 3
		fmt.Println(math.Round(rowcountfloat*100) / 100)
	}

	got := len(o.Metrics)
	want := 2

	if got != want {
		t.Errorf("Hello() = got %q, want %q", got, want)
	}
}
