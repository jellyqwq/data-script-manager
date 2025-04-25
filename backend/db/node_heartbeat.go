package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StartNodeHeartbeat() {
	go func() {
		for {
			time.Sleep(10 * time.Second)

			hostName, _ := os.Hostname()
			addr := "127.0.0.1"

			cpuPercent, _ := cpu.Percent(0, false)
			memStat, _ := mem.VirtualMemory()
			diskStat, _ := disk.Usage("/")

			update := bson.M{
				"name":       hostName,
				"address":    addr,
				"cpu_usage":  round(cpuPercent[0]),
				"mem_usage":  round(memStat.UsedPercent),
				"disk_usage": round(diskStat.UsedPercent),
				"online":     true,
				"updated_at": time.Now(),
			}

			col := Mongo.Database("scriptdb").Collection("nodes")

			_, err := col.UpdateOne(context.TODO(),
				bson.M{"address": addr},
				bson.M{"$set": update},
				// 如果没有就插入
				// ⚠️ 这一步非常重要！
				// 确保没有 InitSelfNode() 也能创建记录
				&options.UpdateOptions{Upsert: newTruePtr()},
			)

			if err != nil {
				log.Println("节点状态写入失败:", err)
			} else {
				log.Println("节点心跳发送成功")
			}
		}
	}()
}

func newTruePtr() *bool {
	v := true
	return &v
}
