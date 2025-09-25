package out

import (
	"fmt"
	"image/color"
	"os"

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
	Track
	Stop
	Erase
)

type OutputProp struct {
	Char  string
	Color ansi.BasicColor
}

var OutputProps = []OutputProp{
	{
		Char:  "",
		Color: lipgloss.White,
	},
	{
		Char:  "●",
		Color: lipgloss.Green,
	},
	{
		Char:  "▲",
		Color: lipgloss.Red,
	},
	{
		Char:  "◆",
		Color: lipgloss.BrightBlack,
	},
	{
		Char:  "▶",
		Color: lipgloss.Cyan,
	},
	{
		Char:  "■",
		Color: lipgloss.Magenta,
	},
	{
		Char:  "◀",
		Color: lipgloss.BrightRed,
	},
}

var OutputChars = []string{
	"",
	"●",
	"▲",
	"◆",
	"▶",
	"■",
	"◀",
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
)

const (
	CharOk    = "●"
	CharError = "▲"
	CharTrack = "▶"
	CharStop  = "■"
	CharErase = "◀"
	CharInfo  = "◆"
)

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
	text := fmt.Sprintf(format, a...)
	if o.InColor() {
		style := lipgloss.NewStyle().Foreground(c)
		return style.Render(text)
	}
	return text
}

func (o *Out) BG(c color.Color, format string, a ...any) string {
	text := fmt.Sprintf(format, a...)
	if o.InColor() {
		style := lipgloss.NewStyle().Background(c)
		return style.Render(text)
	}
	return text
}

func (o *Out) FGBG(
	fgc color.Color,
	bgc color.Color,
	format string, a ...any,
) string {
	text := fmt.Sprintf(format, a...)
	if o.InColor() {
		style := lipgloss.NewStyle().Foreground(fgc).Background(bgc)
		return style.Render(text)
	}
	return text
}

func (o *Out) Put(ot OutputType, format string, a ...any) {
	var formatted string = fmt.Sprintf(format, a...)

	if o.oc == ColorAlways {
		style := lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)
		fmt.Printf("%s %s\n",
			style.Render(OutputProps[ot].Char),
			formatted,
		)
	} else {
		fmt.Printf("%s %s\n",
			OutputProps[ot].Char,
			formatted)
	}
}

func (o *Out) NilOrDie(err error, format string, a ...any) {
	if err != nil {
		o.Put(Error, format, a...)
		os.Exit(1)
	}
}

func (o *Out) Die(format string, a ...any) {
	o.Put(Error, format, a...)
	os.Exit(1)
}
