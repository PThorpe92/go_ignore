package go_ignore

import (
	"regexp"
	"testing"
)

func TestIsIgnored(t *testing.T) {
	ignore := &GitIgnore{
		Rules: &Rules{
			Character: []*regexp.Regexp{
				regexp.MustCompile(`\.txt$`),          // Match .txt files
				regexp.MustCompile(`^test_file\.md$`), // Match exact file "test_file.md"
			},
			Directory: []*regexp.Regexp{
				regexp.MustCompile(`^ignored_dir/`),       // Match paths starting with "ignored_dir/"
				regexp.MustCompile(`^docs/`),              // Match paths starting with "docs/"
				regexp.MustCompile(`^ignored_file\.txt$`), // Match exact file "ignored_file.txt"
			},
			Subdir: []*regexp.Regexp{
				regexp.MustCompile(`^src/.*\.js$`), // Match .js files in "src/" subdirectories
			},
			Negate: []*regexp.Regexp{
				regexp.MustCompile(`^!important\.txt$`), // Exclude file "important.txt"
			},
		},
	}

	tests := []struct {
		file      string
		isIgnored bool
	}{
		{"test.txt", true},                // Matched by a character rule (.txt file)
		{"not_ignored.txt", true},         // Not matched by any rule
		{"test_file.md", true},            // Matched by a character rule (exact file name match)
		{"ignored_dir/file.txt", true},    // Matched by a directory rule
		{"docs/document.txt", true},       // Matched by a directory rule
		{"ignored_file.txt", true},        // Matched by a directory rule (exact file name match)
		{"src/main.js", true},             // Matched by a subdirectory rule
		{"src/subfolder/script.js", true}, // Matched by a subdirectory rule
		{"important.txt", true},           // Not ignored due to a negation rule
		{"ignored_dir", false},            // Not matched by any rule
		{"docs", false},                   // Not matched by any rule
		{"not_ignored", false},            // Not matched by any rule
	}

	for _, test := range tests {
		isIgnored := ignore.IsIgnored(test.file)

		if isIgnored != test.isIgnored {
			t.Errorf("Expected IsIgnored(%q) to be %v, got %v", test.file, test.isIgnored, isIgnored)
		}
	}
}

// Helper function to check if a string is present in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
