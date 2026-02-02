package pathspec

import (
	"regexp"
	"strings"
)

// A Pathspec matches full path strings
// a pathspec of 'directory/subdirectory'
// will return the following results
//
//	'directory/subdirectory/file1' true
//	'directory/subdirectory/file2' true
//	'directory/subdirectory/file2' true
//	'directory/subdirectory'       true
//	'directory/subdirector'        false
//	'directory'                    false
//
// for a pathspec of 'directory/*'
//
//	'directory/subdirectory' true
type Pathspec struct {
	pattern string
	regex   []*regexp.Regexp
}

func Compile(pattern string) (pathspec *Pathspec, err error) {
	pathspec = new(Pathspec)

	if pattern == "" {
		return
	}

	// remove trailing separator
	pattern = strings.TrimSuffix(pattern, "/\\")
	// tokenize by the presence of wildcards
	parts := strings.Split(pattern, "*")

	// produce a regex that will match individual files/directories
	var file_regex_source strings.Builder
	file_regex_source.WriteString("^")
	// join wildcard splits and process with QuoteMeta
	// effectively we're sanitizing everything that's not a wildcard
	for i := range parts {
		file_regex_source.WriteString(regexp.QuoteMeta(parts[i]))
		if i < len(parts)-1 {
			file_regex_source.WriteString(".*")
		}
	}
	basic_regex_source := file_regex_source.String()
	// do not match anything else
	file_regex_source.WriteString("$")
	var file_regex *regexp.Regexp
	file_regex, err = regexp.Compile(file_regex_source.String())
	if err != nil {
		return
	}
	pathspec.regex = append(pathspec.regex, file_regex)

	// produce a regex that
	// e.g. if pattern is 'directory'
	// regex would be '^directory/*$'
	var directory_regex_source strings.Builder
	directory_regex_source.WriteString(basic_regex_source)
	directory_regex_source.WriteString("/.*$")
	var directory_regex *regexp.Regexp
	directory_regex, err = regexp.Compile(directory_regex_source.String())
	if err != nil {
		return
	}
	pathspec.regex = append(pathspec.regex, directory_regex)

	// fmt.Println(file_regex_source.String())
	// fmt.Println(directory_regex_source.String())
	return
}

func MustCompile(pattern string) (pathspec *Pathspec) {
	var err error
	pathspec, err = Compile(pattern)
	if err != nil {
		panic(err)
	}
	return
}

func (pathspec *Pathspec) MatchString(str string) (matched bool) {
	for _, regex := range pathspec.regex {
		if regex.MatchString(str) {
			matched = true
			break
		}
	}
	return
}
