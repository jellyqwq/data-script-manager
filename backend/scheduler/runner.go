package scheduler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jellyqwq/data-script-manager/backend/db"
)

type RunOptions struct {
	ScheduleID   primitive.ObjectID
	ScriptID     primitive.ObjectID
	UserID       primitive.ObjectID
	ScriptPath   string
	EnvGroupID   *primitive.ObjectID
	EnvGroupName string
	EnvVars      map[string]string
}

func RunScript(options RunOptions) {
	runID := primitive.NewObjectID()
	startedAt := time.Now()

	scriptPath, err := filepath.Abs(options.ScriptPath)
	if err != nil {
		insertTaskRun(runID, options, "", startedAt)
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}
	options.ScriptPath = scriptPath

	stateDir, err := ensureStateDir(options)
	insertTaskRun(runID, options, stateDir, startedAt)
	if err != nil {
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	command, args, err := buildScriptCommand(options)
	if err != nil {
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}
	writeRunLog(runID, options, "INFO", "使用运行时: "+command)

	if err := preflightScript(runID, options, command, stateDir); err != nil {
		message := "环境检查失败: " + err.Error()
		writeRunLog(runID, options, "ERROR", message)
		finishTaskRun(runID, "failed", nil, message, startedAt)
		return
	}
	writeRunLog(runID, options, "INFO", "环境检查通过")

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = stateDir
	cmd.Env = buildProcessEnv(options, runID, stateDir)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}

	if err := cmd.Start(); err != nil {
		finishTaskRun(runID, "failed", nil, err.Error(), startedAt)
		return
	}

	go captureLog(stdout, runID, options, "INFO")
	go captureLog(stderr, runID, options, "ERROR")

	err = cmd.Wait()
	exitCode := cmd.ProcessState.ExitCode()
	if ctx.Err() == context.DeadlineExceeded {
		finishTaskRun(runID, "timeout", &exitCode, "执行超时", startedAt)
		return
	}
	if err != nil {
		finishTaskRun(runID, "failed", &exitCode, err.Error(), startedAt)
		return
	}

	finishTaskRun(runID, "success", &exitCode, "", startedAt)
}

func buildScriptCommand(options RunOptions) (string, []string, error) {
	switch strings.ToLower(filepath.Ext(options.ScriptPath)) {
	case ".py":
		return pythonRuntime(options), []string{options.ScriptPath}, nil
	case ".sh":
		return runtimeBinary(options, "bash", "DSM_BASH_BIN", "SCRIPT_BASH_BIN", "BASH_BIN"), []string{options.ScriptPath}, nil
	case ".js":
		return runtimeBinary(options, "node", "DSM_NODE_BIN", "SCRIPT_NODE_BIN", "NODE_BIN"), []string{options.ScriptPath}, nil
	default:
		return "", nil, fmt.Errorf("unsupported script extension: %s", filepath.Ext(options.ScriptPath))
	}
}

func runtimeBinary(options RunOptions, fallback string, envNames ...string) string {
	for _, name := range envNames {
		if value := strings.TrimSpace(options.EnvVars[name]); value != "" {
			return value
		}
	}
	for _, name := range envNames {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return fallback
}

func ensureStateDir(options RunOptions) (string, error) {
	root := strings.TrimSpace(os.Getenv("SCRIPT_STATE_DIR"))
	if root == "" {
		root = filepath.Join("storage", "state")
	}

	envGroupID := "default"
	if options.EnvGroupID != nil {
		envGroupID = options.EnvGroupID.Hex()
	}

	stateDir := filepath.Join(root, options.UserID.Hex(), options.ScriptID.Hex(), envGroupID)
	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(absStateDir, 0750); err != nil {
		return absStateDir, err
	}
	return absStateDir, nil
}

func captureLog(pipe io.ReadCloser, runID primitive.ObjectID, options RunOptions, level string) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		writeRunLog(runID, options, level, scanner.Text())
	}
}

func writeRunLog(runID primitive.ObjectID, options RunOptions, level string, message string) {
	col := db.Mongo.Database("scriptdb").Collection("logs")
	entry := bson.M{
		"run_id":      runID,
		"schedule_id": options.ScheduleID,
		"script_id":   options.ScriptID,
		"timestamp":   time.Now(),
		"level":       level,
		"message":     message,
		"user_id":     options.UserID,
	}
	_, err := col.InsertOne(context.Background(), entry)
	if err != nil {
		fmt.Println("日志写入失败：", err)
	}
}

func buildProcessEnv(options RunOptions, runID primitive.ObjectID, stateDir string) []string {
	env := os.Environ()
	for key, value := range options.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	env = append(env,
		"DSM_RUN_ID="+runID.Hex(),
		"DSM_SCHEDULE_ID="+options.ScheduleID.Hex(),
		"DSM_SCRIPT_ID="+options.ScriptID.Hex(),
		"DSM_SCRIPT_PATH="+options.ScriptPath,
		"DSM_SCRIPT_DIR="+filepath.Dir(options.ScriptPath),
		"DSM_STATE_DIR="+stateDir,
		"DSM_WORKDIR="+stateDir,
	)
	if options.EnvGroupID != nil {
		env = append(env, "DSM_ENV_GROUP_ID="+options.EnvGroupID.Hex())
	}
	if options.EnvGroupName != "" {
		env = append(env, "DSM_ENV_GROUP_NAME="+options.EnvGroupName)
	}
	return env
}

func insertTaskRun(runID primitive.ObjectID, options RunOptions, stateDir string, startedAt time.Time) {
	col := db.Mongo.Database("scriptdb").Collection("task_runs")
	doc := bson.M{
		"_id":            runID,
		"schedule_id":    options.ScheduleID,
		"script_id":      options.ScriptID,
		"user_id":        options.UserID,
		"env_group_id":   options.EnvGroupID,
		"env_group_name": options.EnvGroupName,
		"status":         "running",
		"started_at":     startedAt,
	}
	if stateDir != "" {
		doc["state_dir"] = stateDir
	}

	_, err := col.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("执行记录写入失败：", err)
	}
}

func finishTaskRun(runID primitive.ObjectID, status string, exitCode *int, errorMessage string, startedAt time.Time) {
	finishedAt := time.Now()
	update := bson.M{
		"status":      status,
		"finished_at": finishedAt,
		"duration_ms": finishedAt.Sub(startedAt).Milliseconds(),
	}
	if exitCode != nil {
		update["exit_code"] = *exitCode
	}
	if errorMessage != "" {
		update["error_message"] = errorMessage
	}

	col := db.Mongo.Database("scriptdb").Collection("task_runs")
	_, err := col.UpdateOne(context.TODO(), bson.M{"_id": runID}, bson.M{"$set": update})
	if err != nil {
		fmt.Println("执行记录更新失败：", err)
	}
}
