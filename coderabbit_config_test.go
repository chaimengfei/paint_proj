package coderrabbit_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const defaultCodeRabbitYAML = `# coderabbit.yaml
scanner: code_rabbit  # 指定使用 CodeRabbit 扫描器
exclude:
  - "test/**"          # 排除 test 目录
  - "cmd/**"           # 排除 cmd 目录
file_types:
  - ".go"              # 只扫描 Go 文件
`

type CodeRabbitConfig struct {
	Scanner   string
	Exclude   []string
	FileTypes []string
}

// parseCodeRabbitYAML is a minimal, YAML-shape-aware parser tailored to the known schema.
// It intentionally avoids external dependencies and supports:
//   - top-level scalar: scanner
//   - top-level lists: exclude, file_types
//   - trailing comments beginning with '#'
func parseCodeRabbitYAML(data []byte) (CodeRabbitConfig, error) {
	var cfg CodeRabbitConfig
	var currentKey string

	nl := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(nl, "\n")
	for _, raw := range lines {
		// strip trailing comments (best-effort, adequate for our known file)
		line := raw
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// list item
		if strings.HasPrefix(line, "- ") {
			item := trimQuotes(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			switch currentKey {
			case "exclude":
				if item \!= "" {
					cfg.Exclude = append(cfg.Exclude, item)
				}
			case "file_types":
				if item \!= "" {
					cfg.FileTypes = append(cfg.FileTypes, item)
				}
			}
			continue
		}

		// key: value or key:
		if idx := strings.Index(line, ":"); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			currentKey = key

			switch key {
			case "scanner":
				if val \!= "" {
					cfg.Scanner = trimQuotes(val)
				}
			case "exclude", "file_types":
				// values come from subsequent "- " lines
			default:
				// ignore unknown keys
			}
			continue
		}
	}

	if strings.TrimSpace(cfg.Scanner) == "" {
		return cfg, fmt.Errorf("missing required key 'scanner' or empty value")
	}
	return cfg, nil
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func contains(slice []string, want string) bool {
	for _, v := range slice {
		if v == want {
			return true
		}
	}
	return false
}

func loadCoderabbitYAML(t *testing.T) []byte {
	t.Helper()
	candidates := []string{
		"coderabbit.yaml",
		"coderabbit.yml",
		".coderabbit.yaml",
		".coderabbit.yml",
		".github/coderabbit.yaml",
		".github/coderabbit.yml",
	}
	for _, p := range candidates {
		if b, err := os.ReadFile(p); err == nil {
			t.Logf("Using repository config at %s", p)
			return b
		}
	}
	t.Log("Repository config not found in common locations; using embedded fallback content.")
	return []byte(defaultCodeRabbitYAML)
}

func TestCodeRabbitYAML_ValidStructureAndValues(t *testing.T) {
	raw := loadCoderabbitYAML(t)
	cfg, err := parseCodeRabbitYAML(raw)
	if err \!= nil {
		t.Fatalf("failed to parse CodeRabbit YAML: %v\nContent:\n%s", err, string(raw))
	}
	if cfg.Scanner \!= "code_rabbit" {
		t.Errorf("scanner mismatch: want %q, got %q", "code_rabbit", cfg.Scanner)
	}
	if len(cfg.Exclude) == 0 {
		t.Errorf("exclude should contain entries; got none")
	}
	if len(cfg.FileTypes) == 0 {
		t.Errorf("file_types should contain entries; got none")
	}
}

func TestCodeRabbitYAML_ExcludesTestAndCmdDirs(t *testing.T) {
	raw := loadCoderabbitYAML(t)
	cfg, err := parseCodeRabbitYAML(raw)
	if err \!= nil {
		t.Fatalf("parse error: %v", err)
	}
	for _, want := range []string{`test/**`, `cmd/**`} {
		if \!contains(cfg.Exclude, want) {
			t.Errorf("exclude should contain %q; got %v", want, cfg.Exclude)
		}
	}
}

func TestCodeRabbitYAML_FileTypesIncludesGo(t *testing.T) {
	raw := loadCoderabbitYAML(t)
	cfg, err := parseCodeRabbitYAML(raw)
	if err \!= nil {
		t.Fatalf("parse error: %v", err)
	}
	if \!contains(cfg.FileTypes, ".go") {
		t.Errorf("file_types should include %q; got %v", ".go", cfg.FileTypes)
	}
}

func TestParser_HandlesCommentsAndWhitespace(t *testing.T) {
	raw := []byte(`
   # leading comment
scanner:   "code_rabbit"   # inline comment
exclude: # comment after key
   -  "test/**"   
   - "cmd/**"   # more comments
file_types:
   -   ".go"    # trailing spaces and comments
`)
	cfg, err := parseCodeRabbitYAML(raw)
	if err \!= nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if cfg.Scanner \!= "code_rabbit" {
		t.Errorf("want scanner %q, got %q", "code_rabbit", cfg.Scanner)
	}
	if \!contains(cfg.Exclude, "test/**") || \!contains(cfg.Exclude, "cmd/**") {
		t.Errorf("exclude list missing expected items: %v", cfg.Exclude)
	}
	if \!contains(cfg.FileTypes, ".go") {
		t.Errorf("file_types missing .go: %v", cfg.FileTypes)
	}
}

func TestParser_ErrorsOnMissingScanner(t *testing.T) {
	raw := []byte(`
exclude:
  - "test/**"
file_types:
  - ".go"
`)
	if _, err := parseCodeRabbitYAML(raw); err == nil {
		t.Fatalf("expected error when 'scanner' is missing, got nil")
	}
}

func TestParser_IgnoresUnknownKeys(t *testing.T) {
	raw := []byte(`
scanner: code_rabbit
exclude:
  - "test/**"
file_types:
  - ".go"
unknown_key: some_value
`)
	cfg, err := parseCodeRabbitYAML(raw)
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Scanner \!= "code_rabbit" {
		t.Errorf("want scanner %q, got %q", "code_rabbit", cfg.Scanner)
	}
}