## Federator-Emulator

### Emulator of Alameda Nodes/Pods

Emulator can generate multiple fake Nodes and Containers for the benchmark
testing in the Alameda environment.

### Emulator configuration 
```apple js
[global]
emulator_listen_address = ":9200"
emulator_namespace = ""
emulator_web_path = "/metrics"
emulator_label_name = "emulator"
emulator_label_value = "emulator"
emulator_node_name = "emulator"
emulator_node_resource_cpu_cores = 8
emulator_node_resource_memory_bytes = 32000000000
emulator_node_resource_cpu_usage_range = "0"
emulator_node_resource_memory_usage_range = "0"
emulator_node_metadata_template = "/etc/emulator/node.json"
emulator_container_metadata_template = "/etc/emulator/pod.json"
emulator_container_ignore_creation = false
emulator_performance_scrap_seconds = 900

[datahub]
address = "alameda-datahub.alameda.svc:50050"

[datahub."retry-interval"]
default = 3 # second

[containers]
container_count = 100
container_prefix_name = "emulator"
container_cpu_csv_filepath = "/etc/emulator/metric_cpu.csv"
container_memory_csv_filepath = "/etc/emulator/metric_memory.csv"
container_pulled_start_hour = "0"
container_cpu_random = true
container_memory_random = true
container_cpu_random_range = "0.1-0.9"
container_memory_random_range = "0.1-0.9"
container_data_step = 3600
```

**[global]**

**emulator_listen_address**: emulator service daemon listen port is 9200  
**emulator_node_resource_cpu_cores**: Set the fake node cpu cores  
**emulator_node_resource_memory_bytes**: Set the fake node memory bytes  
**emulator_performance_scrap_seconds**: Same as the Prometheus scrape 
configuration

**[datahub]**

**address**: Config the DataHub service ip and port  

**[containers]**

**container_count**: Config the container(pod) generate count  
**container_cpu_csv_filepath**: Reference the pod CPU performance metrics file  
**container_memory_csv_filepath**: Reference the pod Memory performance metrics file  
**container_data_step**: Performance data generate time seconds step

 ### How to build the docker image  
 
 `make docker-build`
 
 Output docker image name is "federatorai-agent:latest" 