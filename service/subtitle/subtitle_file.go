package subtitle

import (
	"io"
	"time"

	"github.com/asticode/go-astisub"
)

type subtitleSub struct {
	subtitles *astisub.Subtitles
}

func NewSubtitleSub() *subtitleSub {
	s := astisub.NewSubtitles()
	s.Metadata = &astisub.Metadata{}

	var (
		mainFontSize   float64 = 20.0
		secondFontSize float64 = 10.0
		secondMarginV  int     = 10.0
	)
	s.Styles["main"] = &astisub.Style{ID: "main", InlineStyle: &astisub.StyleAttributes{SSAFontSize: &mainFontSize, SSAPrimaryColour: astisub.ColorWhite}}
	secondInlineStyle := *s.Styles["main"].InlineStyle
	secondInlineStyle.SSAFontSize = &secondFontSize
	secondInlineStyle.SSAMarginVertical = &secondMarginV

	s.Styles["second"] = &astisub.Style{ID: "second", InlineStyle: &secondInlineStyle}

	return &subtitleSub{
		subtitles: s,
	}
}

func (sub *subtitleSub) Metadata() *astisub.Metadata {
	return sub.subtitles.Metadata
}

func (sub *subtitleSub) AddLine(index int, startTimeStamp, endTimeStamp float64, mainText, secondText string) {
	item := &astisub.Item{
		Index:   index,
		StartAt: time.Duration(startTimeStamp * float64(time.Second)),
		EndAt:   time.Duration(endTimeStamp * float64(time.Second)),
		Style:   &astisub.Style{ID: "main"},
		Lines:   []astisub.Line{{Items: []astisub.LineItem{{Text: mainText}}}},
	}
	sub.subtitles.Items = append(sub.subtitles.Items, item)

	if secondText != "" {
		secondItem := &astisub.Item{
			Index:   index,
			StartAt: time.Duration(startTimeStamp * float64(time.Second)),
			EndAt:   time.Duration(endTimeStamp * float64(time.Second)),
			Style:   &astisub.Style{ID: "second"},
			Lines:   []astisub.Line{{Items: []astisub.LineItem{{Text: secondText}}}},
		}
		sub.subtitles.Items = append(sub.subtitles.Items, secondItem)

	}
}

func (sub *subtitleSub) WriteToFile(o io.Writer) error {
	return sub.subtitles.WriteToSSA(o)
}
