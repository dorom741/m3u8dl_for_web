package model

import (
	"github.com/orestonce/m3u8d"
	"path/filepath"
)

type M3u8dlInput struct {
	Name    string              `json:"name"`
	URL     string              `json:"url"`
	SaveDir string              `json:"saveDir"`
	Headers map[string][]string `json:"headers"`
}

func (m3u8dlInput *M3u8dlInput) GetSavePath() string {
	return filepath.Join(m3u8dlInput.SaveDir, m3u8dlInput.Name+".mp4")
}

func (m3u8dlInput *M3u8dlInput) ToStartDownloadReq() m3u8d.StartDownload_Req {
	return m3u8d.StartDownload_Req{
		M3u8Url:                  m3u8dlInput.URL,
		Insecure:                 true,
		SaveDir:                  m3u8dlInput.SaveDir,
		FileName:                 m3u8dlInput.Name,
		SkipTsExpr:               "",
		SetProxy:                 "",
		HeaderMap:                m3u8dlInput.Headers,
		SkipRemoveTs:             false,
		ProgressBarShow:          false,
		ThreadCount:              4,
		SkipCacheCheck:           false,
		SkipMergeTs:              false,
		Skip_EXT_X_DISCONTINUITY: false,
		DebugLog:                 false,
	}
}

type M3u8dlOutput struct {
	SaveDir string `json:"saveDir"`
}
