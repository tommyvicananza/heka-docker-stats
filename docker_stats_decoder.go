package dockerstats

import (
	"bytes"
	"encoding/json"

	"github.com/mozilla-services/heka/pipeline"
)

type StatsPayload struct {
	Hostname      string  `json:"hostname"`
	ContainerName string  `json:"container_name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemPercent    float64 `json:"mem_percent"`
	MemUsage      uint64  `json:"mem_usage"`
	MemLimit      uint64  `json:"mem_limit"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	BlockRead     uint64  `json:"block_read"`
	BlockWrite    uint64  `json:"block_write"`
	TimeStamp     uint64  `json:"timestamp"`
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
	stats := StatsPayload{
		Hostname:      pack.Message.GetFieldValue("Hostname"),
		ContainerName: pack.Message.GetFieldValue("ContainerName"),
		CPUPercent:    pack.Message.GetFieldValue("CPUPercent"),
		MemPercent:    pack.Message.GetFieldValue("MemoryPercent"),
		MemUsage:      pack.Message.GetFieldValue("MemoryUsage"),
		MemLimit:      pack.Message.GetFieldValue("MemoryLimit"),
		NetworkRx:     pack.Message.GetFieldValue("NetInput"),
		NetworkTx:     pack.Message.GetFieldValue("NetOuput"),
		BlockRead:     pack.Message.GetFieldValue("BlockRead"),
		BlockWrite:    pack.Message.GetFieldValue("BlockWrite"),
		TimeStamp:     pack.Message.GetFieldValue("Timestamp"),
	}

	json, _ := json.Marshal(stats)
	return *bytes.NewBuffer(json)
}

func init() {
	pipeline.RegisterPlugin("DockerStatsDecoder", func() interface{} {
		return new(DockerStatsDecoder)
	})
}
