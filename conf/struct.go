package conf

import (
	"m3u8dl_for_web/model/aggregate"
)

type ServerConfig struct {
	Listen     string `yaml:"listen"`
	Dsn        string `yaml:"dsn"`
	StaticPath string `yaml:"staticPath"`

	DownlaodMaxWorker int64  `yaml:"downlaodMaxWorker"`
	SavePath          string `yaml:"saveDir"`
	CacheDir          string `yaml:"cacheDir"`
	HttpClientProxy   string `yaml:"httpClientProxy"`
}

type LogConfig struct {
	Path    string `yaml:"path"`
	Level   string `yaml:"level"`
	MaxSize int    `yaml:"maxSize"`
	MaxAge  int    `yaml:"maxAge"`
}

type GroqConfig struct {
	ApiKey string `yaml:"apiKey"`
}

type TranslationConfig struct {
	DeeplX *struct {
		Url    string `yaml:"url"`
		ApiKey string `yaml:"apiKey"`
	} `yaml:"deeplX"`
}

type SubtitleConfig struct {
	DirPath              string                  `yaml:"dirPath"`
	Pattern              string                  `yaml:"pattern"`
	Watch                bool                    `yaml:"watch"`
	SubtitleInput        aggregate.SubtitleInput `yaml:"subtitleInput"`
	FixMissTranslate     bool                    `yaml:"fixMissTranslate"`
	JustFixMissTranslate bool                    `yaml:"justFixMissTranslate"`
}
