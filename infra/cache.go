package infra

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
)

type FileCache struct {
	cacheDir string
}

// NewFileCache 创建新的文件缓存
func NewFileCache(dir string) (*FileCache, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return &FileCache{cacheDir: dir}, nil
}

// escapeFileName 转义文件名中的非法字符
func escapeFileName(name string) string {
	re := regexp.MustCompile(`[\/:*?"<>|]`)
	return re.ReplaceAllString(name, "_")
}

// Set 存储键值对到文件
func (fc *FileCache) Set(key string, value interface{}) error {
	escapedKey := escapeFileName(key)
	filePath := filepath.Join(fc.cacheDir, escapedKey)

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0o644)
}

// Get 从文件中获取值
func (fc *FileCache) Get(key string, value interface{}) error {
	escapedKey := escapeFileName(key)
	filePath := filepath.Join(fc.cacheDir, escapedKey)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

// ClearByPrefix 清除以特定前缀开头的文件
func (fc *FileCache) ClearByPrefix(prefix string) error {
	escapedPrefix := escapeFileName(prefix)
	files, err := os.ReadDir(fc.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if len(file.Name()) >= len(escapedPrefix) && file.Name()[:len(escapedPrefix)] == escapedPrefix {
			os.Remove(filepath.Join(fc.cacheDir, file.Name()))
		}
	}
	return nil
}
