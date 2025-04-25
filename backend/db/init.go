package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
)

type initNode struct {
	Name      string    `bson:"name"`
	Address   string    `bson:"address"`
	CPUUsage  float64   `bson:"cpu_usage"`
	MemUsage  float64   `bson:"mem_usage"`
	DiskUsage float64   `bson:"disk_usage"`
	Online    bool      `bson:"online"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// 初始化本机为节点（仅首次）
func InitSelfNode() {
	hostInfo, _ := host.Info()
	fmt.Println("操作系统：", hostInfo.OS)
	hostName, _ := os.Hostname()
	addr := "127.0.0.1" // 可以换成真实公网地址或局域网地址

	// 检查是否已存在当前地址节点
	nodeCol := Mongo.Database("scriptdb").Collection("nodes")
	count, err := nodeCol.CountDocuments(context.TODO(), bson.M{"address": addr})
	if err != nil {
		fmt.Println("节点检测失败:", err)
		return
	}
	if count > 0 {
		fmt.Println("节点已存在，不再重复插入")
		return
	}

	// 获取资源状态
	cpuPercent, _ := cpu.Percent(0, false)
	memStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")

	node := initNode{
		Name:      hostName,
		Address:   addr,
		CPUUsage:  round(cpuPercent[0]),
		MemUsage:  round(memStat.UsedPercent),
		DiskUsage: round(diskStat.UsedPercent),
		Online:    true,
		UpdatedAt: time.Now(),
	}


	_, err = nodeCol.InsertOne(context.TODO(), node)
	if err != nil {
		fmt.Println("插入节点失败:", err)
	} else {
		fmt.Println("初始化节点成功:", hostName)
	}
}

// 保留两位小数
func round(val float64) float64 {
	return float64(int(val*100)) / 100
}
