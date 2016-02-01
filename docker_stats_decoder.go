package dockerstats

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mozilla-services/heka/pipeline"
)

type DockerStatsDecoder struct {
	*DockerStatsInputConfig
}

func (input *DockerStatsDecoder) Init(config interface{}) error {
	fmt.Printf("Decoder Init")
	return nil
}

func (input *DockerStatsDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	fmt.Printf(pack.Message.GetPayload())
	var buf bytes.Buffer

	buf = input.decode(pack)
	pack.Message.SetPayload(string(buf.Bytes()))
	fmt.Printf(pack.Message.GetPayload())
	packs = []*pipeline.PipelinePack{pack}
	return
}

func (*DockerStatsDecoder) decode(pack *pipeline.PipelinePack) bytes.Buffer {

	fmt.Printf("decode")
	var stats = make(map[string]string)

	reader := bufio.NewReader(strings.NewReader(pack.Message.GetPayload()))
	for {
		data, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		fields := strings.Split(string(data), " ")

		stats[fields[0]] = fields[1]
	}

	fmt.Println("hola")
	json, _ := json.Marshal(stats)
	return *bytes.NewBuffer(json)
}

func init() {
	pipeline.RegisterPlugin("DockerStatsDecoder", func() interface{} {
		return new(DockerStatsDecoder)
	})
}
