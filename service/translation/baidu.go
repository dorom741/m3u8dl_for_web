package translation

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type BaiduTranslation struct {
	apikey    string
	secretKey string
}

type BaiduTranslationResult struct {
	ErrorCode   string `json:"error_code"`
	ErrorMsg    string `json:"error_msg"`
	From        string `json:"from"`
	To          string `json:"to"`
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
}

func NewBaiduTranslation(apikey string, secretKey string) *BaiduTranslation {
	translation := &BaiduTranslation{
		apikey:    apikey,
		secretKey: secretKey,
	}

	return translation
}

func (translation BaiduTranslation) generateSign(query, salt string) string {
	// 拼接字符串
	data := translation.apikey + query + salt + translation.secretKey
	// 计算 MD5
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (translation BaiduTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	// 随机数
	randValue := rand.New(rand.NewSource(time.Now().UnixNano()))
	salt := fmt.Sprintf("%d", randValue.Intn(32768)+32768)
	// 生成签名
	sign := translation.generateSign(text, salt)

	// 构建请求参数
	params := url.Values{}
	params.Set("q", url.QueryEscape(text)) // 进行 URL 编码
	params.Set("from", sourceLang)
	params.Set("to", targetLang)
	params.Set("appid", translation.apikey)
	params.Set("salt", salt)
	params.Set("sign", sign)

	// 发送请求
	resp, err := http.Get("https://fanyi-api.baidu.com/api/trans/vip/translate?" + params.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result BaiduTranslationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.ErrorCode != "" {
		return "", fmt.Errorf("baidu translation error with %s:%s ", result.ErrorCode, result.ErrorMsg)
	}

	if len(result.TransResult) > 0 {
		return result.TransResult[0].Dst, nil
	}

	return "", fmt.Errorf("baidu translation resul parsing failure:%s", body)
}
