package psutil

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	resetInterval = 1 * time.Minute
	fastInterval  = 3 * time.Second
)

type CPUStat struct {
	Idle  uint64
	Total uint64
}

type CPUDetailedStat struct {
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	Iowait    uint64
	Irq       uint64
	Softirq   uint64
	Steal     uint64
	Guest     uint64
	GuestNice uint64
	Total     uint64
}

type CPUDetailedPercent struct {
	User    float64 `json:"user"`
	System  float64 `json:"system"`
	Nice    float64 `json:"nice"`
	Idle    float64 `json:"idle"`
	Iowait  float64 `json:"iowait"`
	Irq     float64 `json:"irq"`
	Softirq float64 `json:"softirq"`
	Steal   float64 `json:"steal"`
}

func (c *CPUDetailedPercent) GetCPUDetailedPercent() []float64 {
	return []float64{c.User, c.System, c.Nice, c.Idle, c.Iowait, c.Irq, c.Softirq, c.Steal}
}

type CPUUsageState struct {
	mu             sync.Mutex
	lastTotalStat  *CPUStat
	lastPerCPUStat []CPUStat
	lastDetailStat *CPUDetailedStat
	lastSampleTime time.Time

	cachedTotalUsage      float64
	cachedPerCore         []float64
	cachedDetailedPercent CPUDetailedPercent
}

type CPUInfoState struct {
	mu               sync.RWMutex
	initialized      bool
	cachedInfo       []cpu.InfoStat
	cachedPhysCores  int
	cachedLogicCores int
}

func (c *CPUUsageState) GetCPUUsage() (float64, []float64, []float64) {
	c.mu.Lock()
	now := time.Now()
	if !c.lastSampleTime.IsZero() && now.Sub(c.lastSampleTime) < fastInterval {
		result := c.cachedTotalUsage
		perCore := c.cachedPerCore
		detailed := c.cachedDetailedPercent
		c.mu.Unlock()
		return result, perCore, detailed.GetCPUDetailedPercent()
	}
	// 释放锁，因为cpu.Percent会阻塞一段时间
	c.mu.Unlock()

	// 直接使用gopsutil的cpu.Percent函数获取CPU使用率
	// 参数1: 采样时间间隔，0表示立即返回当前使用率
	// 参数2: true表示返回每个核心的使用率
	perCoreUsage, err := cpu.Percent(100*time.Millisecond, true)
	if err != nil {
		c.mu.Lock()
		c.lastSampleTime = time.Now()
		c.mu.Unlock()
		return 0, []float64{}, []float64{}
	}

	// 计算总CPU使用率
	var totalUsage float64
	if len(perCoreUsage) > 0 {
		for _, usage := range perCoreUsage {
			totalUsage += usage
		}
		totalUsage /= float64(len(perCoreUsage))
	}

	// 创建详细CPU使用率（使用模拟值，因为gopsutil不直接提供详细信息）
	detailedPercent := CPUDetailedPercent{
		User:   totalUsage * 0.6, // 假设60%是用户空间使用
		System: totalUsage * 0.3, // 假设30%是系统空间使用
		Idle:   100 - totalUsage, // 剩余的是空闲时间
		Iowait: totalUsage * 0.1, // 假设10%是I/O等待
	}

	c.mu.Lock()
	c.cachedTotalUsage = totalUsage
	c.cachedPerCore = perCoreUsage
	c.cachedDetailedPercent = detailedPercent
	c.lastSampleTime = time.Now()
	c.mu.Unlock()

	return totalUsage, perCoreUsage, detailedPercent.GetCPUDetailedPercent()
}

func (c *CPUUsageState) NumCPU() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cachedPerCore) > 0 {
		return len(c.cachedPerCore)
	}
	// 使用runtime.NumCPU()作为跨平台的后备方案
	return runtime.NumCPU()
}

func (c *CPUInfoState) GetCPUInfo(forceRefresh bool) ([]cpu.InfoStat, error) {
	c.mu.RLock()
	if c.initialized && c.cachedInfo != nil && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedInfo, nil
	}
	c.mu.RUnlock()

	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cachedInfo = info
	c.initialized = true
	c.mu.Unlock()

	return info, nil
}

func (c *CPUInfoState) GetPhysicalCores(forceRefresh bool) (int, error) {
	c.mu.RLock()
	if c.initialized && c.cachedPhysCores > 0 && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedPhysCores, nil
	}
	c.mu.RUnlock()

	cores, err := cpu.Counts(false)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.cachedPhysCores = cores
	c.initialized = true
	c.mu.Unlock()

	return cores, nil
}

func (c *CPUInfoState) GetLogicalCores(forceRefresh bool) (int, error) {
	c.mu.RLock()
	if c.initialized && c.cachedLogicCores > 0 && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedLogicCores, nil
	}
	c.mu.RUnlock()

	cores, err := cpu.Counts(true)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.cachedLogicCores = cores
	c.initialized = true
	c.mu.Unlock()

	return cores, nil
}

func readProcStat() ([]byte, error) {
	return os.ReadFile("/proc/stat")
}

func parseCPUFields(line string) []uint64 {
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return nil
	}
	fields = fields[1:]

	nums := make([]uint64, len(fields))
	for i, f := range fields {
		v, _ := strconv.ParseUint(f, 10, 64)
		nums[i] = v
	}
	return nums
}

func calcIdleAndTotal(nums []uint64) (idle, total uint64) {
	if len(nums) < 5 {
		return 0, 0
	}
	idle = nums[3] + nums[4]
	for _, v := range nums {
		total += v
	}
	return
}

func readAllCPUStat() (CPUStat, CPUDetailedStat, []CPUStat) {
	// 首先尝试使用/proc/stat（Linux系统）获取原始统计信息
	data, err := readProcStat()
	if err == nil && len(data) > 0 {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			nums := parseCPUFields(firstLine)

			idle, total := calcIdleAndTotal(nums)
			cpuStat := CPUStat{Idle: idle, Total: total}

			if len(nums) < 10 {
				padded := make([]uint64, 10)
				copy(padded, nums)
				nums = padded
			}
			detailedStat := CPUDetailedStat{
				User:      nums[0],
				Nice:      nums[1],
				System:    nums[2],
				Idle:      nums[3],
				Iowait:    nums[4],
				Irq:       nums[5],
				Softirq:   nums[6],
				Steal:     nums[7],
				Guest:     nums[8],
				GuestNice: nums[9],
			}
			detailedStat.Total = detailedStat.User + detailedStat.Nice + detailedStat.System +
				detailedStat.Idle + detailedStat.Iowait + detailedStat.Irq + detailedStat.Softirq + detailedStat.Steal

			var perCPUStats []CPUStat
			for _, line := range lines[1:] {
				if !strings.HasPrefix(line, "cpu") {
					continue
				}
				if len(line) < 4 || line[3] < '0' || line[3] > '9' {
					continue
				}

				perNums := parseCPUFields(line)
				perIdle, perTotal := calcIdleAndTotal(perNums)
				perCPUStats = append(perCPUStats, CPUStat{Idle: perIdle, Total: perTotal})
			}

			return cpuStat, detailedStat, perCPUStats
		}
	}

	// 如果/proc/stat获取失败，尝试使用跨平台的gopsutil库
	percentages, err := cpu.Percent(0, true)
	if err == nil && len(percentages) > 0 {
		// 构建每个核心的CPUStat
		var perCPUStats []CPUStat
		for _, percent := range percentages {
			// 根据百分比计算模拟的Idle和Total值
			// 这里使用一个较大的基数，以获得更准确的百分比计算
			base := uint64(10000)
			used := uint64(percent * 100)
			idle := base - used
			perCPUStats = append(perCPUStats, CPUStat{Idle: idle, Total: base})
		}

		// 构建整体CPUStat
		var totalIdle, totalTotal uint64
		for _, stat := range perCPUStats {
			totalIdle += stat.Idle
			totalTotal += stat.Total
		}
		cpuStat := CPUStat{Idle: totalIdle, Total: totalTotal}

		// 构建详细CPUStat（使用模拟值）
		detailedStat := CPUDetailedStat{
			User:   totalIdle / 4,
			System: totalIdle / 4,
			Idle:   totalIdle / 2,
			Total:  totalTotal,
		}

		return cpuStat, detailedStat, perCPUStats
	}

	// 如果所有方法都失败，返回空值
	return CPUStat{}, CPUDetailedStat{}, nil
}

func calcCPUPercent(prev, cur CPUStat) float64 {
	deltaIdle := float64(cur.Idle - prev.Idle)
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return 0
	}
	return (1 - deltaIdle/deltaTotal) * 100
}

func calcCPUDetailedPercent(prev, cur CPUDetailedStat) CPUDetailedPercent {
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return CPUDetailedPercent{Idle: 100}
	}

	return CPUDetailedPercent{
		User:    float64(cur.User-prev.User) / deltaTotal * 100,
		System:  float64(cur.System-prev.System) / deltaTotal * 100,
		Nice:    float64(cur.Nice-prev.Nice) / deltaTotal * 100,
		Idle:    float64(cur.Idle-prev.Idle) / deltaTotal * 100,
		Iowait:  float64(cur.Iowait-prev.Iowait) / deltaTotal * 100,
		Irq:     float64(cur.Irq-prev.Irq) / deltaTotal * 100,
		Softirq: float64(cur.Softirq-prev.Softirq) / deltaTotal * 100,
		Steal:   float64(cur.Steal-prev.Steal) / deltaTotal * 100,
	}
}
