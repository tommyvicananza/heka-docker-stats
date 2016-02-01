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

type DockerStatsInputConfig struct {
	TickerInterval uint `toml:"ticker_interval"`
}

type DockerStatsInput struct {
	*DockerStatsInputConfig
	stop   chan bool
	runner pipeline.InputRunner
}

func (input *DockerStatsInput) ConfigStruct() interface{} {
	return &DockerStatsInputConfig{
		TickerInterval: uint(60),
	}
}
func (input *DockerStatsInput) Init(config interface{}) error {
	input.DockerStatsInputConfig = config.(*DockerStatsInputConfig)
	input.stop = make(chan bool)
	return nil
}

func (input *DockerStatsInput) Stop() {
	close(input.stop)
}

func (input *DockerStatsInput) Run(runner pipeline.InputRunner,
	helper pipeline.PluginHelper) error {

	fmt.Printf("Run")

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
			previousCPU, previousSystem uint64
			preCPUStats, stats          docker.Stats
		)
		client, _ := docker.NewClientFromEnv()
		containers, _ := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"status": {"running"}}})
		fmt.Printf("Input Run: ahí antes de meterme, %d", len(containers))
		for _, container := range containers {
			fmt.Printf("Input Run: ahí me meto")
			pack = <-packSupply
			pack.Message.SetUuid(uuid.NewRandom())
			pack.Message.SetTimestamp(time.Now().UnixNano())
			pack.Message.SetType("docker.metrics")
			pack.Message.SetHostname(hostname)

			preCPUStats, _ = client.StatsStatic(container.ID)
			previousCPU = preCPUStats.CPUStats.CPUUsage.TotalUsage
			previousSystem = preCPUStats.CPUStats.SystemCPUUsage
			stats, _ = client.StatsStatic(container.ID)

			//container.Names, mstats[container.ID].CPUPercent, mstats[container.ID].MemUsage, mstats[container.ID].MemLimit,
			//mstats[container.ID].MemPercent, mstats[container.ID].NetworkRx, mstats[container.ID].NetworkTx, mstats[container.ID].BlockRead, mstats[container.ID].BlockWrite
			containerID, _ := message.NewField("ContainerId", string(container.ID), "")
			pack.Message.AddField(containerID)

			cpuPercent, _ := message.NewField("CPUPercent", calculateCPUPercent(previousCPU, previousSystem, &stats), "")
			pack.Message.AddField(cpuPercent)

			memLimit, _ := message.NewField("MemLimit", string(stats.MemoryStats.Limit), "")
			pack.Message.AddField(memLimit)

			memUsage, _ := message.NewField("MemUsage", string(stats.MemoryStats.Usage), "")
			pack.Message.AddField(memUsage)

			memPercent, _ := message.NewField("MemPercent", calculateMemPercent(&stats), "")
			pack.Message.AddField(memPercent)

			for _, networkstat := range stats.Networks {
				networkRx, _ := message.NewField("NetworkRx", string(networkstat.RxBytes), "")
				pack.Message.AddField(networkRx)

				networkTx, _ := message.NewField("NetworkRx", string(networkstat.TxBytes), "")
				pack.Message.AddField(networkTx)
			}
			br, bw := calculateBlockIO(stats)
			blockRead, _ := message.NewField("BlockRead", string(br), "")
			pack.Message.AddField(blockRead)

			blockWrite, _ := message.NewField("BlockWrite", string(bw), "")
			pack.Message.AddField(blockWrite)

			//mstats[container.ID] = &dockerStat{}
			//mstats[container.ID].CPUPercent = calculateCPUPercent(previousCPU, previousSystem, &stats)
			//mstats[container.ID].MemPercent = calculateMemPercent(&stats)
			//mstats[container.ID].MemUsage = stats.MemoryStats.Usage
			//mstats[container.ID].MemLimit = stats.MemoryStats.Limit
			//mstats[container.ID].BlockRead, mstats[container.ID].BlockWrite = calculateBlockIO(stats)
			//fmt.Printf("Container%s\tCPU: %.2f\tMEM USAGE / LIMIT: %d / %d\tMEM: %.2f\tNET I/O: %d / %d\tBLOCK I/O: %d, %d\n", container.Names, mstats[container.ID].CPUPercent, mstats[container.ID].MemUsage, mstats[container.ID].MemLimit, mstats[container.ID].MemPercent, mstats[container.ID].NetworkRx, mstats[container.ID].NetworkTx, mstats[container.ID].BlockRead, mstats[container.ID].BlockWrite)
			runner.Deliver(pack)
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

func calculateBlockIO(stats docker.Stats) (blkRead uint64, blkWrite uint64) {
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
