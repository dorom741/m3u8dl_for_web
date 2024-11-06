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
	var (
		sub = &subtitleSub{}
		s   = astisub.NewSubtitles()
	)

	s.Metadata = &astisub.Metadata{}

	s.Styles["main"] = &astisub.Style{ID: "main", InlineStyle: sub.newStyleAttributes()}
	secondInlineStyle := *s.Styles["main"].InlineStyle
	secondInlineStyle.SSAFontSize = floatPointer(10)
	secondInlineStyle.SSAMarginVertical = intPointer(5)

	s.Styles["second"] = &astisub.Style{ID: "second", InlineStyle: &secondInlineStyle}

	sub.subtitles = s
	return sub
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

func (sub *subtitleSub) newStyleAttributes() *astisub.StyleAttributes {

	return &astisub.StyleAttributes{
		SSAAlignment:       intPointer(2),
		SSAAlphaLevel:      floatPointer(1.0),
		SSAAngle:           floatPointer(0.0),
		SSABackColour:      &astisub.Color{Alpha: 100, Red: 0, Green: 0, Blue: 0}, // #64000000,
		SSABold:            boolPointer(true),
		SSABorderStyle:     intPointer(1),
		SSAEffect:          "",
		SSAEncoding:        intPointer(1),
		SSAFontName:        "Noto Sans",
		SSAFontSize:        floatPointer(15.0),
		SSAItalic:          boolPointer(false),
		SSALayer:           intPointer(0),
		SSAMarginLeft:      intPointer(10),
		SSAMarginRight:     intPointer(10),
		SSAMarginVertical:  intPointer(20),
		SSAMarked:          boolPointer(false),
		SSAOutline:         floatPointer(2.0),
		SSAOutlineColour:   &astisub.Color{Alpha: 0, Red: 0, Green: 0, Blue: 0},         // #00000000,
		SSAPrimaryColour:   &astisub.Color{Alpha: 255, Red: 255, Green: 255, Blue: 255}, // #00FFFFFF,
		SSAScaleX:          floatPointer(100.0),
		SSAScaleY:          floatPointer(100.0),
		SSASecondaryColour: &astisub.Color{Alpha: 255, Red: 0, Green: 0, Blue: 0}, // #000000FF,
		SSAShadow:          floatPointer(0.0),
		SSASpacing:         floatPointer(0.0),
		SSAStrikeout:       boolPointer(false),
		SSAUnderline:       boolPointer(false),
	}
}

func intPointer(i int) *int {
	return &i
}

func floatPointer(f float64) *float64 {
	return &f
}

func boolPointer(b bool) *bool {
	return &b
}
