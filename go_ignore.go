package go_ignore

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A blank line matches no files, so it can serve as a separator for readability.
// A line starting with # serves as a comment. Put a backslash ("\") in front of the first hash for patterns that begin with a hash.
// Trailing spaces are ignored unless they are quoted with backslash ("\").
// An optional prefix "!" which negates the pattern; any matching file excluded by a previous pattern will become included again.
// It is not possible to re-include a file if a parent directory of that file is excluded. Git doesn’t list excluded directories for performance reasons,
//
//	so any patterns on contained files have no effect, no matter where they are defined. Put a backslash ("\") in front of the first "!" for patterns that
//	begin with a literal "!", for example, "\!important!.txt".
//
// // If the pattern ends with a slash, it is removed for the purpose of the following description, but it would only find a match with a directory. In other
// words, foo/ will match a directory foo and paths underneath it, but will not match a regular file or a symbolic link foo (this is consistent with the way how
//
//	pathspec works in general in Git).
//
// // If the pattern does not contain a slash /, Git treats it as a shell glob pattern and checks for a match against the pathname relative to the location of the
//
//	.gitignore file (relative to the toplevel of the work tree if not from a .gitignore file).
//
// // Otherwise, Git treats the pattern as a shell glob suitable for consumption by fnmatch(3) with the FNM_PATHNAME flag: wildcards in the pattern will not match
//
//	a / in the pathname. For example, "Documentation/*.html" matches "Documentation/git.html" but not "Documentation/ppc/ppc.html" or "tools/perf/Documentation/perf.html".
//
// A leading slash matches the beginning of the pathname. For example, "/*.c" matches "cat-file.c" but not "mozilla-sha1/sha1.c".
// Two consecutive asterisks ("**") in patterns matched against full pathname may have special meaning: i. A leading "**" followed by a slash means match in all directories. For example, "** /foo" matches file or directory "foo" anywhere, the same as pattern "foo". "** /foo/bar" matches file or directory "bar" anywhere that is directly under directory "foo". ii. A trailing "/**" matches everything inside. For example, "abc/**" matches all files inside directory "abc", relative to the location of the .gitignore file, with infinite depth. iii. A slash followed by two consecutive asterisks then a slash matches zero or more directories. For example, "a/** /b" matches "a/b", "a/x/b", "a/x/y/b" and so on. iv. Other consecutive asterisks are considered invalid.

type GitignoreNotFoundError struct {
	Dir string
}

func (e GitignoreNotFoundError) Error() string {
	return fmt.Sprintf("gitignore file not found in directory %s", e.Dir)
}

type ParsingGitignoreError struct {
	Dir string
}

func (e ParsingGitignoreError) Error() string {
	return fmt.Sprintf("Error parsing the .gitignore file found in directory %s", e.Dir)
}

type GitIgnore struct {
	Path         string
	Rules        *Rules
	IgnoredFiles []string
	TotalFiles   int
}

type Rules struct {
	Directory []*regexp.Regexp
	Character []*regexp.Regexp
	Subdir    []*regexp.Regexp
	Negate    []*regexp.Regexp
}
type Pattern struct {
	Line        string
	Line_number string
	Negate      bool
	Match       *regexp.Regexp
}

/*
If the absolute path to the .gitignore file isn't provided, the function
checks the supplied path for one and returns the compiled object which paths
can be checked against with
the IsIgnored method.
*/

func CheckPath(path string) (*GitIgnore, error) {
	if path == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("Error with supplied path (symbolic link)")
		}
		path = cwd
	}

	if path[len(path)-1] != filepath.Separator {
		path += string(filepath.Separator)
	}
	path = filepath.Join(path, ".gitignore")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Gitignore file not found in directory %s", filepath.Dir(path))
	} else if err != nil {
		return nil, fmt.Errorf("Error checking .gitignore file: %v", err)
	}

	ignoreFile, err := parseGitignore(path)
	if err != nil {
		return nil, fmt.Errorf("Error parsing .gitignore file: %v", err)
	}

	return ignoreFile, nil
}

func parseGitignore(path string) (*GitIgnore, error) {
	if !strings.Contains(path, ".gitignore") {
		_, err := os.Stat(filepath.Join(path + ".gitignore"))
		if err == nil {
			path = filepath.Join(path + ".gitignore")
		}
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rules := &Rules{
		Directory: make([]*regexp.Regexp, 0),
		Subdir:    make([]*regexp.Regexp, 0),
		Character: make([]*regexp.Regexp, 0),
		Negate:    make([]*regexp.Regexp, 0),
	}

	scanner := bufio.NewScanner(bytes.NewReader(file))
	for scanner.Scan() {
		line := scanner.Text()
		// if the line is empty, or starts with a '#', ignore as a comment
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.Trim(line, " ")
		// if the line starts with a '!' we treat it as a negated pattern
		if strings.HasPrefix(line, "!") {
			negatePattern := line[1:]
			pattern := regexp.MustCompile(negatePattern)
			if err != nil {
				return nil, err
			}
			rules.Negate = append(rules.Negate, pattern)
		} else {
			pattern, err := convertToRegexp(line)
			switch {
			case strings.HasPrefix(line, "/"):
				rules.Directory = append(rules.Directory, pattern)
			case strings.Contains(line, "**"):
				rules.Subdir = append(rules.Subdir, pattern)
			default:
				rules.Character = append(rules.Character, pattern)
			}
			if err != nil {
				return nil, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	result := &GitIgnore{
		Path:         path,
		Rules:        rules,
		IgnoredFiles: make([]string, 0),
	}
	return result, nil
}

// This function needs to convert the found patterns in the .gitignore file
// into a regular expression object
func convertToRegexp(pattern string) (*regexp.Regexp, error) {
	// Escape special characters in the pattern
	pattern = regexp.QuoteMeta(pattern)
	// Replace the escaped wildcard pattern "\*" with the unescaped version ".*"
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	// Replace the escaped wildcard character "?" with the regex pattern "."
	pattern = strings.ReplaceAll(pattern, "\\?", ".")

	// Handle negation patterns that start with "!"
	if strings.HasPrefix(pattern, "\\!") {
		// Negation patterns should match the entire line
		pattern = "^(?!" + pattern[2:] + ").*$"
		fmt.Println(pattern)
	} else {
		// Non-negation patterns should match the pattern within the line
		pattern = ".*" + pattern + ".*"
		fmt.Println(pattern)
	}

	// Compile the regular expression
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}

func removeRule(rules *[]*regexp.Regexp, strline string) {
	for i := 0; i < len(*rules); i++ {
		if (*rules)[i].String() == "^"+regexp.QuoteMeta(strline)+"$" {
			*rules = append((*rules)[:i], (*rules)[i+1:]...)
			i--
		}
	}
}

func (g *GitIgnore) IsIgnored(file string) bool {
	// Check for negation rules first
	for _, rule := range g.Rules.Negate {
		if rule.MatchString(file) {
			return false
		}
	}

	// Check character rules
	for _, rule := range g.Rules.Character {
		if rule.MatchString(file) {
			g.IgnoredFiles = append(g.IgnoredFiles, file)
			return true
		}
	}

	// Check directory rules
	for _, rule := range g.Rules.Directory {
		if rule.MatchString(file) {
			g.IgnoredFiles = append(g.IgnoredFiles, file)
			return true
		}
	}

	// Check subdirectory rules
	for _, rule := range g.Rules.Subdir {
		if rule.MatchString(file) {
			g.IgnoredFiles = append(g.IgnoredFiles, file)
			return true
		}
	}

	return false
}
