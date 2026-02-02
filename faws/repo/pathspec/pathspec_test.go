package pathspec

import "testing"

type test_case_match struct {
	str    string
	result bool
}

type test_case struct {
	pattern string
	matches []test_case_match
}

var test_cases = []test_case{
	// basic
	{"directory", []test_case_match{
		{"directory/child1", true},
		{"directory/child2", true},
		{"directory/", true},
		{"directory", true},
		{"director", false},
		{"", false},
	}},
	// * wildcard
	{"directory/*", []test_case_match{
		{"directory/child1", true},
		{"directory/child2", true},
		{"directory/", true},
		{"director", false},
		{"", false},
	}},
	// * wildcard
	{"directory/*/child*", []test_case_match{
		{"directory/1/child1", true},
		{"directory/2/child2", true},
		{"directory/child3", false},
		{"director", false},
		{"", false},
	}},
}

func TestPathspec(t *testing.T) {
	for test_case_index, test_case := range test_cases {
		pathspec := MustCompile(test_case.pattern)
		for match_index, match := range test_case.matches {
			was := pathspec.MatchString(match.str)
			should_have_been := match.result
			if was != should_have_been {
				t.Fatalf("test case %d: match %d (%s): result should have been %t, was instead %t", test_case_index, match_index, match.str, should_have_been, was)
			}
		}
	}
}
