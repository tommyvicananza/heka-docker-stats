package dockerstats

import (
	"fmt"
	"strings"
	"time"

	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
	"github.com/pborman/uuid"
	"github.com/tommyvicananza/go-dockerclient"
)

type dockerStat struct {
	CPUPercent float64
	MemPercent float64
	MemUsage   uint64
	MemLimit   uint64
	NetworkRx  uint64
	NetworkTx  uint64
	BlockRead  uint64
	BlockWrite uint64
}

type DockerStatsInputConfig struct {
	TickerInterval uint   `toml:"ticker_interval"`
	NameFromEnv    string `toml:"name_from_env"`
}

type DockerStatsInput struct {
	*DockerStatsInputConfig
	stop           chan bool
	runner         pipeline.InputRunner
	cacheHostnames map[string]string
}

func (input *DockerStatsInput) ConfigStruct() interface{} {
	return &DockerStatsInputConfig{
		TickerInterval: uint(60),
	}
}
func (input *DockerStatsInput) Init(config interface{}) error {
	input.DockerStatsInputConfig = config.(*DockerStatsInputConfig)
	input.stop = make(chan bool)
	input.cacheHostnames = make(map[string]string)
	return nil
}

func (input *DockerStatsInput) Stop() {
	close(input.stop)
}

func (input *DockerStatsInput) Run(runner pipeline.InputRunner,
	helper pipeline.PluginHelper) error {

	var pack *pipeline.PipelinePack

	input.runner = runner
	packSupply := runner.InChan()
	tickChan := runner.Ticker()

	hostname := helper.PipelineConfig().Hostname()

	for {
		select {
		case <-input.stop:
			return nil
		case <-tickChan:
		}
		var (
			// test                        chan bool
			//err                         error
			previousCPU, previousSystem uint64
			mstats                      *dockerStat
			preCPUStats, stats          *docker.Stats
		)
		endpoint := "unix:///var/run/docker.sock"
		client, _ := docker.NewClient(endpoint)
		containers, _ := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"status": {"running"}}})
		for _, container := range containers {
			if containerName, exists := input.cacheHostnames[container.ID]; !exists {
				containerName = strings.Replace(container.Names[0], "/", "", -1)
				input.cacheHostnames[container.ID] = containerName
				if input.NameFromEnv != "" {
					con, _ := client.InspectContainer(container.ID)
					for _, value := range con.Config.Env {
						parts := strings.SplitN(value, "=", 2)
						if len(parts) == 2 {
							if input.NameFromEnv == parts[0] {
								containerName = parts[1]
								input.cacheHostnames[container.ID] = containerName
								break
							}
						}
					}
				}
			}
			fmt.Println(input.cacheHostnames[container.ID])

			// go func() {
			// 	test = make(chan bool)
			opts := docker.StatsStaticOptions{ID: container.ID, Stream: false}
			preCPUStats, _ = client.StatsStatic(opts)
			if preCPUStats == nil {
				fmt.Println("Es vacio")
				continue
			}
			fmt.Println("Hecho precpu")
			previousCPU = preCPUStats.CPUStats.CPUUsage.TotalUsage
			previousSystem = preCPUStats.CPUStats.SystemCPUUsage

			stats, _ = client.StatsStatic(opts)
			if stats == nil {
				fmt.Println("Es vacio")
				continue
			}
			fmt.Println("Hecho cpu")
			mstats = &dockerStat{}
			mstats.CPUPercent = calculateCPUPercent(previousCPU, previousSystem, stats)
			mstats.MemPercent = calculateMemPercent(stats)
			mstats.MemUsage = stats.MemoryStats.Usage
			mstats.MemLimit = stats.MemoryStats.Limit
			mstats.BlockRead, mstats.BlockWrite = calculateBlockIO(stats)
			for _, networkstat := range stats.Networks {
				mstats.NetworkRx = networkstat.RxBytes
				mstats.NetworkTx = networkstat.TxBytes
			}
			pack = <-packSupply
			pack.Message.SetUuid(uuid.NewRandom())
			pack.Message.SetTimestamp(time.Now().UnixNano())
			pack.Message.SetType("DockerStats")
			pack.Message.SetHostname(hostname)

			containerName, _ := message.NewField("ContainerName", string(strings.Replace(input.cacheHostnames[container.ID], "-", "_", -1)), "")
			pack.Message.AddField(containerName)
			cpuPercent, _ := message.NewField("CPUPercent", float64(mstats.CPUPercent), "")
			pack.Message.AddField(cpuPercent)
			memUsage, _ := message.NewField("MemoryUsage", uint64(mstats.MemUsage), "")
			pack.Message.AddField(memUsage)
			memLimit, _ := message.NewField("MemoryLimit", uint64(mstats.MemLimit), "")
			pack.Message.AddField(memLimit)
			memPercent, _ := message.NewField("MemoryPercent", float64(mstats.MemPercent), "")
			pack.Message.AddField(memPercent)
			netInput, _ := message.NewField("NetInput", uint64(mstats.NetworkRx), "")
			pack.Message.AddField(netInput)
			netOutput, _ := message.NewField("NetOutput", uint64(mstats.NetworkTx), "")
			pack.Message.AddField(netOuput)
			blockRead, _ := message.NewField("BlockRead", uint64(mstats.BlockRead), "")
			pack.Message.AddField(blockRead)
			blockWrite, _ := message.NewField("BlockWrite", uint64(mstats.BlockWrite), "")
			pack.Message.AddField(blockWrite)

			pack.Message.SetPayload(fmt.Sprintf("container_name %s\ncpu %.2f\nmem_usage %d\nmem_limit %d\nmem %.2f\nnet_input %d\nnet_output %d\nblock_input %d\nblock_output %d",
				strings.Replace(input.cacheHostnames[container.ID], "-", "_", -1),
				mstats.CPUPercent,
				mstats.MemUsage,
				mstats.MemLimit,
				mstats.MemPercent,
				mstats.NetworkRx,
				mstats.NetworkTx,
				mstats.BlockRead,
				mstats.BlockWrite))
			runner.Deliver(pack)
			// test <- true
			// }()
		}
	}
	return nil
}

func init() {
	pipeline.RegisterPlugin("DockerStatsInput", func() interface{} {
		return new(DockerStatsInput)
	})
}

func calculateCPUPercent(previousCPU, previousSystem uint64, stats *docker.Stats) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)

		// calculate the change for the entire system between readings
		systemDelta = float64(stats.CPUStats.SystemCPUUsage) - float64(previousSystem)
	)
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateBlockIO(stats *docker.Stats) (blkRead uint64, blkWrite uint64) {
	blkio := stats.BlkioStats
	for _, bioEntry := range blkio.IOServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return
}

func calculateMemPercent(stats *docker.Stats) float64 {
	var memPercent = 0.0
	if stats.MemoryStats.Limit != 0 {
		memPercent = float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return memPercent
}
