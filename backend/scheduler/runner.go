package scheduler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jellyqwq/data-script-manager/backend/db"
)

func RunScript(scriptID primitive.ObjectID, scriptPath string) {
	cmd := exec.Command("python3", scriptPath)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	_ = cmd.Start()

	go captureLog(stdout, scriptID, "INFO")
	go captureLog(stderr, scriptID, "ERROR")

	_ = cmd.Wait()
}

func captureLog(pipe io.ReadCloser, scriptID primitive.ObjectID, level string) {
	scanner := bufio.NewScanner(pipe)
	col := db.Mongo.Database("scriptdb").Collection("logs")
	ctx := context.Background()

	for scanner.Scan() {
		line := scanner.Text()
		entry := bson.M{
			"script_id": scriptID,
			"timestamp": time.Now(),
			"level":     level,
			"message":   line,
		}
		_, err := col.InsertOne(ctx, entry)
		if err != nil {
			fmt.Println("日志写入失败：", err)
		}
	}
}
