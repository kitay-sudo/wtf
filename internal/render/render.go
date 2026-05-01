package render

import (
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"golang.org/x/term"
)

// Markdown рендерит markdown-текст для терминала через glamour. Кастомная
// тёмная тема в стиле wtf: жёлтые заголовки/жирный, голубой инлайн-код,
// серый dim для второстепенного текста. Word-wrap делает glamour сам, ширину
// мы задаём (terminalWidth - reservedIndent) — финальный отступ под префикс
// [HH:MM:SS] ★ накладывается уже снаружи в ui.FinalBlock.
//
// При NO_COLOR / non-TTY возвращаем исходный текст без обработки.
func Markdown(s string, width int) string {
	if !isTTY() || os.Getenv("NO_COLOR") != "" {
		return s
	}
	r := getRenderer(width)
	if r == nil {
		return s
	}
	out, err := r.Render(s)
	if err != nil {
		return s
	}
	// Glamour добавляет ведущий/висячий перевод строки — снимаем, FinalBlock
	// сам управляет вертикальными отбивками.
	return strings.Trim(out, "\n")
}

var (
	rendererMu    sync.Mutex
	cachedR       *glamour.TermRenderer
	cachedRWidth  int
)

func getRenderer(width int) *glamour.TermRenderer {
	if width < 20 {
		width = 80
	}
	rendererMu.Lock()
	defer rendererMu.Unlock()
	if cachedR != nil && cachedRWidth == width {
		return cachedR
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(wtfStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	cachedR = r
	cachedRWidth = width
	return r
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// wtfStyle — кастомная тёмная тема под палитру wtf (жёлтый/серый/голубой).
// Базируется на DarkStyleConfig из glamour, но с нашими цветами.
func ptrStr(s string) *string { return &s }
func ptrUint(u uint) *uint    { return &u }
func ptrBool(b bool) *bool    { return &b }

var wtfStyle = ansi.StyleConfig{
	Document: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockPrefix: "",
			BlockSuffix: "",
			Color:       ptrStr("252"),
		},
		Margin: ptrUint(0),
	},
	BlockQuote: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{Color: ptrStr("244"), Italic: ptrBool(true)},
		Indent:         ptrUint(1),
		IndentToken:    ptrStr("│ "),
	},
	Paragraph: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{},
	},
	List: ansi.StyleList{
		LevelIndent: 2,
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{},
		},
	},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			BlockSuffix: "\n",
			Color:       ptrStr("220"),
			Bold:        ptrBool(true),
		},
	},
	H1: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Suffix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	H2: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	H3: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	H4: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	H5: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	H6: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Prefix: "",
			Color:  ptrStr("220"),
			Bold:   ptrBool(true),
		},
	},
	Strikethrough: ansi.StylePrimitive{CrossedOut: ptrBool(true)},
	Emph:          ansi.StylePrimitive{Italic: ptrBool(true)},
	Strong:        ansi.StylePrimitive{Bold: ptrBool(true), Color: ptrStr("220")},
	HorizontalRule: ansi.StylePrimitive{
		Color:  ptrStr("240"),
		Format: "\n--------\n",
	},
	Item: ansi.StylePrimitive{BlockPrefix: ""},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: "",
		Color:       ptrStr("220"),
	},
	Task: ansi.StyleTask{
		StylePrimitive: ansi.StylePrimitive{},
		Ticked:         "[✓] ",
		Unticked:       "[ ] ",
	},
	Link: ansi.StylePrimitive{
		Color:     ptrStr("39"),
		Underline: ptrBool(true),
	},
	LinkText: ansi.StylePrimitive{
		Color: ptrStr("39"),
		Bold:  ptrBool(true),
	},
	Image: ansi.StylePrimitive{
		Color:     ptrStr("39"),
		Underline: ptrBool(true),
	},
	ImageText: ansi.StylePrimitive{
		Color:  ptrStr("39"),
		Format: "Image: {{.text}} →",
	},
	Code: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{
			Color: ptrStr("51"),
		},
	},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: ptrStr("220"),
			},
			Margin: ptrUint(2),
		},
		Chroma: &ansi.Chroma{
			Text:                ansi.StylePrimitive{Color: ptrStr("252")},
			Error:               ansi.StylePrimitive{Color: ptrStr("203")},
			Comment:             ansi.StylePrimitive{Color: ptrStr("244")},
			CommentPreproc:      ansi.StylePrimitive{Color: ptrStr("207")},
			Keyword:             ansi.StylePrimitive{Color: ptrStr("204")},
			KeywordReserved:     ansi.StylePrimitive{Color: ptrStr("204")},
			KeywordNamespace:    ansi.StylePrimitive{Color: ptrStr("204")},
			KeywordType:         ansi.StylePrimitive{Color: ptrStr("204")},
			Operator:            ansi.StylePrimitive{Color: ptrStr("220")},
			Punctuation:         ansi.StylePrimitive{Color: ptrStr("217")},
			Name:                ansi.StylePrimitive{Color: ptrStr("252")},
			NameBuiltin:         ansi.StylePrimitive{Color: ptrStr("220")},
			NameTag:             ansi.StylePrimitive{Color: ptrStr("220")},
			NameAttribute:       ansi.StylePrimitive{Color: ptrStr("220")},
			NameClass:           ansi.StylePrimitive{Color: ptrStr("220")},
			NameConstant:        ansi.StylePrimitive{Color: ptrStr("220")},
			NameDecorator:       ansi.StylePrimitive{Color: ptrStr("220")},
			NameFunction:        ansi.StylePrimitive{Color: ptrStr("220")},
			LiteralNumber:       ansi.StylePrimitive{Color: ptrStr("141")},
			LiteralString:       ansi.StylePrimitive{Color: ptrStr("114")},
			LiteralStringEscape: ansi.StylePrimitive{Color: ptrStr("220")},
			GenericDeleted:      ansi.StylePrimitive{Color: ptrStr("203")},
			GenericEmph:         ansi.StylePrimitive{Italic: ptrBool(true)},
			GenericInserted:     ansi.StylePrimitive{Color: ptrStr("114")},
			GenericStrong:       ansi.StylePrimitive{Bold: ptrBool(true)},
			GenericSubheading:   ansi.StylePrimitive{Color: ptrStr("251")},
			Background:          ansi.StylePrimitive{BackgroundColor: ptrStr("234")},
		},
	},
	Table: ansi.StyleTable{
		StyleBlock: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{}},
		CenterSeparator: ptrStr("┼"),
		ColumnSeparator: ptrStr("│"),
		RowSeparator:    ptrStr("─"),
	},
	DefinitionDescription: ansi.StylePrimitive{BlockPrefix: "\n  "},
}
