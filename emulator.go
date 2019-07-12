package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"net/http"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/containers-ai/federatorai-emulator/pkg"
	"github.com/containers-ai/alameda/pkg/utils/log"
	dataRepo "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"time"
)

var (
	emulatorConf *emulator.Config
	scope *log.Scope
	containersName []string
	containerCPULabelName []string
	containerMemoryLabelName []string
	containersCPUPerformance map[string][]string
	containersMemoryPerformance map[string][]string

)

const (
	envVarPrefix = "Emulator"
)

type Exporter struct {
	// Node
	countvecNodeCPU prometheus.CounterVec
	gaugeNodeMemAvailable prometheus.Gauge
	gaugeNodeMemBuffers prometheus.Gauge
	gaugeNodeMemCached prometheus.Gauge
	gaugeNodeMemFree prometheus.Gauge
	gaugeNodeMemTotal prometheus.Gauge
	// Pods
	countvecPodCPU prometheus.CounterVec
	gaugevecPodMemory prometheus.GaugeVec
	// Container Name
	containerNames []string
	// Start time
	startTime time.Time
}

func NewExporter(namespace string) *Exporter {
	startTime := time.Now().Local()
	 // Node metrics
	countvecNodeCPU := *prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "node_cpu_seconds_total",
		Help:      "Seconds the cpus spent in each mode."},
		[]string{"cpu", "mode"})
	gaugeNodeMemAvailable := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name: "node_memory_MemAvailable_bytes",
		Help: "Memory information field MemAvailable_bytes."})
	gaugeNodeMemFree := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name: "node_memory_MemFree_bytes",
		Help: "Memory information field MemFree_bytes."})
	gaugeNodeMemTotal := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name: "node_memory_MemTotal_bytes",
		Help: "Memory information field MemTotal_bytes."})
	gaugeNodeMemCached := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name: "node_memory_Cached_bytes",
		Help: "Memory information field Cached_bytes."})
	gaugeNodeMemBuffers := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name: "node_memory_Buffers_bytes",
		Help: "Memory information field Buffers_bytes."})

	// Pod metrics
	containerCPULabelName = []string{"container_name", "cpu", "id", "image", "name", "namespace", "pod_name"}
	containerMemoryLabelName = []string{"container_name", "id", "image", "name", "namespace", "pod_name"}

	countvecPodCPU := *prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "container_cpu_usage_seconds_total",
		Help: "Cumulative cpu time consumed in seconds."},
		containerCPULabelName)
	gaugevecPodMem := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "container_memory_usage_bytes",
		Help: "Current memory usage in bytes, including all memory regardless of when it was accessed."},
		containerMemoryLabelName)

	// Generator container name
	for i := 0; i < emulatorConf.Containers.ContainerCount; i++ {
		var containerName string
		if len(emulatorConf.Containers.ContainerPrefixName) != 0 {
			containerName = emulatorConf.Containers.ContainerPrefixName
		}
		containerName = containerName + "_" + strconv.Itoa(i)
		containersName = append(containersName, containerName)
	}

	// Load container cpu performance
	containersCPUPerformance = make(map[string][]string)
	if len(emulatorConf.Containers.ContainerCPUCsvFilepath) > 0 {
		data, err := emulator.ReadCSV(emulatorConf.Containers.ContainerCPUCsvFilepath)
		if err != nil {
			scope.Errorf(fmt.Sprintf("Unable load cpu performance data %s, %v", emulatorConf.Containers.ContainerCPUCsvFilepath, err))
		} else if len(data) > 0 {
			scope.Debugf(fmt.Sprintf("Data count:", len(data)))
			scope.Debugf(fmt.Sprintf("Containers name:", containersName))
			index := 0
			for ; index < (len(containersName)); {
				for _, v := range data {
					containersCPUPerformance[containersName[index]] = v
					index ++
					if index >= (len(containersName) - 1) {
						break
					}
				}
			}
		}
	}

	// Load container memory performance
	containersMemoryPerformance = make(map[string][]string)
	if len(emulatorConf.Containers.ContainerMemoryCsvFilepath) > 0 {
		data, err := emulator.ReadCSV(emulatorConf.Containers.ContainerMemoryCsvFilepath)
		if err != nil {
			scope.Errorf(fmt.Sprintf("Unable load memory performance data %s, %v", emulatorConf.Containers.ContainerMemoryCsvFilepath, err))
		} else if len(data) > 0 {
			index := 0
			for  ; index < (len(containersName)); {
				for _, v := range data {
					containersMemoryPerformance[containersName[index]] = v
					index ++
					if index >= (len(containersName) - 1)  {
						break
					}
				}
			}
		}
	}

	return &Exporter{
		// Node performance metrics
		countvecNodeCPU: countvecNodeCPU,
		gaugeNodeMemAvailable: gaugeNodeMemAvailable,
		gaugeNodeMemBuffers: gaugeNodeMemBuffers,
		gaugeNodeMemCached: gaugeNodeMemCached,
		gaugeNodeMemFree: gaugeNodeMemFree,
		gaugeNodeMemTotal: gaugeNodeMemTotal,
		// Pod performance metrics
		countvecPodCPU: countvecPodCPU,
		gaugevecPodMemory: gaugevecPodMem,
		// Config containers name
		containerNames: containersName,
		startTime: startTime}
}

// Collect Node cpu
func (e *Exporter) collectNodeCPU(ch chan<- prometheus.Metric) {
	cpuItem := []string{"guest", "idle", "iowait", "irq", "nice", "softirq", "steal", "user"}
	for i := int64(0); i < emulatorConf.Global.EmulatorNodeResourceCPUCores; i++ {
		for _, v := range cpuItem {
			var value float64
			if v == "irq" {
				value = 0
			}
			e.countvecNodeCPU.With(prometheus.Labels{"cpu": strconv.FormatInt(i, 10), "mode": v}).Add(value)
		}
	}
	e.countvecNodeCPU.Collect(ch)
}

func (e *Exporter) describeCPU(ch chan <- *prometheus.Desc) {
	e.countvecNodeCPU.Describe(ch)
}

// Collect Node memory
func (e *Exporter) collectNodeMemory(ch chan<- prometheus.Metric) {
	totalMemory := float64(emulatorConf.Global.EmulatorNodeResourceMemoryBytes)
	minMemory := totalMemory / float64(2)
	usedMemory := emulator.GenerateRandomFloat64(&minMemory, &totalMemory, 0)
	availMemory := totalMemory - usedMemory
	cachedMemory := float64(0)
	bufferMemory := float64(0)
	e.gaugeNodeMemAvailable.Set(availMemory)
	e.gaugeNodeMemAvailable.Collect(ch)
	e.gaugeNodeMemBuffers.Set(cachedMemory)
	e.gaugeNodeMemBuffers.Collect(ch)
	e.gaugeNodeMemCached.Set(bufferMemory)
	e.gaugeNodeMemCached.Collect(ch)
	e.gaugeNodeMemTotal.Set(totalMemory)
	e.gaugeNodeMemTotal.Collect(ch)
	e.gaugeNodeMemFree.Set(availMemory)
	e.gaugeNodeMemFree.Collect(ch)
}

func (e *Exporter) describeMemory(ch chan<- *prometheus.Desc) {
	e.gaugeNodeMemAvailable.Describe(ch)
	e.gaugeNodeMemBuffers.Describe(ch)
	e.gaugeNodeMemCached.Describe(ch)
	e.gaugeNodeMemTotal.Describe(ch)
	e.gaugeNodeMemFree.Describe(ch)
}

func (e *Exporter) collectPodCPU(ch chan <- prometheus.Metric) {
	var (
		randomMin float64
		randomMax float64
	)
	if len(emulatorConf.Containers.ContainerCPURandomRange) > 0 {
		vArray := strings.Split(emulatorConf.Containers.ContainerCPURandomRange, "-")
		if len(vArray) == 1 {
			randomMax, _ = strconv.ParseFloat(vArray[0],64)
		} else if len(vArray) >= 2 {
			randomMin, _ = strconv.ParseFloat(vArray[0], 64)
			randomMax, _ = strconv.ParseFloat(vArray[1], 64)
		} else {
			randomMin = 0
			randomMax = 0
		}
	}
	for _, v := range containersName {
		value := float64(0)
		i, ok := containersCPUPerformance[v]
		if ok == false  {
			if emulatorConf.Containers.ContainerCPURandom == true {
				if randomMax == 0 && randomMin == 0 {
					value = emulator.GenerateRandomFloat64(nil, nil, 6)
				} else {
					value = emulator.GenerateRandomFloat64(&randomMin, &randomMax, 6)
				}
			}
		} else {
			tm := time.Now().Local()
			startHour, _ := strconv.ParseInt(emulatorConf.Containers.ContainerPulledStartHour, 10, 64)
			dIndex := emulator.ConvertTimeMappingDataIndex(
				&e.startTime, &tm, emulatorConf.Containers.ContainerDataStep, startHour, int64(len(i)))
			value, _ = strconv.ParseFloat(i[dIndex], 64)
			value = value * float64(emulatorConf.Global.EmulatorPrometheusScrapSeconds) / float64(emulatorConf.Containers.ContainerDataStep)
		}
		// 	containerCPULabelName = []string{"container_name", "cpu", "id", "image", "name", "namespace", "pod_name"}
		e.countvecPodCPU.With(prometheus.Labels{"container_name": v, "cpu": "total", "id": v + "_1", "image": v + "emulator", "name": v, "namespace": emulatorConf.Containers.ContainersNamespace, "pod_name": v}).Add(value)
	}
	e.countvecPodCPU.Collect(ch)
}

func (e *Exporter) describePodCPU(ch chan <- *prometheus.Desc) {
	e.countvecPodCPU.Describe(ch)
}

func (e *Exporter) collectPodMemory(ch chan <- prometheus.Metric) {
	var (
		randomMin float64
		randomMax float64
	)
	if len(emulatorConf.Containers.ContainerMemoryRandomRange) > 0 {
		vArray := strings.Split(emulatorConf.Containers.ContainerMemoryRandomRange, "-")
		if len(vArray) == 1 {
			randomMax, _ = strconv.ParseFloat(vArray[0],64)
		} else if len(vArray) >= 2 {
			randomMin, _ = strconv.ParseFloat(vArray[0], 64)
			randomMax, _ = strconv.ParseFloat(vArray[1], 64)
		} else {
			randomMin = 0
			randomMax = 0
		}
	}
	for _, v := range containersName {
		value := float64(0)
		i, ok := containersMemoryPerformance[v]
		if ok == false  {
			if emulatorConf.Containers.ContainerMemoryRandom == true {
				if randomMax == 0 && randomMin == 0 {
					value = emulator.GenerateRandomFloat64(nil, nil, 6)
				} else {
					value = emulator.GenerateRandomFloat64(&randomMin, &randomMax, 6)
				}
			}
		} else {
			tm := time.Now().Local()
			startHour, _ := strconv.ParseInt(emulatorConf.Containers.ContainerPulledStartHour, 10, 64)
			dIndex := emulator.ConvertTimeMappingDataIndex(
				&e.startTime, &tm, emulatorConf.Containers.ContainerDataStep, startHour, int64(len(i)))
			value, _ = strconv.ParseFloat(i[dIndex], 64)
		}
		// containerMemoryLabelName = []string{"container_name", "id", "image", "name", "namespace", "pod_name"}
		e.gaugevecPodMemory.With(prometheus.Labels{"container_name": v, "id": v + "_1", "image": v + "emulator", "name": v, "namespace": emulatorConf.Containers.ContainersNamespace, "pod_name": v}).Set(value)
	}
	e.gaugevecPodMemory.Collect(ch)
}

func (e *Exporter) describePodMemory(ch chan <- *prometheus.Desc) {
	e.gaugevecPodMemory.Describe(ch)
}

func (e *Exporter) createNodesMetadata() error {
	data, err := ioutil.ReadFile(emulatorConf.Global.EmulatorNodeMetadata)
	if err != nil {
		panic ("Unable to load node metadata from file")
		return status.Errorf(
			codes.Internal, "Unable to load node metadata from %s, %v",
			emulatorConf.Global.EmulatorContainerMetadata, err)
	}
	nodeMetadata := emulator.NewNodeMetadata(data)
	if nodeMetadata == nil {
		return status.Errorf(codes.Internal, "Unable to convert node metadata")
	}

	nodeMetadata.Node.Name = emulatorConf.Global.EmulatorNodeName
	nodeMetadata.Node.Capacity.CpuCores = emulatorConf.Global.EmulatorNodeResourceCPUCores
	nodeMetadata.Node.Capacity.MemoryBytes = emulatorConf.Global.EmulatorNodeResourceMemoryBytes
	nodeMetadata.Node.Capacity.NetwotkMegabitsPerSecond = emulatorConf.Global.EmulatorNodeResourceNetwotkMegabitsPerSecond

	conn, err := grpc.Dial(emulatorConf.Datahub.Address, grpc.WithInsecure())
	if err != nil {
		scope.Error(fmt.Sprintf("Failed to connect data repository %s, %v", emulatorConf.Datahub.Address, err))
		panic(fmt.Sprintf("Failed to connect data repository %s, %v", emulatorConf.Datahub.Address, err))
	}
	defer conn.Close()

	var nodes []*dataRepo.Node
	nodes = append(nodes, nodeMetadata.Node)
	dataClient := dataRepo.NewDatahubServiceClient(conn)
	req := &dataRepo.CreateAlamedaNodesRequest{
		AlamedaNodes: nodes,
	}

	resp, err := dataClient.CreateAlamedaNodes(context.Background(), req)
	if err != nil {
		scope.Errorf(fmt.Sprintf("Unable to create alameda node, %v", err))
		return err
	}
	if resp.Code != 0 {
		scope.Errorf(fmt.Sprintf("Unable to create alameda node, %s", resp.Message))
		return status.Errorf(codes.Internal, "Unable to create alameda node, %s", resp.Message)
	}
	scope.Debug(fmt.Sprintf("Create nodes status: %v", resp))
	return nil
}

func (e *Exporter) createPodsMetadata() error {
	var pods []*dataRepo.Pod
	data, err := ioutil.ReadFile(emulatorConf.Global.EmulatorContainerMetadata)
	if err != nil {
		panic ("Unable to load container metadata from file")
		return status.Errorf(
			codes.Internal, "Unable to load container metadata from %s, %v",
			emulatorConf.Global.EmulatorContainerMetadata, err)
	}

	for _, containerName := range e.containerNames {
		podMetadata := emulator.NewPodMetadata(data)
		if podMetadata == nil {
			return status.Errorf(codes.Internal, "Unable to convert container metadata")
		}
		podMetadata.SetNamesapce(emulatorConf.Containers.ContainersNamespace)
		podMetadata.SetNodeName(emulatorConf.Global.EmulatorNodeName)

		pMetadata := new(emulator.ConvPodMetadata)
		pMetadata.SetPod(podMetadata.Pod)
		for _, v := range podMetadata.Pod.Containers {
			pMetadata.SetContainer(v)
		}
		pMetadata.SetPodName(containerName)
		pMetadata.SetContainerName(emulatorConf.Containers.ContainerPrefixName)
		pMetadata.EnableHPA(true)
		pMetadata.EnableVPA(true)
		pods = append(pods, pMetadata.GetPod())
	}

	scope.Infof(fmt.Sprintf("sample pod: %v", pods[0]))
	for _, container := range pods[0].Containers {
		scope.Infof(fmt.Sprintf("container: %v", container))
	}
	conn, err := grpc.Dial(emulatorConf.Datahub.Address, grpc.WithInsecure())
	if err != nil {
		scope.Error(fmt.Sprintf("Failed to connect data repository %s, %v", emulatorConf.Datahub.Address, err))
		panic(fmt.Sprintf("Failed to connect data repository %s, %v", emulatorConf.Datahub.Address, err))
	}
	defer conn.Close()
	req := &dataRepo.CreatePodsRequest{
		Pods: pods}
	dataClient := dataRepo.NewDatahubServiceClient(conn)

	resp, err := dataClient.CreatePods(context.Background(), req)
	if err != nil {
		scope.Errorf(fmt.Sprintf("Unable to create alameda container, %v", err))
		return err
	}
	if resp.Code != 0 {
		scope.Errorf(fmt.Sprintf("Unable to create alameda container, %s", resp.Message))
		return status.Errorf(codes.Internal, "Unable to create alameda container, %s", resp.Message)
	}
	scope.Infof(fmt.Sprintf("Create container status: %v", resp))
	return nil
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric ) {
	if emulatorConf.Global.EmulatorContainerIgnoreCreation == false {
		// Create Nodes metadata
		e.createNodesMetadata()
		// Create pods metadata
		e.createPodsMetadata()
	}

	// Generate Nodes performance
	e.collectNodeCPU(ch)
	e.collectNodeMemory(ch)

	// Generate Container performance
	e.collectPodCPU(ch)
	e.collectPodMemory(ch)
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.describeCPU(ch)
	e.describeMemory(ch)
	e.describePodCPU(ch)
	e.describePodMemory(ch)
}

func loadEmulatorConfiguration(emulatorConfPath string) *emulator.Config {
	var eConf emulator.Config
	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.SetConfigFile(emulatorConfPath)
	err := viper.ReadInConfig()
	if err != nil {
		panic(errors.New("Read configuration failed: " + err.Error()))
	}
	err = viper.Unmarshal(&eConf)
	if err != nil {
		panic(errors.New("Unmarshal configuration failed: " + err.Error()))
	} else {
		if transmitterConfBin, err := json.MarshalIndent(eConf, "", "  "); err == nil {
			scope.Debugf("Display configuration")
			scope.Info(fmt.Sprintf("Transmitter configuration: %s", string(transmitterConfBin)))
		} else {
			scope.Debugf("Unable to display configuration")
		}
	}
	return &eConf
}

func init() {
	scope = log.RegisterScope("Emulator", "Emulator operator entry point", 0)
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9200").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		namespace = kingpin.Flag(
			"namespace",
			"Define container output namespace.",
		).String()
		labelName = kingpin.Flag(
			"lablename",
			"Define container output labelname.",
		).String()
		labelValue = kingpin.Flag(
			"labelvalue",
			"Define container output lablevalue.",
		).String()
		configFile = kingpin.Flag(
			"emulator_config",
			"Define emulator configuration path.",
		).Default("/etc/emulator/emulator.toml").String()
		/*
		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()
		*/
	)
	kingpin.Parse()
	emulatorConf = loadEmulatorConfiguration(*configFile)
	*namespace = emulatorConf.Global.EmulatorNamespace
	*listenAddress = emulatorConf.Global.EmulatorListenAddress
	*metricsPath = emulatorConf.Global.EmulatorWebPath
	*labelName = emulatorConf.Global.EmulatorLabelName
	*labelValue = emulatorConf.Global.EmulatorLabelValue
	scope.Infof(fmt.Sprintf("Pattern: %s, listen address: %s, namespace: %s, label name: %s, label value: %s",
		*metricsPath, *listenAddress, *namespace, *labelName, *labelValue))

	// Register dummy exporter
	exporter := NewExporter(*namespace)
	// registry := prometheus.NewRegistry()
	// registry.MustRegister(exporter.count)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Fake Exporter</title></head>
			<body>
			<h1>Fake Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	// Start emulator service
	fmt.Println(http.ListenAndServe(*listenAddress, nil))
}