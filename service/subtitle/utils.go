package subtitle

import (
	"fmt"
	"regexp"
	"sort"
)

func FindRepeatedWords(text string) []string {
	minLen := 2
	wordMap := make(map[string]int)
	var repeatedWords []string
	runes := []rune(text)
	runesLen := len(runes)

	// 遍历所有可能的起始位置
	for i := 0; i < runesLen; i++ {
		for j := i + minLen; j < runesLen; j++ {
			word := runes[i:j]
			wordMap[string(word)]++
		}
	}

	// 找出重复出现的词
	for word, count := range wordMap {
		if count > 1 {
			pattern := fmt.Sprintf("(%s){2,}", regexp.QuoteMeta(word))
			re := regexp.MustCompile(pattern)
			match := re.FindString(text)
			if match != "" {
				repeatedWords = append(repeatedWords, word)
			}
		}
	}

	sort.Slice(repeatedWords, func(i, j int) bool {
		return len(repeatedWords[i]) > len(repeatedWords[j])
	})
	return repeatedWords
}

func ReplaceRepeatedWords(text string) string {
	symbolRe := regexp.MustCompile(`[，、]`)
	newText := symbolRe.ReplaceAllString(text, ",")
	result := FindRepeatedWords(newText)
	for _, s := range result {
		pattern := fmt.Sprintf("(%s)+", regexp.QuoteMeta(s))
		re := regexp.MustCompile(pattern)
		newText = re.ReplaceAllString(newText, s)
	}

	return newText
}
