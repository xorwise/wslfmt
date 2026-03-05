package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xorwise/wslfmt/internal/fixer"
)

var (
	writeFlag = flag.Bool("w", false, "write result to (source) file instead of stdout")
	listFlag  = flag.Bool("l", false, "list files whose formatting differs from wslfmt's")
	diffFlag  = flag.Bool("d", false, "display diffs instead of rewriting files")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: wslfmt [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		// Read from stdin, write to stdout.
		if err := processReader(os.Stdin, os.Stdout, "<standard input>"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	exitCode := 0
	for _, path := range args {
		// Handle Go-style recursive patterns like ./... or pkg/...
		if strings.HasSuffix(path, "/...") || path == "..." {
			dir := strings.TrimSuffix(path, "/...")
			if dir == "" || dir == "..." {
				dir = "."
			}
			if err := walkDir(dir, &exitCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
				exitCode = 1
			}
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			exitCode = 1
			continue
		}
		if info.IsDir() {
			if err := walkDir(path, &exitCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
				exitCode = 1
			}
		} else {
			if err := processFile(path, &exitCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
				exitCode = 1
			}
		}
	}
	os.Exit(exitCode)
}

func walkDir(dir string, exitCode *int) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip vendor directory.
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			if err := processFile(path, exitCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		return nil
	})
}

func processFile(path string, exitCode *int) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fixed, err := fixer.Fix(src)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	if bytes.Equal(src, fixed) {
		return nil
	}

	if *listFlag {
		fmt.Println(path)
	}

	if *diffFlag {
		printDiff(path, src, fixed)
	}

	if *writeFlag {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, fixed, info.Mode()); err != nil {
			return err
		}
	} else if !*listFlag && !*diffFlag {
		os.Stdout.Write(fixed)
	}

	if *listFlag || *diffFlag {
		*exitCode = 1
	}

	return nil
}

func processReader(r io.Reader, w io.Writer, name string) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	fixed, err := fixer.Fix(src)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	_, err = w.Write(fixed)
	return err
}

// printDiff prints a simple unified diff between src and fixed.
func printDiff(path string, src, fixed []byte) {
	srcLines := strings.Split(string(src), "\n")
	fixedLines := strings.Split(string(fixed), "\n")

	fmt.Printf("--- %s\n+++ %s (fixed)\n", path, path)

	// Simple line-by-line diff output (not a full unified diff, but readable).
	maxLen := len(srcLines)
	if len(fixedLines) > maxLen {
		maxLen = len(fixedLines)
	}

	i, j := 0, 0
	for i < len(srcLines) || j < len(fixedLines) {
		srcLine := ""
		fixedLine := ""
		if i < len(srcLines) {
			srcLine = srcLines[i]
		}
		if j < len(fixedLines) {
			fixedLine = fixedLines[j]
		}
		if srcLine != fixedLine {
			if i < len(srcLines) {
				fmt.Printf("-%s\n", srcLine)
			}
			if j < len(fixedLines) {
				fmt.Printf("+%s\n", fixedLine)
			}
		}
		i++
		j++
		_ = maxLen
	}
}
