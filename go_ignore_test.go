package go_ignore

import (
	"reflect"
	"regexp"
	"testing"
)

func TestCheckPath(t *testing.T) {
	// Define the expected rules based on the provided .gitignore file
	expectedRules := &Rules{
		Directory: []*regexp.Regexp{
			regexp.MustCompile(`^ignored_dir/`),
			regexp.MustCompile(`^docs/`),
			regexp.MustCompile(`^ignored_file\.txt$`),
		},
		Character: []*regexp.Regexp{
			regexp.MustCompile(`\.txt$`),
			regexp.MustCompile(`^test_file\.md$`),
		},
		Subdir: []*regexp.Regexp{
			regexp.MustCompile(`^src/.*\.js$`),
		},
		Negate: []*regexp.Regexp{
			regexp.MustCompile(`^!important\.txt$`),
		},
	}

	// Specify the directory containing the .gitignore file
	testDir := "./test"

	// Call the CheckPath function
	gitIgnore, err := CheckPath(testDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		return
	}

	// Compare the rules
	if !reflect.DeepEqual(gitIgnore.Rules, expectedRules) {
		t.Errorf("Expected rules to be %v, got %v", expectedRules, gitIgnore.Rules)
	}
}

func TestParseGitignore(t *testing.T) {
	// Specify the path to the valid .gitignore file
	validGitignorePath := "./test/.gitignore"

	// Define the expected rules based on the valid .gitignore file
	expectedRules := &Rules{
		Directory: []*regexp.Regexp{
			regexp.MustCompile(`^ignored_dir/`),
			regexp.MustCompile(`^docs/`),
			regexp.MustCompile(`^ignored_file\.txt$`),
		},
		Character: []*regexp.Regexp{
			regexp.MustCompile(`\.txt$`),
			regexp.MustCompile(`^test_file\.md$`),
		},
		Subdir: []*regexp.Regexp{
			regexp.MustCompile(`^src/.*\.js$`),
		},
		Negate: []*regexp.Regexp{
			regexp.MustCompile(`^!important\.txt$`),
		},
	}

	// Call the parseGitignore function
	gitIgnore, err := parseGitignore(validGitignorePath)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
		return
	}

	// Compare the rules
	if !reflect.DeepEqual(gitIgnore.Rules, expectedRules) {
		t.Errorf("Expected rules to be %v, got %v", expectedRules, gitIgnore.Rules)
	}

	// Specify the path to the invalid .gitignore file
	invalidGitignorePath := "./testdata/invalid_gitignore"

	// Call the parseGitignore function for the invalid file
	_, err = parseGitignore(invalidGitignorePath)
	if err == nil {
		t.Errorf("Expected an error, got nil")
		return
	}
}

func TestIsIgnored(t *testing.T) {
	var err error
	// Create a GitIgnore object with the expected rules
	rules := &Rules{
		Directory: []*regexp.Regexp{
			regexp.MustCompile(`^ignored_dir/`),
			regexp.MustCompile(`^docs/`),
			regexp.MustCompile(`^ignored_file\.txt$`),
		},
		Character: []*regexp.Regexp{
			regexp.MustCompile(`\.txt$`),
			regexp.MustCompile(`^test_file\.md$`),
		},
		Subdir: []*regexp.Regexp{
			regexp.MustCompile(`^src/.*\.js$`),
		},
		Negate: []*regexp.Regexp{
			regexp.MustCompile(`^!important\.txt$`),
		},
	}
	gitIgnore := &GitIgnore{
		Rules: rules,
	}
	gitIgnore, err = parseGitignore("./test/.gitignore")
	if err != nil {
		t.Error(ParsingGitignoreError{})
	}

	// Test files
	testFiles := []string{
		"not_ignored.txt",
		"ignored_file.txt",
		"important.txt",
	}

	// Specify the expected results
	expectedResults := []bool{false, true, true}

	// Test the IsIgnored function
	for i, file := range testFiles {
		result := gitIgnore.IsIgnored(file)
		if result != expectedResults[i] {
			t.Errorf("Expected IsIgnored(%q) to be %v, got %v", file, expectedResults[i], result)
		}
	}

	// Check the ignored files
	expectedIgnoredFiles := []string{"ignored_file.txt", "important.txt"}
	if !reflect.DeepEqual(gitIgnore.IgnoredFiles, expectedIgnoredFiles) {
		t.Errorf("Expected ignored files to be %v, got %v", expectedIgnoredFiles, gitIgnore.IgnoredFiles)
	}
}
