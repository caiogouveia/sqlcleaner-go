package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
)

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// formatThousands formata um inteiro com vírgula como separador de milhar,
// igual ao f"{n:,}" do Python.
func formatThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var groups []string
	for len(s) > 3 {
		groups = append([]string{s[len(s)-3:]}, groups...)
		s = s[:len(s)-3]
	}
	groups = append([]string{s}, groups...)
	out := strings.Join(groups, ",")
	if neg {
		out = "-" + out
	}
	return out
}

func indexOf(parts []string, target string) int {
	for i, p := range parts {
		if p == target {
			return i
		}
	}
	return -1
}

func parseTargetTables(tables []string) []sqldump.QualifiedName {
	targets := make([]sqldump.QualifiedName, len(tables))
	for i, t := range tables {
		targets[i] = sqldump.ParseQualifiedName(strings.TrimSpace(t))
	}
	return targets
}

// unmatchedNames devolve, na ordem original de targets, os nomes que não
// tiveram nenhuma correspondência no dump.
func unmatchedNames(targets []sqldump.QualifiedName, matched map[sqldump.QualifiedName]bool) []string {
	var out []string
	for _, t := range targets {
		if !matched[t] {
			out = append(out, t.String())
		}
	}
	return out
}

// splitCommaList separa por vírgula e remove espaços, igual a
// "[t.strip() for t in s.split(',')]" do Python.
func splitCommaList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}
