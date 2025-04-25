package scheduler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jellyqwq/data-script-manager/backend/db"
	"github.com/jellyqwq/data-script-manager/backend/models"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	c      *cron.Cron              = cron.New()
	jobMap map[string]cron.EntryID = make(map[string]cron.EntryID)
)

// 启动调度器
func StartScheduler() {
	log.Println("[调度器] 启动调度器...")
	loadSchedules()
	c.Start()
	log.Println("[调度器] 所有启用任务已加载完成")
}

// 加载所有任务
func loadSchedules() {
	col := db.Mongo.Database("scriptdb").Collection("schedules")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		log.Println("[调度器] 任务查询失败：", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var sched models.ScheduleItem
		if err := cursor.Decode(&sched); err != nil {
			log.Println("[调度器] 任务解码失败：", err)
			continue
		}
		log.Printf("[调度器] 加载任务 ID=%s CRON=%s 启用=%v\n", sched.ID.Hex(), sched.Cron, sched.Enabled)
		ReloadSchedule(sched.ID.Hex())
	}
}

// 重新加载任务
func ReloadSchedule(id string) {
	// 移除旧任务
	if old, ok := jobMap[id]; ok {
		c.Remove(old)
		delete(jobMap, id)
		log.Printf("[调度器] 移除旧任务 ID=%s\n", id)
	}

	objID, _ := primitive.ObjectIDFromHex(id)
	col := db.Mongo.Database("scriptdb").Collection("schedules")
	var sched models.ScheduleItem
	err := col.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&sched)
	if err != nil {
		log.Println("[调度器] 查询调度任务失败：", err)
		return
	}

	if !sched.Enabled {
		log.Printf("[调度器] 任务 ID=%s 已禁用，跳过注册\n", id)
		return
	}

	registerSchedule(sched)
}

// 注册任务
func registerSchedule(sched models.ScheduleItem) {
	scriptCol := db.Mongo.Database("scriptdb").Collection("scripts")
	var script struct {
		Content string `bson:"content"`
	}
	log.Println(sched)
	err := scriptCol.FindOne(context.TODO(), bson.M{"_id": sched.ScriptID}).Decode(&script)
	if err != nil {
		log.Printf("[调度器] 无法找到脚本 ID=%s\n", sched.ScriptID.Hex())
		return
	}

	// 保存为本地 .py 文件
	scriptPath := fmt.Sprintf("scripts/%s.py", sched.ID.Hex())

	// ✅ 检查 scripts 目录是否存在，不存在就创建
	if err := os.MkdirAll("scripts", os.ModePerm); err != nil {
		log.Printf("[调度器] 创建 scripts 目录失败: %v\n", err)
		return
	}

	err = os.WriteFile(scriptPath, []byte(script.Content), 0644)
	if err != nil {
		log.Printf("[调度器] 写入脚本文件失败 ID=%s：%v\n", sched.ID.Hex(), err)
		return
	}

	entryID, err := c.AddFunc(sched.Cron, func() {
		log.Printf("[调度器] 执行任务 ID=%s 路径=%s\n", sched.ID.Hex(), scriptPath)
		RunScript(sched.ScriptID, sched.UserID, scriptPath)
	})
	if err != nil {
		log.Printf("[调度器] 注册任务失败 ID=%s：%v\n", sched.ID.Hex(), err)
		return
	}

	jobMap[sched.ID.Hex()] = entryID
	log.Printf("[调度器] 成功注册任务 ID=%s\n", sched.ID.Hex())
}

func RemoveSchedule(id string) {
	if entryID, ok := jobMap[id]; ok {
		c.Remove(entryID)
		delete(jobMap, id)
		log.Printf("[调度器] 已移除任务 ID=%s\n", id)
	}
}

// 全量重载（备用）
func ReloadAll() {
	log.Println("[调度器] 全量重载所有任务")
	for id := range jobMap {
		ReloadSchedule(id)
	}
}
