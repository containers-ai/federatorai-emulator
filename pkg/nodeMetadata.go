package emulator

import (
	"encoding/json"
	Datahub "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type ConvNodeMetaData struct {
	Node *Datahub.Node
}

func NewNodeMetadata(rawData []byte) *ConvNodeMetaData{
	var nodeMetaData Datahub.Node
	err := json.Unmarshal(rawData, &nodeMetaData)
	if err != nil {
		return nil
	}
	return &ConvNodeMetaData{&nodeMetaData}
}
