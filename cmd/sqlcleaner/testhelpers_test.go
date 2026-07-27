package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// neutralTempDir cria um diretório temporário com nome neutro (não
// baseado no nome do teste/subteste), para evitar que o caminho do
// arquivo apareça nos relatórios de stdout e contamine buscas por
// substring como "Aviso".
func neutralTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sqlcleaner-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func writeTempFile(t *testing.T, content string, gz bool) string {
	t.Helper()
	dir := neutralTempDir(t)
	suffix := ".sql"
	if gz {
		suffix = ".sql.gz"
	}
	path := filepath.Join(dir, "input"+suffix)
	if gz {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		gw := gzip.NewWriter(f)
		if _, err := gw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := gw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func tempOutputPath(t *testing.T, gz bool) string {
	t.Helper()
	dir := neutralTempDir(t)
	suffix := ".sql"
	if gz {
		suffix = ".sql.gz"
	}
	return filepath.Join(dir, "output"+suffix)
}

func readOutputFile(t *testing.T, path string, gz bool) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !gz {
		return string(data)
	}
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("esperava que contivesse %q, mas não continha.\n--- conteúdo ---\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("esperava que NÃO contivesse %q, mas continha.\n--- conteúdo ---\n%s", substr, s)
	}
}

// --- runners por subcomando ---

func runEsvaziarT(t *testing.T, content, tables string, inGz, outGz bool) string {
	t.Helper()
	in := writeTempFile(t, content, inGz)
	out := tempOutputPath(t, outGz)
	var buf bytes.Buffer
	if err := runEsvaziar(&buf, in, out, splitCommaList(tables)); err != nil {
		t.Fatalf("runEsvaziar: %v", err)
	}
	return readOutputFile(t, out, outGz)
}

func runEsvaziarStdoutT(t *testing.T, content, tables string) string {
	t.Helper()
	in := writeTempFile(t, content, false)
	out := tempOutputPath(t, false)
	var buf bytes.Buffer
	if err := runEsvaziar(&buf, in, out, splitCommaList(tables)); err != nil {
		t.Fatalf("runEsvaziar: %v", err)
	}
	return buf.String()
}

func runRemoverT(t *testing.T, content, tables string, inGz, outGz bool) string {
	t.Helper()
	in := writeTempFile(t, content, inGz)
	out := tempOutputPath(t, outGz)
	var buf bytes.Buffer
	if err := runRemover(&buf, in, out, splitCommaList(tables)); err != nil {
		t.Fatalf("runRemover: %v", err)
	}
	return readOutputFile(t, out, outGz)
}

func runRemoverStdoutT(t *testing.T, content, tables string) string {
	t.Helper()
	in := writeTempFile(t, content, false)
	out := tempOutputPath(t, false)
	var buf bytes.Buffer
	if err := runRemover(&buf, in, out, splitCommaList(tables)); err != nil {
		t.Fatalf("runRemover: %v", err)
	}
	return buf.String()
}

func runTruncarT(t *testing.T, content, tablesArg string, limit *int, inGz, outGz bool) string {
	t.Helper()
	in := writeTempFile(t, content, inGz)
	out := tempOutputPath(t, outGz)
	var buf bytes.Buffer
	if err := runTruncar(&buf, in, out, tablesArg, limit); err != nil {
		t.Fatalf("runTruncar: %v", err)
	}
	return readOutputFile(t, out, outGz)
}

func runTruncarStdoutT(t *testing.T, content, tablesArg string, limit *int) string {
	t.Helper()
	in := writeTempFile(t, content, false)
	out := tempOutputPath(t, false)
	var buf bytes.Buffer
	if err := runTruncar(&buf, in, out, tablesArg, limit); err != nil {
		t.Fatalf("runTruncar: %v", err)
	}
	return buf.String()
}

func runSetT(t *testing.T, content, params string, inGz, outGz bool) (output, stdout string) {
	t.Helper()
	in := writeTempFile(t, content, inGz)
	out := tempOutputPath(t, outGz)
	var buf bytes.Buffer
	if err := runSet(&buf, in, out, params); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	return readOutputFile(t, out, outGz), buf.String()
}

func runAnalisarT(t *testing.T, content string, inGz bool, top int) string {
	t.Helper()
	in := writeTempFile(t, content, inGz)
	var buf bytes.Buffer
	if err := runAnalisar(&buf, in, top); err != nil {
		t.Fatalf("runAnalisar: %v", err)
	}
	return buf.String()
}

func intPtr(n int) *int { return &n }
