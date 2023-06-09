package go_ignore

import (
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

type GitIgnore struct {
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

// This function checks the supplied path for a .gitignorefile
// and returns the compiled object which paths can be checked against with
// the IsIgnored method.
func CheckPath(path string) (*GitIgnore, error) {

	if path == "." {
		path, err := os.Getwd()
		if err != nil {
			err = fmt.Errorf("Error with supplied path (symbolic link)")
			return nil, err
		}
		if !strings.Contains(path, ".gitignore") {
			path = filepath.Join(path, ".gitignore")
			_, err = os.Stat(path)
			if err != nil {
				err = fmt.Errorf("No .gitignore file found in the specified filepath \n")
				return nil, err
			}
		}
	}
	ignoreFile, err := parseGitignore(path)
	if err != nil {
		err = fmt.Errorf("Error parsing .gitignore file, improper format?")
		return nil, err
	}
	return ignoreFile, nil
}

// This looks for a .gitignore file in the supplied path and returns
// an object with the rules for ignoring files/folders
func parseGitignore(path string) (*GitIgnore, error) {
	var rules = new(Rules)
	file, err := os.ReadFile(path)
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
			// Handle negation rules
			// If the pattern contains a '*', convert it to '.*' for regex
			if strings.Contains(strline[1:], "*") {
				strline = strings.ReplaceAll(strline, "*", ".*")
			}
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline[1:]) + `$`)
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
				if strings.Contains(rules.Subdir[i].String(), "**") {
					// If the rule contains '**', replace it with '.*' before comparing with strline
					str := strings.ReplaceAll(regexp.QuoteMeta(strline), "**", ".*")
					if rules.Subdir[i].String() == "^"+str+"$" {
						rules.Subdir = append(rules.Subdir[:i], rules.Subdir[i+1:]...)
						i--
					}
				} else {
					if rules.Subdir[i].String() == "^"+regexp.QuoteMeta(strline)+"(/.*)?$" {
						rules.Subdir = append(rules.Subdir[:i], rules.Subdir[i+1:]...)
						i--
					}
				}
			}
			negation = false
			continue
		}
		if strings.HasPrefix(strline, "/") {
			// Handle directory rules
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline) + `(/\.*)?\?$`)
			if err != nil {
				return nil, err
			}
			rules.Directory = append(rules.Directory, pattern)
		}
		if strings.Contains(strline, "**") {
			// Handle subdirectory rules
			pattern, err := regexp.Compile("^" + strings.ReplaceAll(regexp.QuoteMeta(strline), "**", `.*`) + `$`)
			if err != nil {
				return nil, err
			}
			rules.Subdir = append(rules.Subdir, pattern)
		}
		if strings.ContainsAny(strline, "[]") {
			// Handle character rules
			pattern, err := regexp.Compile("^" + regexp.QuoteMeta(strline) + "$")
			if err != nil {
				return nil, err
			}
			rules.Character = append(rules.Character, pattern)
		}
	}
	result := new(GitIgnore)
	result.Rules = rules
	return result, nil
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
