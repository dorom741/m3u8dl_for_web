package conf

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"m3u8dl_for_web/model/aggregate"

	"github.com/sirupsen/logrus"
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
	Blacklist            []string                `yaml:"blacklist"`
	BlacklistFilePath    string                  `yaml:"blacklistFilePath"`
	LockFilePath         string                  `yaml:"lockFilePath"`
}

func (subtitleConfig *SubtitleConfig) saveBlacklistFile(blacklist []string) error {
	seen := make(map[string]bool)
	uniqueSlice := []string{}

	for _, value := range blacklist {
		if !seen[value] {
			uniqueSlice = append(uniqueSlice, value)
			seen[value] = true
		}
	}

	if subtitleConfig.BlacklistFilePath == "" {
		return nil
	}
	return os.WriteFile(subtitleConfig.BlacklistFilePath, []byte(strings.Join(uniqueSlice, "\n")), os.ModePerm)
}

func (subtitleConfig *SubtitleConfig) WriteLockFile(data string) error {
	return os.WriteFile(subtitleConfig.LockFilePath, []byte(data), os.ModePerm)
}

func (subtitleConfig *SubtitleConfig) ReadLastLockFile() string {
	lockFileData, err := os.ReadFile(subtitleConfig.LockFilePath)
	if err == nil {
		return string(lockFileData)
	}

	return ""
}

func (subtitleConfig *SubtitleConfig) RemoveLastLockFile() error {
	return os.Remove(subtitleConfig.LockFilePath)
}

func (subtitleConfig *SubtitleConfig) GenerateBlacklistJudgement() (func(filePath string) bool, error) {
	if lastLockFile := subtitleConfig.ReadLastLockFile(); lastLockFile != "" {
		subtitleConfig.Blacklist = append(subtitleConfig.Blacklist, lastLockFile)
	}

	if subtitleConfig.BlacklistFilePath != "" {
		blacklistFileBytes, err := os.ReadFile(subtitleConfig.BlacklistFilePath)
		if err != nil {
			return nil, err
		}

		blacklistFromFile := strings.Split(string(blacklistFileBytes), "\n")
		subtitleConfig.Blacklist = append(subtitleConfig.Blacklist, blacklistFromFile...)
	}

	if len(subtitleConfig.Blacklist) == 0 {
		return func(filePath string) bool { return false }, nil
	}

	if err := subtitleConfig.saveBlacklistFile(subtitleConfig.Blacklist); err != nil {
		return nil, err
	}

	patten := fmt.Sprintf(".*(%s).*", strings.Join(subtitleConfig.Blacklist, "|"))
	logrus.Debugf("patten %s", patten)
	reMulti := regexp.MustCompile(patten)

	return func(filePath string) bool { return reMulti.MatchString(filePath) }, nil
}
