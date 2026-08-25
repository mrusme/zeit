package out

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-isatty"
)

type OutputType int

const (
	Plain OutputType = iota
	Ok
	Error
	Info
	Start
	Switch
	Resume
	End
	Pause
	Erase
)

type outputPrefix struct {
	Char  string
	Color ansi.BasicColor
}

var OutputPrefixes = []outputPrefix{
	{
		Char:  "",
		Color: lipgloss.White,
	},
	{
		Char:  "● ",
		Color: lipgloss.Green,
	},
	{
		Char:  "▲ ",
		Color: lipgloss.Red,
	},
	{
		Char:  "◆ ",
		Color: lipgloss.BrightBlack,
	},
	{
		Char:  "▶ ",
		Color: lipgloss.Cyan,
	},
	{
		Char:  "▰ ",
		Color: lipgloss.Cyan,
	},
	{
		Char:  "► ",
		Color: lipgloss.Cyan,
	},
	{
		Char:  "■ ",
		Color: lipgloss.Magenta,
	},
	{
		Char:  "▮ ",
		Color: lipgloss.Magenta,
	},
	{
		Char:  "◀ ",
		Color: lipgloss.BrightRed,
	},
}

type Style struct {
	FG color.Color
	BG color.Color
	PX int
}

type OutputColor int

const (
	ColorNever = iota
	ColorAuto
	ColorAlways
)

const (
	ColorRed           = lipgloss.Red
	ColorYellow        = lipgloss.Yellow
	ColorGreen         = lipgloss.Green
	ColorBlue          = lipgloss.Blue
	ColorCyan          = lipgloss.Cyan
	ColorMagenta       = lipgloss.Magenta
	ColorWhite         = lipgloss.White
	ColorBlack         = lipgloss.Black
	ColorBrightRed     = lipgloss.BrightRed
	ColorBrightYellow  = lipgloss.BrightYellow
	ColorBrightGreen   = lipgloss.BrightGreen
	ColorBrightBlue    = lipgloss.BrightBlue
	ColorBrightCyan    = lipgloss.BrightCyan
	ColorBrightMagenta = lipgloss.BrightMagenta
	ColorBrightWhite   = lipgloss.BrightWhite
	ColorBrightBlack   = lipgloss.BrightBlack

	ColorPrimary   = ColorYellow
	ColorSecondary = ColorBrightBlack
)

type Opts struct {
	Type      OutputType
	NoNL      bool
	NL        string
	Typewrite time.Duration
}

type stream struct {
	writer     io.Writer
	isTerminal bool
	inColor    bool
}

func newStream(file *os.File, oc OutputColor) *stream {
	s := new(stream)

	s.writer = file
	s.isTerminal = isatty.IsTerminal(file.Fd()) == true ||
		isatty.IsCygwinTerminal(file.Fd()) == true

	switch oc {
	case ColorNever:
		s.inColor = false
	case ColorAuto:
		s.inColor = s.isTerminal
	case ColorAlways:
		s.inColor = true
	}

	return s
}

type Out struct {
	stdout *stream
	stderr *stream
}

func New(oc OutputColor) *Out {
	o := new(Out)

	o.stdout = newStream(os.Stdout, oc)
	o.stderr = newStream(os.Stderr, oc)

	return o
}

func (o *Out) streamFor(ot OutputType) *stream {
	if ot == Error {
		return o.stderr
	}
	return o.stdout
}

func (o *Out) InColor() bool {
	return o.stdout.inColor
}

func (o *Out) FG(c color.Color, format string, a ...any) string {
	return o.Stylize(Style{FG: c}, format, a...)
}

func (o *Out) BG(c color.Color, format string, a ...any) string {
	return o.Stylize(Style{BG: c}, format, a...)
}

func (o *Out) Stylize(
	st Style,
	format string, a ...any,
) string {
	text := fmt.Sprintf(format, a...)
	if o.InColor() {
		style := lipgloss.NewStyle()
		if st.FG != nil {
			style = style.Foreground(st.FG)
		}
		if st.BG != nil {
			style = style.Background(st.BG)
		}
		if st.PX > 0 {
			style = style.PaddingLeft(st.PX).PaddingRight(st.PX)
		}
		return style.Render(text)
	}
	return text
}

func (o *Out) Put(opts Opts, format string, a ...any) {
	var s *stream = o.streamFor(opts.Type)
	var formatted string = fmt.Sprintf(format, a...)
	var prefix string = OutputPrefixes[opts.Type].Char
	var nl string = "\n"

	if opts.NoNL == true {
		nl = ""
	} else {
		if opts.NL != "" {
			nl = opts.NL
		}
	}

	if s.inColor == true {
		style := lipgloss.NewStyle().Foreground(OutputPrefixes[opts.Type].Color)
		prefix = style.Render(prefix)
	}

	var output string = fmt.Sprintf("%s%s%s", prefix, formatted, nl)

	if s.isTerminal == true && opts.Typewrite > 0 {
		for _, char := range output {
			fmt.Fprintf(s.writer, "%c", char)
			time.Sleep(time.Millisecond * opts.Typewrite)
		}
		time.Sleep(time.Millisecond * opts.Typewrite * 2)
	} else {
		fmt.Fprint(s.writer, output)
	}
}
