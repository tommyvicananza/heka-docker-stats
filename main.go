package main

import (
	"fmt"
	"strings"

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

type dockerStats map[string]*dockerStat

func main() {
	var (
		previousCPU, previousSystem uint64
		preCPUStats, stats          docker.Stats
		mstats                      dockerStats
	)
	client, _ := docker.NewClientFromEnv()
	containers, _ := client.ListContainers(docker.ListContainersOptions{Filters: map[string][]string{"status": {"running"}}})
	mstats = make(map[string]*dockerStat)
	for _, container := range containers {
		preCPUStats, _ = client.StatsStatic(container.ID)
		previousCPU = preCPUStats.CPUStats.CPUUsage.TotalUsage
		previousSystem = preCPUStats.CPUStats.SystemCPUUsage
		stats, _ = client.StatsStatic(container.ID)

		mstats[container.ID] = &dockerStat{}
		mstats[container.ID].CPUPercent = calculateCPUPercent(previousCPU, previousSystem, &stats)
		mstats[container.ID].MemPercent = calculateMemPercent(&stats)
		mstats[container.ID].MemUsage = stats.MemoryStats.Usage
		mstats[container.ID].MemLimit = stats.MemoryStats.Limit
		for _, networkstat := range stats.Networks {
			mstats[container.ID].NetworkRx = networkstat.RxBytes
			mstats[container.ID].NetworkTx = networkstat.TxBytes
		}
		mstats[container.ID].BlockRead, mstats[container.ID].BlockWrite = calculateBlockIO(stats)
		fmt.Printf("Container%s\tCPU: %.2f\tMEM USAGE / LIMIT: %d / %d\tMEM: %.2f\tNET I/O: %d / %d\tBLOCK I/O: %d, %d\n", container.Names, mstats[container.ID].CPUPercent, mstats[container.ID].MemUsage, mstats[container.ID].MemLimit, mstats[container.ID].MemPercent, mstats[container.ID].NetworkRx, mstats[container.ID].NetworkTx, mstats[container.ID].BlockRead, mstats[container.ID].BlockWrite)
	}
}

func calculateMemPercent(stats *docker.Stats) float64 {
	var memPercent = 0.0
	if stats.MemoryStats.Limit != 0 {
		memPercent = float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}
	return memPercent
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
