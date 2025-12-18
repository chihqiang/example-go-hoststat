package handles

import (
	"chihqiang/hoststat/psutil"
	"cmp"
	"github.com/shirou/gopsutil/v4/process"
	"net"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	psNet "github.com/shirou/gopsutil/v4/net"
)

func getBaseInfo() (*BaseInfo, error) {
	hostInfo, err := psutil.HOST.GetHostInfo(false)
	if err != nil {
		return nil, err
	}
	bi := BaseInfo{
		Hostname:        hostInfo.Hostname,
		OS:              hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformFamily:  hostInfo.PlatformFamily,
		PlatformVersion: hostInfo.PlatformVersion,
		PrettyDistro:    psutil.HOST.GetDistro(),
		KernelArch:      hostInfo.KernelArch,
		KernelVersion:   hostInfo.KernelVersion,
		IPV4Addr:        loadOutboundIP(),
		SystemProxy:     "noProxy",
	}
	if proxy := cmp.Or(os.Getenv("http_proxy"), os.Getenv("HTTP_PROXY")); proxy != "" {
		bi.SystemProxy = proxy
	}
	cpuInfo, err := psutil.CPUInfo.GetCPUInfo(false)
	if err == nil && len(cpuInfo) > 0 {
		bi.CPUModelName = cpuInfo[0].ModelName
	}
	bi.CPUCores, _ = psutil.CPUInfo.GetPhysicalCores(false)
	bi.CPULogicalCores, _ = psutil.CPUInfo.GetLogicalCores(false)
	bi.CPUMhz = cpuInfo[0].Mhz

	currentInfo, _ := getCurrentInfo()
	bi.CurrentInfo = currentInfo
	return &bi, nil
}

func getCurrentInfo() (*CurrentInfo, error) {
	var currentInfo CurrentInfo
	hostInfo, _ := psutil.HOST.GetHostInfo(false)
	currentInfo.Uptime = hostInfo.Uptime
	currentInfo.TimeSinceUptime = time.Unix(int64(hostInfo.BootTime), 0).Format(DateTimeLayout)
	currentInfo.Procs = hostInfo.Procs
	currentInfo.CPUTotal, _ = psutil.CPUInfo.GetLogicalCores(false)
	cpuUsedPercent, perCore, cpuDetailedPercent := psutil.CPU.GetCPUUsage()
	if len(perCore) == 0 {
		currentInfo.CPUTotal = psutil.CPU.NumCPU()
	} else {
		currentInfo.CPUTotal = len(perCore)
	}
	currentInfo.CPUPercent = perCore
	currentInfo.CPUUsedPercent = cpuUsedPercent
	currentInfo.CPUUsed = cpuUsedPercent * 0.01 * float64(currentInfo.CPUTotal)
	currentInfo.CPUDetailedPercent = cpuDetailedPercent

	loadInfo, _ := load.Avg()
	currentInfo.Load1 = loadInfo.Load1
	currentInfo.Load5 = loadInfo.Load5
	currentInfo.Load15 = loadInfo.Load15
	currentInfo.LoadUsagePercent = loadInfo.Load1 / (float64(currentInfo.CPUTotal*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	currentInfo.MemoryTotal = memoryInfo.Total
	currentInfo.MemoryUsed = memoryInfo.Used
	currentInfo.MemoryFree = memoryInfo.Free
	currentInfo.MemoryCache = memoryInfo.Cached + memoryInfo.Buffers
	currentInfo.MemoryShard = memoryInfo.Shared
	currentInfo.MemoryAvailable = memoryInfo.Available
	currentInfo.MemoryUsedPercent = memoryInfo.UsedPercent

	swapInfo, _ := mem.SwapMemory()
	currentInfo.SwapMemoryTotal = swapInfo.Total
	currentInfo.SwapMemoryAvailable = swapInfo.Free
	currentInfo.SwapMemoryUsed = swapInfo.Used
	currentInfo.SwapMemoryUsedPercent = swapInfo.UsedPercent

	currentInfo.DiskData = loadDiskInfo()

	diskInfos, _ := disk.IOCounters()
	for _, state := range diskInfos {
		currentInfo.IOReadBytes += state.ReadBytes
		currentInfo.IOWriteBytes += state.WriteBytes
		currentInfo.IOCount += state.ReadCount + state.WriteCount
		currentInfo.IOReadTime += state.ReadTime
		currentInfo.IOWriteTime += state.WriteTime
	}

	netInfos, _ := psNet.IOCounters(false)
	if len(netInfos) != 0 {
		currentInfo.NetBytesSent = netInfos[0].BytesSent
		currentInfo.NetBytesRecv = netInfos[0].BytesRecv
	}
	currentInfo.ShotTime = time.Now()
	return &currentInfo, nil
}

func loadTopCPU() []ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	top5 := make([]ProcessInfo, 0, 5)
	for _, p := range processes {
		percent, err := p.CPUPercent()
		if err != nil {
			continue
		}
		minIndex := 0
		if len(top5) >= 5 {
			minCPU := top5[0].Percent
			for i := 1; i < len(top5); i++ {
				if top5[i].Percent < minCPU {
					minCPU = top5[i].Percent
					minIndex = i
				}
			}
			if percent < minCPU {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undifine"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undifine"
		}
		user, err := p.Username()
		if err != nil {
			user = "undifine"
		}
		if len(top5) == 5 {
			top5[minIndex] = ProcessInfo{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd}
		} else {
			top5 = append(top5, ProcessInfo{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd})
		}
	}
	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Percent > top5[j].Percent
	})

	return top5
}

func loadTopMem() []ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}
	top5 := make([]ProcessInfo, 0, 5)
	for _, p := range processes {
		stat, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		memItem := stat.RSS
		minIndex := 0
		if len(top5) >= 5 {
			min := top5[0].Memory
			for i := 1; i < len(top5); i++ {
				if top5[i].Memory < min {
					min = top5[i].Memory
					minIndex = i
				}
			}
			if memItem < min {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undifine"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undifine"
		}
		user, err := p.Username()
		if err != nil {
			user = "undifine"
		}
		percent, _ := p.MemoryPercent()
		if len(top5) == 5 {
			top5[minIndex] = ProcessInfo{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem}
		} else {
			top5 = append(top5, ProcessInfo{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem})
		}
	}

	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Memory > top5[j].Memory
	})
	return top5
}

func loadDiskInfo() []DiskInfo {
	var datas []DiskInfo

	// 使用gopsutil获取所有分区
	partitions, err := psutil.DISK.GetPartitions(true, false)
	if err != nil {
		return datas
	}

	var mounts []diskInfo
	var excludes = []string{
		"/mnt/cdrom",
		"/boot",
		"/boot/efi",
		"/dev",
		"/dev/shm",
		"/run/lock",
		"/run",
		"/run/shm",
		"/run/user",
		"/snap",
	}
	var excludesType = []string{
		"tmpfs",
		"overlay",
		"proc",
		"cgroup",
		"sysfs",
		"mqueue",
		"devpts",
	}

	// 过滤分区
	for _, partition := range partitions {
		if slices.Contains(excludesType, partition.Fstype) {
			continue
		}
		if slices.Contains(excludes, partition.Mountpoint) {
			continue
		}
		// 跳过挂载点路径太深的分区（>10个斜杠）
		if len(strings.Split(partition.Mountpoint, "/")) > 10 {
			continue
		}
		mounts = append(mounts, diskInfo{Type: partition.Fstype, Device: partition.Device, Mount: partition.Mountpoint})
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	wg.Add(len(mounts))
	for i := 0; i < len(mounts); i++ {
		go func(mount diskInfo) {
			defer wg.Done()

			var itemData DiskInfo
			itemData.Path = mount.Mount
			itemData.Type = mount.Type
			itemData.Device = mount.Device

			type diskResult struct {
				state *disk.UsageStat
				err   error
			}
			resultCh := make(chan diskResult, 1)

			go func() {
				state, err := psutil.DISK.GetUsage(mount.Mount, false)
				resultCh <- diskResult{state: state, err: err}
			}()

			select {
			case <-time.After(5 * time.Second):
				mu.Lock()
				datas = append(datas, itemData)
				mu.Unlock()
			case result := <-resultCh:
				if result.err != nil {
					mu.Lock()
					datas = append(datas, itemData)
					mu.Unlock()
					return
				}
				itemData.Total = result.state.Total
				itemData.Free = result.state.Free
				itemData.Used = result.state.Used
				itemData.UsedPercent = result.state.UsedPercent
				itemData.InodesTotal = result.state.InodesTotal
				itemData.InodesUsed = result.state.InodesUsed
				itemData.InodesFree = result.state.InodesFree
				itemData.InodesUsedPercent = result.state.InodesUsedPercent
				mu.Lock()
				datas = append(datas, itemData)
				mu.Unlock()
			}
		}(mounts[i])
	}
	wg.Wait()

	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Path < datas[j].Path
	})
	return datas
}

func loadOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "IPNotFound"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
