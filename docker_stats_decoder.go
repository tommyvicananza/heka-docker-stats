package dockerstats

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mozilla-services/heka/pipeline"
)

type StatsPayload struct {
	Hostname      string  `json:"hostname"`
	ContainerName string  `json:"container_name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemPercent    float64 `json:"mem_percent"`
	MemUsage      int64   `json:"mem_usage"`
	MemLimit      int64   `json:"mem_limit"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
	BlockRead     int64   `json:"block_read"`
	BlockWrite    int64   `json:"block_write"`
	TimeStamp     int64   `json:"timestamp"`
}

type DockerStatsDecoder struct {
	*DockerStatsInputConfig
}

func (input *DockerStatsDecoder) Init(config interface{}) error {
	return nil
}

func (input *DockerStatsDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	var buf bytes.Buffer
	buf = input.decode(pack)
	pack.Message.SetPayload(string(buf.Bytes()))
	packs = []*pipeline.PipelinePack{pack}
	return
}

func (*DockerStatsDecoder) decode(pack *pipeline.PipelinePack) bytes.Buffer {
	containerName, _ := pack.Message.GetFieldValue("ContainerName")
	cpuPercent, _ := pack.Message.GetFieldValue("CPUPercent")
	memoryPercent, _ := pack.Message.GetFieldValue("MemoryPercent")
	memoryUsage, _ := pack.Message.GetFieldValue("MemoryUsage")
	memoryLimit, _ := pack.Message.GetFieldValue("MemoryLimit")
	networkInput, _ := pack.Message.GetFieldValue("NetworkInput")
	networkOutput, _ := pack.Message.GetFieldValue("NetworkOutput")
	blockInput, _ := pack.Message.GetFieldValue("BlockInput")
	blockOutput, _ := pack.Message.GetFieldValue("BlockOutput")

	stats := StatsPayload{
		Hostname:      *pack.Message.Hostname,
		ContainerName: string(containerName),
		CPUPercent:    cpuPercent,
		MemPercent:    memoryPercent,
		MemUsage:      memoryUsage,
		MemLimit:      memoryLimit,
		NetworkRx:     networkInput,
		NetworkTx:     networkOutput,
		BlockRead:     blockInput,
		BlockWrite:    blockOutput,
		TimeStamp:     *pack.Message.Timestamp,
	}

	json, err := json.Marshal(stats)
	if err != nil {
		fmt.Println(err)
	}
	return *bytes.NewBuffer(json)
}

func init() {
	pipeline.RegisterPlugin("DockerStatsDecoder", func() interface{} {
		return new(DockerStatsDecoder)
	})
}
