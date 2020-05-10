package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"io/ioutil"
)

// Type alias and struct types for recording entries.
type Topic string

type Note struct {
	name      string
	timestamp time.Time
}

type Notes []Note

func (notes Notes) Len() int {
	return len(notes)
}

func (notes Notes) Less(i int, j int) bool {
	// Sort by name, then by position.
	return notes[i].name < notes[j].name
}

func (notes Notes) Swap(i int, j int) {
	notes[i], notes[j] = notes[j], notes[i]
}

type Entry struct {
	notes     Notes
	subTopics map[Topic]*Entry
}

// Constructor for Entry
func blankEntry() Entry {
	return Entry{notes: []Note{}, subTopics: map[Topic]*Entry{}}
}

func (entry Entry) dump(path string, indent int, fileExt string) string {
	result := ""
	indentStr := strings.Repeat(" ", indent)

	notes := entry.notes
	sort.Stable(notes)

	for _, note := range notes {
		timestamp := note.timestamp.Format("02 Jan 2006")

		url := note.name
		if path != "" {
			url = path + "/" + url
		}

		name := note.name[:len(note.name)-len(fileExt)]

		dump := fmt.Sprintf("%s- [%s](%s) [%s]", indentStr, name, url, timestamp)
		result += dump + "\n"
	}

	// Make a list of all keys so that we can sort them, and thus iterate over
	// all keys in sorted order.
	keys := make([]string, 0, len(entry.subTopics))

	for key := range entry.subTopics {
		keys = append(keys, string(key))
	}

	sort.Strings(keys)

	for _, key := range keys {
		subPath := key
		if path != "" {
			subPath = path + "/" + subPath
		}

		name := key
		subTopic := Topic(key)

		headingMarker := strings.Repeat("#", indent)

		dump := fmt.Sprintf("\n%s%s %s", indentStr, headingMarker, name)
		result += dump + "\n"

		subEntry := entry.subTopics[subTopic]
		result += subEntry.dump(subPath, indent+1, fileExt)
	}

	return result
}

func (entry Entry) Dump(fileExt string) string {
	result := "# Notes\n"
	result += entry.dump("", 2, fileExt)

	return result
}

// Key traversal function.  Start with `basePath`, check for files with
// `fileExt` extension, add them (and subdirs) to `entry`, but make sure you
// don't add the same file as `outInfo`.
func __traverseDir(basePath string, fileExt string, entry *Entry, outInfo os.FileInfo) {
	files, err := ioutil.ReadDir(basePath)

	if err != nil {
		// TODO: Add better error handling.
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		subTopic := Topic(name)
		fullPath := basePath + string(os.PathSeparator) + name

		if file.IsDir() {
			// Ignore any hidden directories.
			if strings.HasPrefix(name, ".") == false {

				// Check whether the map entry exists.  If not, create one.
				_, test := entry.subTopics[subTopic]
				if test == false {
					subEntry := blankEntry()
					entry.subTopics[subTopic] = &subEntry
				}

				// Recurse down to the next level.
				subEntry := entry.subTopics[subTopic]
				__traverseDir(fullPath, fileExt, subEntry, outInfo)
			}
		} else if strings.HasSuffix(file.Name(), fileExt) {
			// Include this note only if it is not the output file.
			if os.SameFile(outInfo, file) == false {

				note := Note{name: name, timestamp: file.ModTime()}
				entry.notes = append(entry.notes, note)
			}
		}
	}
}

// Top-level traversal function.
func traverseDir(basePath string, fileExt string, outInfo os.FileInfo) Entry {
	rootEntry := blankEntry()
	__traverseDir(basePath, fileExt, &rootEntry, outInfo)

	return rootEntry
}

func main() {
	outputFile := flag.String("out", "README.md", "Path to output file.")
	fileExt := flag.String("ext", ".md", "Index files that have this extension.")

	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		panic("I need a path to parse, terminating.")
	}

	dirPath := args[0]

	outInfo, err := os.Stat(*outputFile)
	if err != nil {
		outInfo = nil
	}

	rootEntry := traverseDir(dirPath, *fileExt, outInfo)

	dumpText := rootEntry.Dump(*fileExt)
	ioutil.WriteFile(*outputFile, []byte(dumpText), 0644)
}
