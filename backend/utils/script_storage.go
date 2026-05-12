package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxScriptFileSize = 5 * 1024 * 1024

type ScriptFileMeta struct {
	OriginalFilename string
	FilePath         string
	SHA1             string
	Size             int64
	Language         string
}

func SaveUploadedScriptFile(userID primitive.ObjectID, scriptID primitive.ObjectID, fileHeader *multipart.FileHeader) (ScriptFileMeta, error) {
	if fileHeader == nil {
		return ScriptFileMeta{}, errors.New("未提供脚本文件")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return ScriptFileMeta{}, err
	}
	defer file.Close()

	content, err := readScriptContent(file)
	if err != nil {
		return ScriptFileMeta{}, err
	}

	return SaveScriptContent(userID, scriptID, fileHeader.Filename, content)
}

func SaveScriptContent(userID primitive.ObjectID, scriptID primitive.ObjectID, filename string, content []byte) (ScriptFileMeta, error) {
	if len(content) == 0 {
		return ScriptFileMeta{}, errors.New("脚本文件不能为空")
	}
	if len(content) > maxScriptFileSize {
		return ScriptFileMeta{}, fmt.Errorf("脚本文件不能超过 %d MB", maxScriptFileSize/1024/1024)
	}

	safeName := sanitizeFilename(filename)
	language, ok := detectScriptLanguage(safeName)
	if !ok {
		return ScriptFileMeta{}, errors.New("仅支持 .py、.sh、.js 脚本文件")
	}

	hash := sha1.Sum(content)
	sha1Text := hex.EncodeToString(hash[:])

	dir := filepath.Join(scriptStorageRoot(), userID.Hex(), scriptID.Hex())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ScriptFileMeta{}, err
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s_%s", sha1Text, safeName))
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return ScriptFileMeta{}, err
	}

	return ScriptFileMeta{
		OriginalFilename: safeName,
		FilePath:         filePath,
		SHA1:             sha1Text,
		Size:             int64(len(content)),
		Language:         language,
	}, nil
}

func RemoveScriptFile(filePath string) {
	if filePath == "" {
		return
	}
	_ = os.Remove(filePath)
}

func scriptStorageRoot() string {
	if root := os.Getenv("SCRIPT_STORAGE_DIR"); root != "" {
		return root
	}
	return filepath.Join("storage", "scripts")
}

func readScriptContent(reader io.Reader) ([]byte, error) {
	limited := io.LimitReader(reader, maxScriptFileSize+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(content) > maxScriptFileSize {
		return nil, fmt.Errorf("脚本文件不能超过 %d MB", maxScriptFileSize/1024/1024)
	}
	return content, nil
}

func sanitizeFilename(filename string) string {
	name := filepath.Base(filename)
	name = strings.TrimSpace(name)
	if name == "" || name == "." {
		name = "script.py"
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	name = re.ReplaceAllString(name, "_")
	return name
}

func detectScriptLanguage(filename string) (string, bool) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".py":
		return "python", true
	case ".sh":
		return "shell", true
	case ".js":
		return "node", true
	default:
		return "", false
	}
}
