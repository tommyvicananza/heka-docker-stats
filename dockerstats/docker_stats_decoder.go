package dockerstats

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/mozilla-services/heka/pipeline"
)

type DockerStatsDecoder struct {
	*DockerStatsInputConfig
}

func (input *DockerStatsDecoder) Init(config interface{}) error {
	return nil
}

func (input *DockerStatsDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	var buf bytes.Buffer
	pack.Message.SetPayload(string(buf.Bytes()))

	packs = []*pipeline.PipelinePack{pack}
	return
}

func (*DockerStatsDecoder) decode(pack *pipeline.PipelinePack) bytes.Buffer {
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

	json, _ := json.Marshal(stats)
	return *bytes.NewBuffer(json)
}

func init() {
	pipeline.RegisterPlugin("DockerStatsDecoder", func() interface{} {
		return new(DockerStatsDecoder)
	})
}
