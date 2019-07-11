package emulator

import (
	"encoding/json"
	Datahub "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/jinzhu/copier"
)
/*
type PodMetadata struct {
	NamespacedName struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	} `json:"namespaced_name"`
	ResourceLink string `json:"resource_link"`
	Containers   []struct {
		Name          string `json:"name"`
		LimitResource []struct {
			MetricType int `json:"metric_type"`
			Data       []struct {
				NumValue string `json:"num_value"`
			} `json:"data"`
		} `json:"limit_resource"`
		RequestResource []struct {
			MetricType int `json:"metric_type"`
			Data       []struct {
				NumValue string `json:"num_value"`
			} `json:"data"`
		} `json:"request_resource"`
		Status struct {
			State struct {
				Waiting struct {
					Reason string `json:"reason"`
				} `json:"waiting"`
				Running struct {
					StartedAt struct {
						Seconds int `json:"seconds"`
					} `json:"started_at"`
				} `json:"running"`
				Terminated struct {
					StartedAt struct {
					} `json:"started_at"`
					FinishedAt struct {
					} `json:"finished_at"`
				} `json:"terminated"`
			} `json:"state"`
			LastTerminationState struct {
				Waiting struct {
				} `json:"waiting"`
				Running struct {
					StartedAt struct {
					} `json:"started_at"`
				} `json:"running"`
				Terminated struct {
					ExitCode  int    `json:"exit_code"`
					Reason    string `json:"reason"`
					StartedAt struct {
						Seconds int `json:"seconds"`
					} `json:"started_at"`
					FinishedAt struct {
						Seconds int `json:"seconds"`
					} `json:"finished_at"`
				} `json:"terminated"`
			} `json:"last_termination_state"`
			RestartCount int `json:"restart_count"`
		} `json:"status"`
	} `json:"containers"`
	AlamedaScaler struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	} `json:"alameda_scaler"`
	NodeName  string `json:"node_name"`
	StartTime struct {
		Seconds int `json:"seconds"`
	} `json:"start_time"`
	Policy        int `json:"policy"`
	TopController struct {
		NamespacedName struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		} `json:"namespaced_name"`
		Kind     int `json:"kind"`
		Replicas int `json:"Replicas"`
	} `json:"top_controller"`
	Status struct {
		Phase int `json:"phase"`
	} `json:"status"`
}
*/

type ConvPodMetadata struct {
	Pod *Datahub.Pod
}

func NewPodMetadata(rawData []byte) *ConvPodMetadata {
	var podMetadata Datahub.Pod
	if len(rawData) > 0 {
		err := json.Unmarshal(rawData, &podMetadata)
		if err != nil {
			return nil
		}
	}
	return &ConvPodMetadata{&podMetadata}
}

func (c *ConvPodMetadata) SetPodName(podName string) {
	c.Pod.NamespacedName.Name = podName
}

func (c *ConvPodMetadata) SetNamesapce(namespace string) {
	c.Pod.NamespacedName.Namespace = namespace
	c.Pod.AlamedaScaler.Namespace = namespace
	c.Pod.TopController.NamespacedName.Namespace = namespace
	c.Pod.NamespacedName.Name = namespace
}

func (c *ConvPodMetadata) SetContainerName(containerName string) {
	c.Pod.AlamedaScaler.Name = "alameda-" + containerName
	c.Pod.TopController.NamespacedName.Name = containerName
	c.Pod.Containers[0].Name = containerName
}

func (c *ConvPodMetadata) SetNodeName(nodeName string) {
	c.Pod.NodeName = nodeName
}

func (c *ConvPodMetadata) GetPod() *Datahub.Pod {
	return c.Pod
}

func (c *ConvPodMetadata) SetPod(pod *Datahub.Pod)  {
	p := new(Datahub.Pod)
	copier.Copy(&p, pod)
	c.Pod = p
}