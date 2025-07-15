package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func GetNowTime() string {
	return FormatTime(time.Now())
}

func WriteToFile(filePath string, data any) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("写入 JSON 失败: %v", err)
	}

	return nil
}
