//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Reference struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Key      string `json:"key"`
	Function string `json:"function"`
	Context  string `json:"context"`
}

type Result struct {
	Method     string      `json:"method"`
	Signature  string      `json:"signature"`
	Definition string      `json:"definition"`
	References []Reference `json:"references"`
	Total      int         `json:"total"`
}

var basePath = "/Users/coissac/Sync/travail/__MOI__/GO/obitools4"

func main() {
	cmd := exec.Command("rg", "-n", `\.SetAttribute\(`, basePath+"/pkg", "--type", "go")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running rg: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(output), "\n")
	lineRe := regexp.MustCompile(`^(.+?):(\d+):\s*(.+)$`)
	keyRe := regexp.MustCompile(`SetAttribute\("([^"]+)"`)
	templateKeyRe := regexp.MustCompile(`SetAttribute\("([^"]+)[^"]*"\s*,`)

	var refs []Reference
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		file := matches[1]
		lineNum, _ := strconv.Atoi(matches[2])
		context := strings.TrimSpace(matches[3])

		// Skip definition
		if strings.Contains(file, "obiseq/attributes.go") && lineNum == 132 {
			continue
		}

		// Extract key
		var key string
		if keyMatches := keyRe.FindStringSubmatch(context); keyMatches != nil {
			key = keyMatches[1]
		} else if tmplMatches := templateKeyRe.FindStringSubmatch(context); tmplMatches != nil {
			key = tmplMatches[1]
		} else {
			continue
		}

		// Get function name using treesitter
		funcName := getFunctionNameTreesitter(file, lineNum)

		uniqueKey := fmt.Sprintf("%s:%d", file, lineNum)
		if seen[uniqueKey] {
			continue
		}
		seen[uniqueKey] = true

		refs = append(refs, Reference{
			File:     filepath.Base(file),
			Line:     lineNum,
			Column:   0,
			Key:      key,
			Function: funcName,
			Context:  context,
		})
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].File != refs[j].File {
			return refs[i].File < refs[j].File
		}
		return refs[i].Line < refs[j].Line
	})

	result := Result{
		Method:     "SetAttribute",
		Signature:  "func (s *BioSequence) SetAttribute(key string, value interface{})",
		Definition: basePath + "/pkg/obiseq/attributes.go:132",
		References: refs,
		Total:      len(refs),
	}

	outputJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(outputJSON))
}

// getFunctionNameTreesitter uses the treesitter_cursor_walk tool to get the containing function
func getFunctionNameTreesitter(file string, targetLine int) string {
	// Convert to 0-based for treesitter
	row := targetLine - 1

	// Use treesitter cursor walk to get ancestors
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(`kilo treesitter_cursor_walk --file_path %q --row %d --column 0 --max_depth 10 2>/dev/null`, file, row))

	output, err := cmd.Output()
	if err != nil {
		return findContainingFunction(file, targetLine)
	}

	// Parse the JSON output to find function_declaration or method_declaration
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return findContainingFunction(file, targetLine)
	}

	// Check ancestors for function declaration
	if ancestors, ok := result["ancestors"].([]interface{}); ok {
		for _, a := range ancestors {
			if anc, ok := a.(map[string]interface{}); ok {
				nodeType, _ := anc["type"].(string)
				if nodeType == "function_declaration" || nodeType == "method_declaration" {
					// Try to get the function name from children
					if children, ok := anc["children"].([]interface{}); ok {
						for _, c := range children {
							if child, ok := c.(map[string]interface{}); ok {
								childType, _ := child["type"].(string)
								if childType == "identifier" {
									if text, ok := child["text"].(string); ok {
										return text
									}
								}
								if childType == "field_identifier" {
									if text, ok := child["text"].(string); ok {
										return text
									}
								}
							}
						}
					}
				}
				if nodeType == "func_literal" {
					return "closure"
				}
			}
		}
	}

	return findContainingFunction(file, targetLine)
}

func findContainingFunction(file string, targetLine int) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")

	for i := targetLine - 1; i >= 0 && i >= targetLine-200; i-- {
		if i >= len(lines) {
			continue
		}
		line := strings.TrimSpace(lines[i])

		if line == "}" && i > 0 {
			for j := i - 1; j >= 0 && j >= i-50; j-- {
				if j >= len(lines) {
					continue
				}
				funcLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(funcLine, "func ") {
					if match := regexp.MustCompile(`func\s+\([^)]+\)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(funcLine); match != nil {
						return match[1]
					}
					if match := regexp.MustCompile(`func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(funcLine); match != nil {
						return match[1]
					}
				}
			}
			continue
		}

		if strings.HasPrefix(line, "func ") {
			if match := regexp.MustCompile(`func\s+\([^)]+\)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(line); match != nil {
				return match[1]
			}
			if match := regexp.MustCompile(`func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(line); match != nil {
				return match[1]
			}
		}
	}

	return ""
}
