package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("TEST!")
	dirlist := []string{"someifle.tmp", "home/etc", "main.go", "readme.md", "scene.nfo", "lotsoffiles.tmp", "build", "build/main.exe", "file.go", "anotherfile.go", "success.go"}
	gitig, err := ParseGitIgnore(".gitignore")
	if err != nil {
		panic(err)
	}
	fmt.Println("%s", gitig.Rules)
	for _, word := range dirlist {
		stat, err := os.Stat(word)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			fmt.Println("dir")
			if gitig.IsIgnored(word, true) {
				fmt.Println("%s", gitig.IgnoredFolders)
			}
		} else {
			fmt.Println("add")
			fmt.Println("not a dir")
			if gitig.IsIgnored(word, false) {
				fmt.Println("%s", gitig.IgnoredFiles)
			}
		}
	}
}

type GitIgnore struct {
	Rules          *Rules
	IgnoredFiles   []string
	IgnoredFolders []string
	TotalFiles     int
	TotalFolders   int
}
type Rules struct {
	Directory []*regexp.Regexp
	Character []*regexp.Regexp
	Subdir    []*regexp.Regexp
	Negate    []*regexp.Regexp
}

// This function parses .gitignore files and returns our objects
// updated with list of files we need to ignore
func ParseGitIgnore(filename string) (*GitIgnore, error) {
	var ignoreList = new(GitIgnore)
	var rules = new(Rules)
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	rules.Directory = make([]*regexp.Regexp, 0)
	rules.Subdir = make([]*regexp.Regexp, 0)
	rules.Character = make([]*regexp.Regexp, 0)
	str := string(file)
	lines := strings.Split(str, "\n")
	var negation bool
	for _, line := range lines {
		strline := strings.TrimSpace(line)
		if strline == "" || strings.HasPrefix(strline, "#") {
			continue
		}

		if strings.HasPrefix(strline, "!") {
			if strings.Contains(strline[1:], "*") {
				strings.Replace(strline, "*", "./", 2)
			}
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline[1:]) + `()$`)
			if err != nil {
				return nil, err
			}
			rules.Negate = append(rules.Negate, pattern)
			negation = true
			continue
		}

		// Check for negation rules that apply to the current line
		if negation {
			for i := 0; i < len(rules.Directory); i++ {
				if rules.Directory[i].String() == "^"+regexp.QuoteMeta(strline)+"$" {
					rules.Directory = append(rules.Directory[:i], rules.Directory[i+1:]...)
					i--
				}
			}
			for i := 0; i < len(rules.Character); i++ {
				if rules.Character[i].String() == "^"+regexp.QuoteMeta(strline)+"$" {
					rules.Character = append(rules.Character[:i], rules.Character[i+1:]...)
					i--
				}
			}
			for i := 0; i < len(rules.Subdir); i++ {
				if rules.Subdir[i].String() == strings.ReplaceAll(regexp.QuoteMeta(strline), "**", `\.*/`) {
					rules.Subdir = append(rules.Subdir[:i], rules.Subdir[i+1:]...)
					i--
				}
			}
			negation = false
			continue
		}
		if strings.HasPrefix(strline, "/") {
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline) + `(\\/\.*)\?$`)
			if err != nil {
				return nil, err
			}
			rules.Directory = append(rules.Directory, pattern)
		}
		if strings.Contains(strline, "**") {
			pattern, err := regexp.Compile(strings.ReplaceAll(regexp.QuoteMeta(strline), "**", `\./*`))
			if err != nil {
				return nil, err
			}
			rules.Subdir = append(rules.Subdir, pattern)
		}
		if strings.ContainsAny(strline, "[]") {
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline) + "$")
			if err != nil {
				return nil, err
			}
			rules.Character = append(rules.Character, pattern)
		}
	}
	ignoreList.Rules = rules
	return ignoreList, nil
}

func (g *GitIgnore) IsIgnored(file string, isDir bool) bool {
	for _, rule := range g.Rules.Negate {
		if rule.MatchString(file) {
			fmt.Println("success")
			return false
		}
	}
	for _, rule := range g.Rules.Character {
		if rule.MatchString(file) {
			g.IgnoredFiles = append(g.IgnoredFiles, file)
			fmt.Println("success")
			return true
		}
	}
	if isDir {
		for _, rule := range g.Rules.Directory {
			if rule.MatchString(file) {
				g.IgnoredFolders = append(g.IgnoredFolders, file)
				fmt.Println("success")
				return true
			}
		}
		for _, rule := range g.Rules.Subdir {
			if rule.MatchString(file) {
				g.IgnoredFolders = append(g.IgnoredFolders, file)
				fmt.Println("success")
				return true
			}
		}
	}
	return false
}
