package conf

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ConfigInstance = NewConfig()

type Config struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`

	Groq GroqConfig `yaml:"groq"`

	Translation TranslationConfig `yaml:"translation"`

	LocalWhisperModels map[string]string `yaml:"localWhisperModels"`
	Subtitle           *SubtitleConfig   `yaml:"subtitle"`
	SherpaConfig       *SherpaConfigType `yaml:"sherpa"`
}

func (conf *Config) GetAbsSavePath() string {
	fullSavePath, err := filepath.Abs(conf.Server.SavePath)
	if err != nil {
		return os.TempDir()
	}

	return fullSavePath
}

func (conf *Config) GetAbsCachePath() string {
	fullCachePath, err := filepath.Abs(conf.Server.CacheDir)
	if err != nil {
		return os.TempDir()
	}

	return fullCachePath
}

func NewConfig() *Config {
	c := new(Config)
	c.Server.Listen = ":2045"
	c.Server.SavePath = "./download"
	c.Server.Dsn = "./data.db"
	c.Server.StaticPath = "./resource/static"
	c.Server.DownlaodMaxWorker = 1
	c.Log.Level = "debug"
	c.Log.Path = "./log"
	c.Log.MaxSize = 10
	c.Log.MaxAge = 30
	return c
}

// var ConfMap map[string]interface{}

func InitConf(configFilePath string) {
	// 读取YAML配置文件内容
	yamlFile, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("无法读取YAML文件：%v", err)
		return
	}

	// 解析YAML配置文件
	err = yaml.Unmarshal(yamlFile, ConfigInstance)
	if err != nil {
		log.Fatalf("无法解析YAML文件：%v", err)
	}

	if !CheckWritePermission(ConfigInstance.Server.SavePath) {
		panic("save_dir: " + ConfigInstance.Server.SavePath + " can not write")
	}
}

// 检查文件夹是否可写
func CheckWritePermission(dirPath string) bool {
	tmpFile, err := os.CreateTemp(dirPath, "test*")
	if err != nil {
		return false
	}
	defer os.Remove(tmpFile.Name()) // 确保临时文件被删除
	defer tmpFile.Close()
	return true
}
