package emulator

import (
	"testing"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomFloat64(t *testing.T) {
	min := float64(10)
	max := float64(100)
	value := GenerateRandomFloat64(&min, &max, 0)
	if value < 10 && value > 100 {
		t.Failed()
	}
	fmt.Printf("value(min: %f, max: %f): %.2f ", min, max, value)
}

func TestReadCSV(t *testing.T) {
	csvPath := "../etc/metric_cpu.csv"
	data, err := ReadCSV(csvPath)
	if err != nil {
		fmt.Printf("Unable to read %s", csvPath)
		t.Failed()
	}
	for i, v := range data {
		fmt.Println("pod:", i)
		fmt.Println("data:", len(v))

	}
	fmt.Println(len(data))
}

func TestConvertTimeMappingDataIndex(t *testing.T) {
	startTime, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-07-02 00:00:00", time.Local)
	inputTime, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-07-08 23:01:00", time.Local)
	index := ConvertTimeMappingDataIndex(&startTime, &inputTime, 3600, 0, 168)
	fmt.Println("data index:", index)
	assert.EqualValues(t, index, 167, "The index value should be 167.")
}