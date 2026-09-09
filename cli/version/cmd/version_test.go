package versionCmd

import (
	"os"
	"strings"
	"testing"

	"xn--gckvb8fzb.com/zeit/helpers/out"
	"xn--gckvb8fzb.com/zeit/runtime"
)

func render(t *testing.T, oc out.OutputColor, build runtime.Build) (string, string) {
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

	outputVersion(out.New(oc), build)

	os.Stdout, os.Stderr = previousOut, previousErr
	outWriter.Close()
	errWriter.Close()

	read := func(f *os.File) string {
		buffer := make([]byte, 64*1024)
		n, _ := f.Read(buffer)
		f.Close()

		return string(buffer[:max(n, 0)])
	}

	return read(outReader), read(errReader)
}

func TestOutputVersion(t *testing.T) {
	build := runtime.Build{
		Version: "v1.2.3",
		Commit:  "0123456789abcdef",
		Date:    "2026-08-26T13:10:14Z",
	}

	stdout, stderr := render(t, out.ColorNever, build)

	for _, want := range []string{
		"zeit v1.2.3",
		"Commit: 0123456789abcdef",
		"Build date: 2026-08-26T13:10:14Z",
	} {
		if strings.Contains(stdout, want) == false {
			t.Errorf("stdout = %q, missing %q", stdout, want)
		}
	}

	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}

func TestOutputVersionWithoutValues(t *testing.T) {
	stdout, _ := render(t, out.ColorNever, runtime.Build{})

	if strings.Contains(stdout, "zeit") == false {
		t.Errorf("stdout = %q, want the program name even without a build", stdout)
	}
}

func TestOutputVersionShowsTheBannerOnlyInColor(t *testing.T) {
	build := runtime.Build{Version: "v1.2.3"}

	plain, _ := render(t, out.ColorNever, build)
	if strings.Contains(plain, "█") == true {
		t.Errorf("the banner was drawn without color")
	}

	colored, _ := render(t, out.ColorAlways, build)
	if strings.Contains(colored, "█") == false {
		t.Errorf("the banner was not drawn in color")
	}
	if strings.Contains(colored, "v1.2.3") == false {
		t.Errorf("colored output = %q, missing the version", colored)
	}
}

func TestOutputVersionNeedsNoDatabase(t *testing.T) {
	t.Setenv(runtime.DATABASE_ENV_VAR, "/nonexistent/zeit/database")

	stdout, stderr := render(t, out.ColorNever, runtime.NewBuild())

	if strings.Contains(stdout, "zeit") == false {
		t.Errorf("stdout = %q, want version output", stdout)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want empty", stderr)
	}
}
