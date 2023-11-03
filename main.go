package main

import (
	"bufio"
	"flag"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var path, lang, mergefile string
var ignore []string
var patterns = map[string]string{
	"go":  `os\.Getenv\("([^"]*)\"\)`,
	"js":  `process.env.([^.]+)`,
	"jsx": `process.env.([^.]+)`,
	"ts":  `process.env.([^.]+)`,
	"tsx": `process.env.([^.]+)`,
}

func option() {
	ppath := flag.String("p", "", "Path")
	plang := flag.String("l", "", "Language")
	pmerge := flag.String("m", ".env", "Merge .env file target")
	pignore := flag.String("i", "", "Ignore names")
	flag.Parse()

	path = *ppath
	lang = *plang
	mergefile = *pmerge
	ignore = strings.Split(*pignore, ",")

	var supported bool
	for support := range patterns {
		if support == lang {
			supported = true
		}
	}
	if !supported {
		log.Panicln("not support language")
	}

	if path == "" {
		log.Panicln("path required")
	}

	return
}

func read(path string, file fs.FileInfo, keys *[]string) {
	ext := filepath.Ext(file.Name())
	if ext != "."+lang {
		return
	}

	f, e := os.Open(path + "/" + file.Name())
	if e != nil {
		log.Println("file ", path+"/"+file.Name(), e.Error())
	}
	defer f.Close()

	b, e := io.ReadAll(f)
	if e != nil {
		return
	}

	pattern := patterns[lang]
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(string(b), -1)

	for _, m := range matches {
		if len(m) >= 2 {
			g := strings.Split(m[1], "\n")
			(*keys) = append((*keys), g[0])
		}
	}
}

func scan(path string, keys *[]string) {
	f, e := os.Open(path)
	if e != nil {
		log.Panicln(e)
	}
	defer f.Close()

	files, e := f.Readdir(-1)
	if e != nil {
		return
	}

	for _, file := range files {
		if slices.Contains[[]string, string](ignore, file.Name()) {
			continue
		}
		if file.IsDir() {
			scan(path+"/"+file.Name(), keys)
			continue
		}
		read(path, file, keys)
	}
}

func merge(keys *[]string) {
	f, _ := os.Open(mergefile)
	newlines := []string{}
	dupkey := map[string]bool{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		ws := strings.Split(line, "=")
		for i := range ws {
			ws[i] = strings.Trim(ws[i], "")
		}

		if len(ws) < 1 {
			continue
		}

		if slices.Contains[[]string, string](*keys, ws[0]) {
			dupkey[ws[0]] = true
			newlines = append(newlines, line+"\n")
			continue
		}
	}
	f.Close()

	for _, k := range *keys {
		if dupkey[k] {
			continue
		}
		newlines = append(newlines, k+"=\n")
	}

	raw := strings.Join(newlines, "")
	os.WriteFile(path+"/.genv", []byte(raw), fs.ModePerm)
}

func main() {
	option()
	keys := []string{}
	scan(path, &keys)
	merge(&keys)
}
