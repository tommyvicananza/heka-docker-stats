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
		Hostname:      pack.Message.read_message("Hostname"),
		ContainerName: pack.Message.read_message("ContainerName"),
		CPUPercent:    pack.Message.read_message("CPUPercent"),
		MemPercent:    pack.Message.read_message("MemoryPercent"),
		MemUsage:      pack.Message.read_message("MemoryUsage"),
		MemLimit:      pack.Message.read_message("MemoryLimit"),
		NetworkRx:     pack.Message.read_message("NetInput"),
		NetworkTx:     pack.Message.read_message("NetOuput"),
		BlockRead:     pack.Message.read_message("BlockRead"),
		BlockWrite:    pack.Message.read_message("BlockWrite"),
		TimeStamp:     pack.Message.read_message("Timestamp"),
	}

	json, _ := json.Marshal(stats)
	return *bytes.NewBuffer(json)
}

func init() {
	pipeline.RegisterPlugin("DockerStatsDecoder", func() interface{} {
		return new(DockerStatsDecoder)
	})
}
