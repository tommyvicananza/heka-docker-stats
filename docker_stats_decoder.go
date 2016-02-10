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
		Hostname:      "hola",
		ContainerName: "hola",
		CPUPercent:    1.0,
		MemPercent:    2.0,
		MemUsage:      3,
		MemLimit:      4,
		NetworkRx:     5,
		NetworkTx:     6,
		BlockRead:     7,
		BlockWrite:    8,
		TimeStamp:     932,
	}
	//a := pack.Message.Hostname
	//fmt.Println(a)
	//b, _ := pack.Message.GetFieldValue("ContainerName")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("CPUPercent")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("MemoryPercent")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("MemoryUsage")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("MemoryLimit")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("NetInput")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("NetOuput")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("BlockRead")
	//fmt.Println(b)
	//b, _ = pack.Message.GetFieldValue("BlockWrite")
	//fmt.Println(b)
	//t := pack.Message.Timestamp
	//fmt.Println(t)

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
