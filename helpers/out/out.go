package out

import (
	"fmt"
	"image/color"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss/v2"
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

type Out struct {
	oc OutputColor
}

func New(oc OutputColor) *Out {
	o := new(Out)

	if oc == ColorAuto {
		if o.isPiped() == true {
			o.oc = ColorNever
		} else {
			o.oc = ColorAlways
		}
	} else {
		o.oc = oc
	}
	return o
}

func (o *Out) isPiped() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) == false &&
		isatty.IsCygwinTerminal(os.Stdout.Fd()) == false
}

func (o *Out) InColor() bool {
	return o.oc == ColorAlways
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
	var formatted string = fmt.Sprintf(format, a...)
	var nl string = "\n"
	var output string = ""

	if opts.NoNL == true {
		nl = ""
	} else {
		if opts.NL != "" {
			nl = opts.NL
		}
	}

	if o.oc == ColorAlways {
		style := lipgloss.NewStyle().Foreground(OutputPrefixes[opts.Type].Color)
		output = fmt.Sprintf("%s%s%s",
			style.Render(OutputPrefixes[opts.Type].Char),
			formatted,
			nl,
		)
	} else {
		output = fmt.Sprintf("%s%s%s",
			OutputPrefixes[opts.Type].Char,
			formatted,
			nl,
		)
	}

	if o.isPiped() == false && opts.Typewrite > 0 {
		for _, char := range output {
			fmt.Printf("%c", char)
			time.Sleep(time.Millisecond * opts.Typewrite)
		}
		time.Sleep(time.Millisecond * opts.Typewrite * 2)
	} else {
		fmt.Printf("%s", output)
	}
}
