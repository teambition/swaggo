package parser

import (
	"go/ast"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// fileExists check if the file existed
func fileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// str2RealType convert string type to inner type by `typ`
func str2RealType(s string, typ string) (ret interface{}, err error) {
	switch typ {
	case "int", "int64", "int32", "int16", "int8":
		ret, err = strconv.Atoi(s)
	case "bool":
		ret, err = strconv.ParseBool(s)
	case "float64":
		ret, err = strconv.ParseFloat(s, 64)
	case "float32":
		ret, err = strconv.ParseFloat(s, 32)
	default:
		ret = s
	}
	return
}

func absPathFromGoPath(importPath, vendor string) (string, bool) {
	vendorImport := filepath.Join(vendor, importPath)
	if fileExists(vendorImport) {
		return vendorImport, true
	}
	// check GOPATH
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		panic("GOPATH environment variable is not set or empty")
	}
	// find absolute path
	for _, goPath := range filepath.SplitList(goPaths) {
		wg, _ := filepath.EvalSymlinks(filepath.Join(goPath, "src", importPath))
		if fileExists(wg) {
			return wg, true
		}
	}
	return "", false
}

func absPathFromGoRoot(importPath string) (string, bool) {
	goRoot := runtime.GOROOT()
	if goRoot == "" {
		panic("GOROOT environment variable is not set or empty")
	}
	wg, _ := filepath.EvalSymlinks(filepath.Join(goRoot, "src", importPath))
	if fileExists(wg) {
		return wg, true
	}
	return "", false
}

// tagTrimPrefixAndSpace if prefix existed then trim it and trim space
func tagTrimPrefixAndSpace(s *string, prefix string) bool {
	existed := strings.HasPrefix(*s, prefix)
	if existed {
		*s = strings.TrimPrefix(*s, prefix)
		*s = strings.TrimSpace(*s)
	}
	return existed
}

// isDocComments check if comments has `@` prefix
func isDocComments(comments *ast.CommentGroup) bool {
	for _, c := range strings.Split(comments.Text(), "\n") {
		if strings.HasPrefix(c, docPrefix) {
			return true
		}
	}
	return false
}

// getparams analisys params return []string
// @Param query form string true "The email for login"
// @Success 200 string "Some Success"
// @Failure 400 string "Some Failure"
func getparams(str string) []string {
	var s []rune
	var j int
	var start bool
	var r []string
	var quoted int8
	for _, c := range []rune(str) {
		if unicode.IsSpace(c) && quoted == 0 {
			if !start {
				continue
			} else {
				start = false
				j++
				r = append(r, string(s))
				s = make([]rune, 0)
				continue
			}
		}

		start = true
		if c == '"' {
			quoted ^= 1
			continue
		}
		s = append(s, c)
	}
	if len(s) > 0 {
		r = append(r, string(s))
	}
	return r
}

// contentTypeByDoc Get content types from comment
func contentTypeByDoc(s string) []string {
	result := []string{}
	tmp := strings.Split(s, ",")
	for _, v := range tmp {
		result = append(result, contentType[v])
	}
	return result
}

// subset returns true if the first array is completely
// contained in the second array. There must be at least
// the same number of duplicate values in second as there
// are in first.
func subset(first, second []string) bool {
	set := make(map[string]struct{})
	for _, value := range second {
		set[value] = struct{}{}
	}

	for _, value := range first {
		if _, found := set[value]; !found {
			return false
		}
	}
	return true
}
