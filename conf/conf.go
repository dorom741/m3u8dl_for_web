package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var ConfigInstance = NewConfig()

type Config struct {
	Server struct {
		Listen     string `yaml:"listen"`
		SavePath   string `yaml:"save_dir"`
		MaxWorker  int    `yaml:"maxWorker"`
		Dsn        string `yaml:"dsn"`
		StaticPath string `yaml:"staticPath"`
	} `yaml:"server"`
	Log struct {
		Path  string `yaml:"path"`
		Level string `yaml:"level"`
		LogNu int    `yaml:"log_Nu"`
	} `yaml:"log"`

	Groq struct {
		ApiKey    string `yaml:"apiKey"`
		CachePath string `yaml:"cachePath"`
	} `yaml:"Groq"`

	Translation struct {
		DeeplX *struct {
			Url    string `yaml:"url"`
			ApiKey string `yaml:"apiKey"`
		} `yaml:"deeplX"`
	} `yaml:"translation"`
}

func NewConfig() *Config {
	c := new(Config)
	c.Server.Listen = ":2045"
	c.Server.SavePath = "./download"
	c.Server.Dsn = "./data.db"
	c.Server.StaticPath = "./resource/static"
	c.Server.MaxWorker = 1
	c.Log.Level = "debug"
	c.Log.Path = "./log"
	c.Log.LogNu = 10
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
	// ConfMap = make(map[string]interface{})
	// ConfMap["Init.Port"] = config.Init.Port
	// ConfMap["log_Nu"] = config.Log.LogNu
	// ConfMap["save_dir"] = config.Init.SavePath
	// ConfMap["work_max"] = config.Init.WorkMax
	if !CheckWritePermission(ConfigInstance.Server.SavePath) {
		panic("save_dir: " + ConfigInstance.Server.SavePath + " can not write")
	}
	// 打印配置项的值
	// confjson, _ := json.Marshal(ConfMap)
	// fmt.Println("conf:", string(confjson))
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
