// Copyright 2022 Marek Dalewski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var (
	Version string = "development"
	Commit  string = "?"
)

func reportError(err interface{}) bool {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true
	}
	return false
}

func reportWarning(err interface{}) bool {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		return true
	}
	return false
}

func fatal(err interface{}) {
	if reportError(err) {
		os.Exit(1)
	}
}

func defaultChromeConfigLocation() (string, error) {
	switch runtime.GOOS {
	case "linux":
		p, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(p, `.config/google-chrome/Default/Bookmarks`), nil
	case "windows":
		p, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(p, `AppData\Local\Google\Chrome\User Data`), nil
	case "darwin":
		p, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(p, `Library/Application Support/Google/Chrome`), nil
	default:
		return "", errors.New("unsupported OS " + runtime.GOOS)
	}
}

type WriteSyncCloser interface {
	io.Writer
	io.Closer
	Sync() error
}

type stdoutWrapper struct {
	w io.Writer
}

func (nc stdoutWrapper) Write(p []byte) (int, error) {
	return nc.w.Write(p)
}

func (nc stdoutWrapper) Sync() error {
	return nil
}

func (nc stdoutWrapper) Close() error {
	return nil
}

func makeOutput(path string) (WriteSyncCloser, error) {
	if path == "" {
		return stdoutWrapper{os.Stdout}, nil
	}
	return os.Create(path)
}

func findAllBookmarksFiles(path string, maxDepth int) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	if maxDepth <= 0 {
		for _, e := range entries {
			if name := e.Name(); name == "Bookmarks" && !e.IsDir() {
				return []string{filepath.Join(path, name)}, nil
			}
		}
		return nil, nil
	}

	res := []string(nil)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			li, err := findAllBookmarksFiles(filepath.Join(path, name), maxDepth-1)
			reportError(err)
			res = append(res, li...)
		} else if name == "Bookmarks" {
			res = append(res, filepath.Join(path, name))
		}
	}

	return res, nil
}

func writef(w io.Writer, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(w, format, args...)
	return err
}

type bookmarks struct {
	Version int                        `json:"version"`
	Roots   map[string]*bookmarksEntry `json:"roots"`
}

type bookmarksEntry struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Url      string `json:"url"`
	Children []*bookmarksEntry
}

func convertBookmarksFile(w io.Writer, bookmarksFile string, profileName string, indent string) error {
	bookmarksData, err := os.ReadFile(bookmarksFile)
	if err != nil {
		return err
	}

	b := &bookmarks{}
	if err := json.Unmarshal(bookmarksData, b); err != nil {
		return err
	}

	if b.Version != 1 {
		reportWarning(fmt.Sprintf("bookmarks file %s: unknown version %d, expected 1", bookmarksFile, b.Version))
	}

	if err := writef(w, "## Profile %s\n", profileName); err != nil {
		return err
	}
	for _, entry := range b.Roots {
		if err := convertBookmarksEntry(w, entry, "", indent); err != nil {
			return err
		}
	}
	return writef(w, "\n")
}

func convertBookmarksEntries(w io.Writer, entries []*bookmarksEntry, prefix, indent string) error {
	for _, e := range entries {
		if err := convertBookmarksEntry(w, e, prefix, indent); err != nil {
			return err
		}
	}
	return nil
}

func convertBookmarksEntry(w io.Writer, entry *bookmarksEntry, prefix, indent string) error {
	if entry.Type == "url" || entry.Url != "" {
		if err := writef(w, "%s- [%s](%s)\n", prefix, entry.Name, entry.Url); err != nil {
			return err
		}
	} else {
		if err := writef(w, "%s- %s\n", prefix, entry.Name); err != nil {
			return err
		}
	}
	return convertBookmarksEntries(w, entry.Children, prefix+indent, indent)
}

func showVersion() {
	fmt.Printf("Version of application: %s, commit: %s\n", Version, Commit)
	fmt.Printf("\n")
	fmt.Printf("Copyright 2022 Marek Dalewski. License: Apache License 2.0\n")
	fmt.Printf("\n")
	fmt.Printf("You should have received a copy of the Apache License 2.0 along with this program. If not, see <https://www.apache.org/licenses/LICENSE-2.0>.\n")
}

func main() {
	defaultInput, _ := defaultChromeConfigLocation() // on error user should provide path with flag

	input := flag.String("input", defaultInput, "path containing Chrome configuration")
	output := flag.String("output", "", "output path for storing generated document, leave empty for stdout")
	profiles := flag.String("profiles", "", "comma separated list of profile names that should be included in output, leave empty for all profiles")
	indent := flag.String("indent", "\\t", "string used for indentation")
	version := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *version {
		showVersion()
		os.Exit(0)
	}

	*input = filepath.Clean(*input)

	if *profiles != "" {
		*profiles = strings.ReplaceAll(*profiles, string(os.PathSeparator), "/")
		*profiles = strings.ReplaceAll(*profiles, "/,", ",")
		*profiles = strings.ReplaceAll(*profiles, ",/", ",")
		*profiles = "," + *profiles + ","
	}

	*indent = strings.ReplaceAll(*indent, "\\t", "\t")
	*indent = strings.ReplaceAll(*indent, "\\n", "\n")
	*indent = strings.ReplaceAll(*indent, "\\r", "\r")

	out, err := makeOutput(*output)
	fatal(err)
	defer out.Close()

	bookmarksFiles, err := findAllBookmarksFiles(*input, 25)
	fatal(err)
	sort.Strings(bookmarksFiles)

	if len(bookmarksFiles) == 0 {
		os.Exit(0)
	}

	fatal(writef(out, "# Chrome bookmarks\n"))
	fatal(writef(out, "\n"))
	fatal(writef(out, "> This document was automatically generated by [chrome-bookmarks-to-markdown](https://github.com/daishe/chrome-bookmarks-to-markdown).\n"))
	fatal(writef(out, "\n"))
	for _, b := range bookmarksFiles {
		p := strings.TrimSuffix(strings.TrimPrefix(b, *input+string(os.PathSeparator)), string(os.PathSeparator)+"Bookmarks")
		if *profiles != "" && !strings.Contains(*profiles, ","+p+",") {
			continue
		}
		fatal(convertBookmarksFile(out, b, p, *indent))
	}
	fatal(out.Sync())
}
