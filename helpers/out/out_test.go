package out

import (
	"os"
	"strings"
	"testing"
	"time"
)

type capture struct {
	stdout string
	stderr string
}

func run(t *testing.T, oc OutputColor, fn func(o *Out)) capture {
	t.Helper()

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %s", err)
	}

	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %s", err)
	}

	previousOut, previousErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outWriter, errWriter

	o := New(oc)
	fn(o)

	os.Stdout, os.Stderr = previousOut, previousErr
	outWriter.Close()
	errWriter.Close()

	var result capture

	outBytes := make([]byte, 64*1024)
	n, _ := outReader.Read(outBytes)
	result.stdout = string(outBytes[:max(n, 0)])
	outReader.Close()

	errBytes := make([]byte, 64*1024)
	n, _ = errReader.Read(errBytes)
	result.stderr = string(errBytes[:max(n, 0)])
	errReader.Close()

	return result
}

func TestPutRoutesErrorsToStderr(t *testing.T) {
	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Error}, "something broke")
	})

	if got.stdout != "" {
		t.Errorf("stdout = %q, want empty", got.stdout)
	}
	if strings.Contains(got.stderr, "something broke") == false {
		t.Errorf("stderr = %q, want it to contain the message", got.stderr)
	}
}

func TestPutRoutesEverythingElseToStdout(t *testing.T) {
	types := []OutputType{Plain, Ok, Info, Start, Switch, Resume, End, Pause, Erase}

	for _, ot := range types {
		got := run(t, ColorNever, func(o *Out) {
			o.Put(Opts{Type: ot}, "regular output")
		})

		if strings.Contains(got.stdout, "regular output") == false {
			t.Errorf("type %d: stdout = %q, want it to contain the message",
				ot, got.stdout)
		}
		if got.stderr != "" {
			t.Errorf("type %d: stderr = %q, want empty", ot, got.stderr)
		}
	}
}

func TestPutKeepsStreamsSeparate(t *testing.T) {
	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Plain}, `{"status":"tracking"}`)
		o.Put(Opts{Type: Error}, "a warning about something")
		o.Put(Opts{Type: Plain}, `{"status":"ended"}`)
	})

	if strings.Contains(got.stdout, "a warning") == true {
		t.Errorf("stdout = %q, want it free of error output", got.stdout)
	}
	if strings.Count(got.stdout, "status") != 2 {
		t.Errorf("stdout = %q, want both payload lines", got.stdout)
	}
}

func TestPutAppliesPrefixes(t *testing.T) {
	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Ok}, "done")
	})

	want := OutputPrefixes[Ok].Char + "done\n"
	if got.stdout != want {
		t.Errorf("stdout = %q, want %q", got.stdout, want)
	}
}

func TestPutNewlineOptions(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
		want string
	}{
		{name: "default", opts: Opts{Type: Plain}, want: "line\n"},
		{name: "suppressed", opts: Opts{Type: Plain, NoNL: true}, want: "line"},
		{name: "custom", opts: Opts{Type: Plain, NL: "!"}, want: "line!"},
		{
			name: "suppressed wins over custom",
			opts: Opts{Type: Plain, NoNL: true, NL: "!"},
			want: "line",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := run(t, ColorNever, func(o *Out) {
				o.Put(test.opts, "line")
			})

			if got.stdout != test.want {
				t.Errorf("stdout = %q, want %q", got.stdout, test.want)
			}
		})
	}
}

func TestPutFormatsArguments(t *testing.T) {
	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Plain}, "%s/%s took %d seconds", "alpha", "one", 42)
	})

	if got.stdout != "alpha/one took 42 seconds\n" {
		t.Errorf("stdout = %q", got.stdout)
	}
}

func TestColorNeverProducesNoEscapes(t *testing.T) {
	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Ok}, "%s", o.FG(ColorRed, "red text"))
		o.Put(Opts{Type: Error}, "an error")
	})

	if strings.Contains(got.stdout, "\x1b[") == true {
		t.Errorf("stdout = %q, want no escape sequences", got.stdout)
	}
	if strings.Contains(got.stderr, "\x1b[") == true {
		t.Errorf("stderr = %q, want no escape sequences", got.stderr)
	}
}

func TestColorAlwaysProducesEscapesOnBothStreams(t *testing.T) {
	got := run(t, ColorAlways, func(o *Out) {
		o.Put(Opts{Type: Ok}, "fine")
		o.Put(Opts{Type: Error}, "broken")
	})

	if strings.Contains(got.stdout, "\x1b[") == false {
		t.Errorf("stdout = %q, want escape sequences", got.stdout)
	}
	if strings.Contains(got.stderr, "\x1b[") == false {
		t.Errorf("stderr = %q, want escape sequences", got.stderr)
	}
}

func TestColorAutoDisablesColorOnPipes(t *testing.T) {
	got := run(t, ColorAuto, func(o *Out) {
		o.Put(Opts{Type: Ok}, "fine")
		o.Put(Opts{Type: Error}, "broken")

		if o.InColor() == true {
			t.Errorf("InColor() = true for a piped stdout")
		}
	})

	if strings.Contains(got.stdout, "\x1b[") == true {
		t.Errorf("stdout = %q, want no escape sequences", got.stdout)
	}
	if strings.Contains(got.stderr, "\x1b[") == true {
		t.Errorf("stderr = %q, want no escape sequences", got.stderr)
	}
}

func TestInColorFollowsTheColorSetting(t *testing.T) {
	tests := []struct {
		oc   OutputColor
		want bool
	}{
		{oc: ColorNever, want: false},
		{oc: ColorAlways, want: true},
		{oc: ColorAuto, want: false},
	}

	for _, test := range tests {
		run(t, test.oc, func(o *Out) {
			if o.InColor() != test.want {
				t.Errorf("InColor() = %t for %d, want %t", o.InColor(), test.oc, test.want)
			}
		})
	}
}

func TestStylize(t *testing.T) {
	run(t, ColorNever, func(o *Out) {
		got := o.Stylize(Style{FG: ColorRed, BG: ColorBlue, PX: 2}, "%s!", "text")
		if got != "text!" {
			t.Errorf("Stylize() = %q, want the plain text", got)
		}
	})

	run(t, ColorAlways, func(o *Out) {
		got := o.Stylize(Style{FG: ColorRed}, "%s!", "text")
		if strings.Contains(got, "text!") == false {
			t.Errorf("Stylize() = %q, want it to contain the text", got)
		}
		if strings.Contains(got, "\x1b[") == false {
			t.Errorf("Stylize() = %q, want escape sequences", got)
		}
	})
}

func TestForegroundAndBackgroundHelpers(t *testing.T) {
	run(t, ColorAlways, func(o *Out) {
		fg := o.FG(ColorRed, "%s", "text")
		bg := o.BG(ColorBlue, "%s", "text")

		if strings.Contains(fg, "text") == false || strings.Contains(fg, "\x1b[") == false {
			t.Errorf("FG() = %q, want styled text", fg)
		}
		if strings.Contains(bg, "text") == false || strings.Contains(bg, "\x1b[") == false {
			t.Errorf("BG() = %q, want styled text", bg)
		}
		if fg == bg {
			t.Errorf("FG() and BG() produced the same output")
		}
	})
}

func TestTypewriteIsSkippedOnPipes(t *testing.T) {
	start := time.Now()

	got := run(t, ColorNever, func(o *Out) {
		o.Put(Opts{Type: Plain, Typewrite: 50 * time.Millisecond}, "abcdefghij")
	})

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("writing to a pipe took %s, want the typewriter to be skipped", elapsed)
	}
	if got.stdout != "abcdefghij\n" {
		t.Errorf("stdout = %q", got.stdout)
	}
}

func TestOutputPrefixesCoverEveryType(t *testing.T) {
	if len(OutputPrefixes) != int(Erase)+1 {
		t.Errorf("OutputPrefixes has %d entries, want %d",
			len(OutputPrefixes), int(Erase)+1)
	}
}
