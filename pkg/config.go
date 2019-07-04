package emulator

import (
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/containers-ai/alameda/operator/datahub"
)

type EmulatorGlobal struct {
	EmulatorNamespace string `mapstructure:"emulator_namespace"`
	EmulatorListenAddress string `mapstructure:"emulator_listen_address"`
	EmulatorWebPath string `mapstructure:"emulator_web_path"`
	EmulatorLabelName string `mapstructure:"emulator_label_name"`
	EmulatorLabelValue string `mapstructure:"emulator_label_value"`
	EmulatorNodeName string `mapstructure:"emulator_node_name"`
	EmulatorNodeResourceCPUCores int64 `mapstructure:"emulator_node_resource_cpu_cores"`
	EmulatorNodeCPUUsageRange string `mapstructure:"emulator_node_resource_cpu_usage_range"`
	EmulatorNodeResourceMemoryBytes int64 `mapstructure:"emulator_node_resource_memory_bytes"`
	EmulatorNodeMemoryBytesRange string `mapstructure:"emulator_node_resource_memory_usage_range"`
	EmulatorNodeResourceNetwotkMegabitsPerSecond int64 `mapstructure:"emulator_node_netwotk_megabits_per_second"`
	EmulatorNodeMetadata string `mapstructure:"emulator_node_metadata_template"`
	EmulatorContainerMetadata string `mapstructure:"emulator_container_metadata_template"`
	EmulatorContainerIgnoreCreation bool `mapstructure:"emulator_container_ignore_creation"`
	EmulatorPrometheusScrapSeconds int64 `mapstructure:"emulator_performance_scrap_seconds"`
}

type SubContainer struct {
	ContainerPrefixName string `mapstructure:"container_prefix_name"`
	ContainerCPUCsvFilepath string `mapstructure:"container_cpu_csv_filepath"`
	ContainerMemoryCsvFilepath string `mapstructure:"container_memory_csv_filepath"`
	ContainerPulledStartHour string `mapstructure:"container_pulled_start_hour"`
	ContainerCPURandom  bool `mapstructure:"container_cpu_random"`
	ContainerMemoryRandom bool `mapstructure:"container_memory_random"`
	ContainerCPURandomRange string `mapstructure:"container_cpu_random_range"`
	ContainerMemoryRandomRange string `mapstructure:"container_memory_random_range"`
	ContainerDataStep int64 `mapstructure:"container_data_step"`
}

type Containers struct {
	ContainerCount int `mapstructure:"container_count"`
	ContainerPrefixName string `mapstructure:"container_prefix_name"`
	ContainerCPUCsvFilepath string `mapstructure:"container_cpu_csv_filepath"`
	ContainerMemoryCsvFilepath string `mapstructure:"container_memory_csv_filepath"`
	ContainerPulledStartHour string `mapstructure:"container_pulled_start_hour"`
	ContainerCPURandom  bool `mapstructure:"container_cpu_random"`
	ContainerMemoryRandom bool `mapstructure:"container_memory_random"`
	ContainerCPURandomRange string `mapstructure:"container_cpu_random_range"`
	ContainerMemoryRandomRange string `mapstructure:"container_memory_random_range"`
	ContainerDataStep int `mapstructure:"container_data_step"`
}


type Config struct {
	Containers *Containers
	Container map[string] SubContainer
	Datahub *datahub.Config `mapstructure:"datahub"`
	Global  *EmulatorGlobal `mapstructure:"global"`
	Log *log.Config      `mapstructure:"log"`
}

// NewConfig returns Config objecdt
func NewConfig() Config {
	c := Config{
	}
	c.init()
	return c
}

func (c *Config) init() {
	defaultLogConfig := log.NewDefaultConfig()
	c.Log = &defaultLogConfig
}

func (c Config) Validate() error {
	return nil
}
