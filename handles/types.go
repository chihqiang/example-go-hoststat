package handles

import "time"

const (
	DateTimeLayout = "2006-01-02 15:04:05" // or use time.DateTime while go version >= 1.20
)

type BaseInfo struct {
	Hostname        string  `json:"hostname"`        // 主机名
	OS              string  `json:"os"`              // 操作系统类型
	Platform        string  `json:"platform"`        // 平台名称
	PlatformFamily  string  `json:"platformFamily"`  // 平台家族
	PlatformVersion string  `json:"platformVersion"` // 平台版本
	PrettyDistro    string  `json:"prettyDistro"`    // 发行版名称（美化）
	KernelArch      string  `json:"kernelArch"`      // 内核架构
	KernelVersion   string  `json:"kernelVersion"`   // 内核版本
	IPV4Addr        string  `json:"ipV4Addr"`        // IPv4地址
	SystemProxy     string  `json:"systemProxy"`
	CPUCores        int     `json:"cpuCores"`        // CPU物理核心数
	CPULogicalCores int     `json:"cpuLogicalCores"` // CPU逻辑核心数
	CPUModelName    string  `json:"cpuModelName"`    // CPU型号名称
	CPUMhz          float64 `json:"cpuMhz"`          // CPU主频（MHz）

	CurrentInfo *CurrentInfo `json:"currentInfo"`
}

type CurrentInfo struct {
	Uptime          uint64 `json:"uptime"`
	TimeSinceUptime string `json:"timeSinceUptime"`
	Procs           uint64 `json:"procs"`

	CPUPercent         []float64 `json:"cpuPercent"`
	CPUUsedPercent     float64   `json:"cpuUsedPercent"`
	CPUUsed            float64   `json:"cpuUsed"`
	CPUTotal           int       `json:"cpuTotal"`
	CPUDetailedPercent []float64 `json:"cpuDetailedPercent"`

	Load1            float64 `json:"load1"`
	Load5            float64 `json:"load5"`
	Load15           float64 `json:"load15"`
	LoadUsagePercent float64 `json:"loadUsagePercent"`

	MemoryTotal       uint64  `json:"memoryTotal"`
	MemoryUsed        uint64  `json:"memoryUsed"`
	MemoryFree        uint64  `json:"memoryFree"`
	MemoryShard       uint64  `json:"memoryShard"`
	MemoryCache       uint64  `json:"memoryCache"`
	MemoryAvailable   uint64  `json:"memoryAvailable"`
	MemoryUsedPercent float64 `json:"memoryUsedPercent"`

	SwapMemoryTotal       uint64  `json:"swapMemoryTotal"`
	SwapMemoryAvailable   uint64  `json:"swapMemoryAvailable"`
	SwapMemoryUsed        uint64  `json:"swapMemoryUsed"`
	SwapMemoryUsedPercent float64 `json:"swapMemoryUsedPercent"`

	DiskData []DiskInfo `json:"diskData"`

	IOReadBytes  uint64 `json:"ioReadBytes"`
	IOWriteBytes uint64 `json:"ioWriteBytes"`
	IOCount      uint64 `json:"ioCount"`
	IOReadTime   uint64 `json:"ioReadTime"`
	IOWriteTime  uint64 `json:"ioWriteTime"`

	NetBytesSent uint64 `json:"netBytesSent"`
	NetBytesRecv uint64 `json:"netBytesRecv"`

	ShotTime time.Time `json:"shotTime"`
}

type DiskInfo struct {
	Path        string  `json:"path"`
	Type        string  `json:"type"`
	Device      string  `json:"device"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`

	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}
type diskInfo struct {
	Type   string
	Mount  string
	Device string
}
type ProcessInfo struct {
	Name    string  `json:"name"`
	Pid     int32   `json:"pid"`
	Percent float64 `json:"percent"`
	Memory  uint64  `json:"memory"`
	Cmd     string  `json:"cmd"`
	User    string  `json:"user"`
}
