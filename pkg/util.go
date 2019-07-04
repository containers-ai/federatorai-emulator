package emulator

import (
	"math/rand"
	"math"
	"time"
	"os"
	"encoding/csv"
	"fmt"
)

func GenerateRandomFloat64(minvalue *float64, maxvalue *float64, precision int) float64 {
	var res float64
	var minV float64
	if minvalue == nil {
		minV = float64(0)
	} else {
		minV = *minvalue
	}
	
	rand.Seed(time.Now().UnixNano())
	if maxvalue != nil {
		res = minV + rand.Float64() * (*maxvalue - minV)
	} else {
		res = minV + rand.Float64()
	}
	output := math.Pow(10, float64(precision))
	return float64(int(res * output + math.Copysign(0.5, res * output))) / output
}

func ReadCSV(file string) (map[string][]string, error) {
	retMap := map[string][]string{}

	csvFile, err := os.Open(file)
	if err != nil {
		return retMap, nil
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	rows, err := csvReader.ReadAll()

	for _, row := range rows {
		podName := row[0]
		retMap[podName] = make([]string, 0)
		for _, data := range row[1:] {
			retMap[podName] = append(retMap[podName], data)
		}
	}

	return retMap, nil
}

func ConvertTimeMappingDataIndex(startTime *time.Time, inputTime *time.Time, timeStep int, startHour int64, dataCount int64) int64 {
	var sTime time.Time
	var iTime time.Time
	if inputTime == nil {
		iTime = time.Now()
	} else {
		iTime = *inputTime
	}

	if startTime == nil {
		sTime = time.Now()
	} else {
		sTime = *startTime
	}
	indexStartTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d %02d:00:00", sTime.Year(), sTime.Month(), sTime.Day(), startHour), time.Local)
	indexData := ((iTime.Unix() - indexStartTime.Unix()) / int64(timeStep)) % dataCount

	return indexData
}