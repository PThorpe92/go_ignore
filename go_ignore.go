package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("TEST!") //Test directory filenames
	dirlist := []string{"someifle.tmp", "home/etc", "main.go", "readme.md", "scene.nfo", "lotsoffiles.tmp", "build", "build/main.exe", "file.go", "anotherfile.go", "success.go"}
	gitig, err := Check("./")
	if err != nil {
		fmt.Printf("Error checking for .gitignore file %s\n", err)
		return
	}
	fmt.Printf("Rules to ignore: \n%v\n", gitig.Rules)
	for _, word := range dirlist {
		stat, err := os.Stat(word)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			fmt.Println("dir")
			if gitig.IsIgnored(word, true) {
				fmt.Printf("Ignored folders: \n%v\n", gitig.IgnoredFolders)
			}
		} else {
			fmt.Println("add")
			fmt.Println("not a dir")
			if gitig.IsIgnored(word, false) {
				fmt.Printf("Ignored files: \n%v\n", gitig.IgnoredFiles)
			}
		}
	}
}

type GitignoreNotFoundError struct {
	dir string
}

func (e GitignoreNotFoundError) Error() string {
	return fmt.Sprintf("gitignore file not found in directory %s", e.dir)
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

// This looks for a .gitignore file in the supplied path and returns
// an object with the rules for ignoring files/folders.
func Check(path string) (*GitIgnore, error) {
	var ignoreList = new(GitIgnore)
	var rules = new(Rules)
	file, err := os.ReadFile(fmt.Sprintf("%s/.testignore", path))
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
		} //TODO
		if strings.HasPrefix(strline, "!") {
			if strings.Contains(strline[1:], "*") {
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
			for i := 0; i < len(rules.Directory); i++ { // TODO
				if rules.Directory[i].String() == "^"+regexp.QuoteMeta(strline)+"$" {
					rules.Directory = append(rules.Directory[:i], rules.Directory[i+1:]...)
					i--
				}
			}
			for i := 0; i < len(rules.Character); i++ { //TODO
				if rules.Character[i].String() == "^"+regexp.QuoteMeta(strline)+"$" {
					rules.Character = append(rules.Character[:i], rules.Character[i+1:]...)
					i--
				}
			}
			for i := 0; i < len(rules.Subdir); i++ { //TODO
				if rules.Subdir[i].String() == strings.ReplaceAll(regexp.QuoteMeta(strline), "**", " ") {
					rules.Subdir = append(rules.Subdir[:i], rules.Subdir[i+1:]...)
					i--
				}
			}
			negation = false
			continue
		}
		if strings.HasPrefix(strline, "/") { //TODO
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline) + `(\\/\.*)\?$`)
			if err != nil {
				return nil, err
			}
			rules.Directory = append(rules.Directory, pattern)
		}
		if strings.Contains(strline, "**") { //TODO
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
