package log

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func capture(t *testing.T, lvl slog.Level, fn func(logger *Logger)) (string, string) {
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

	fn(New(lvl))

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

func TestEverythingGoesToStderr(t *testing.T) {
	stdout, stderr := capture(t, slog.LevelDebug, func(logger *Logger) {
		logger.Error("an error")
		logger.Warning("a warning")
		logger.Info("some info")
		logger.Debug("a debug line")
		logger.Errorf("an %s error", "formatted")
		logger.Warningf("a %s warning", "formatted")
		logger.Infof("some %s info", "formatted")
		logger.Debugf("a %s debug line", "formatted")
	})

	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}

	for _, want := range []string{
		"an error", "a warning", "some info", "a debug line",
		"an formatted error", "a formatted warning",
		"some formatted info", "a formatted debug line",
	} {
		if strings.Contains(stderr, want) == false {
			t.Errorf("stderr is missing %q", want)
		}
	}
}

func TestLevelFiltering(t *testing.T) {
	_, stderr := capture(t, slog.LevelError, func(logger *Logger) {
		logger.Error("an error")
		logger.Warning("a warning")
		logger.Info("some info")
		logger.Debug("a debug line")
	})

	if strings.Contains(stderr, "an error") == false {
		t.Errorf("stderr is missing the error line")
	}
	for _, unwanted := range []string{"a warning", "some info", "a debug line"} {
		if strings.Contains(stderr, unwanted) == true {
			t.Errorf("stderr contains %q, which is below the level", unwanted)
		}
	}
}

func TestStructuredArguments(t *testing.T) {
	_, stderr := capture(t, slog.LevelDebug, func(logger *Logger) {
		logger.Debug("looking up", "key", "block:1", "count", 3)
	})

	for _, want := range []string{"looking up", "key=block:1", "count=3"} {
		if strings.Contains(stderr, want) == false {
			t.Errorf("stderr = %q, missing %q", stderr, want)
		}
	}
}
