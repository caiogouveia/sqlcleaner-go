package sqldump

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenWriterOpenReaderPlano(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "arquivo.sql")

	w, err := OpenWriter(path)
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	if _, err := w.Write([]byte("conteudo de teste\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close (writer): %v", err)
	}

	r, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "conteudo de teste\n" {
		t.Errorf("conteúdo lido = %q, want %q", string(data), "conteudo de teste\n")
	}
}

func TestOpenWriterOpenReaderGz(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "arquivo.sql.gz")

	w, err := OpenWriter(path)
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	if _, err := w.Write([]byte("conteudo comprimido\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close (writer): %v", err)
	}

	// Confere que o arquivo gravado é de fato gzip (não texto puro).
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(raw) < 2 || raw[0] != 0x1f || raw[1] != 0x8b {
		t.Fatalf("arquivo não parece gzip: % x", raw[:min(len(raw), 4)])
	}

	r, err := OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "conteudo comprimido\n" {
		t.Errorf("conteúdo lido = %q, want %q", string(data), "conteudo comprimido\n")
	}
}

func TestOpenReaderArquivoInexistente(t *testing.T) {
	_, err := OpenReader("/caminho/que/nao/existe.sql")
	if err == nil {
		t.Fatal("esperava erro para arquivo inexistente")
	}
}

func TestLineScanner(t *testing.T) {
	t.Run("linhas_com_e_sem_terminador_final", func(t *testing.T) {
		s := NewLineScanner(strings.NewReader("linha1\nlinha2\nlinha3"))
		var got []string
		for {
			line, ok := s.Next()
			if !ok {
				break
			}
			got = append(got, line)
		}
		want := []string{"linha1\n", "linha2\n", "linha3"}
		if len(got) != len(want) {
			t.Fatalf("got %d linhas, want %d (got=%q)", len(got), len(want), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("linha[%d] = %q, want %q", i, got[i], want[i])
			}
		}
		if err := s.Err(); err != nil {
			t.Errorf("Err() = %v, want nil", err)
		}
	})

	t.Run("entrada_vazia", func(t *testing.T) {
		s := NewLineScanner(strings.NewReader(""))
		if _, ok := s.Next(); ok {
			t.Error("esperava ok=false para entrada vazia")
		}
		if err := s.Err(); err != nil {
			t.Errorf("Err() = %v, want nil", err)
		}
	})
}
