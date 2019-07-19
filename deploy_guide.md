## Federatorai-Emulator deploy guide

### Prepare yaml file

**comfigmap.yaml**: Emulator configuration file, define emulator node/pods data.  
**daemonset.yaml**: Deploy emulator pod service.  
**prometheus_fake_node_rule.yaml**: Define prometheus rule to match the emulator job name.  
**service.yaml**: Define the emulator service port for the Prometheus scraping data.  
**servicemonitor.yaml**: Register the Prometheus to monitor the emulator pod.

### Deploy step

$ oc apply -f ./deployment

### Deploy another node

Reference the folder "othernode" to modify or add second job name. Use deploy command to use the folder "othernode"
 yaml files.
