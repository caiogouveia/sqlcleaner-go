package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRestaurarDryRunNaoAcessaRede(t *testing.T) {
	in := writeTempFile(t, dumpSample, false)
	var buf bytes.Buffer
	// --dry-run só indexa e resolve a DSN (sem abrir conexão), então roda
	// mesmo sem um Postgres disponível.
	err := runRestaurar(&buf, strings.NewReader(""), in, restaurarOptions{
		dsn:    "postgres://user:pass@127.0.0.1:1/db",
		dryRun: true,
	})
	if err != nil {
		t.Fatalf("runRestaurar: %v", err)
	}
	out := buf.String()
	assertContains(t, out, "public.users")
	assertContains(t, out, "public.logs")
	assertContains(t, out, "--dry-run")
}

func TestRestaurarArquivoNaoEncontrado(t *testing.T) {
	var buf bytes.Buffer
	err := runRestaurar(&buf, strings.NewReader(""), "/caminho/inexistente.sql", restaurarOptions{
		dsn: "postgres://user:pass@127.0.0.1:1/db",
	})
	if err != nil {
		t.Fatalf("runRestaurar: %v", err)
	}
	assertContains(t, buf.String(), "não encontrado")
}

func TestRestaurarDsnInvalida(t *testing.T) {
	in := writeTempFile(t, dumpSample, false)
	var buf bytes.Buffer
	err := runRestaurar(&buf, strings.NewReader(""), in, restaurarOptions{
		dsn:    "isto não é uma dsn válida ://",
		dryRun: true,
	})
	if err != nil {
		t.Fatalf("runRestaurar: %v", err)
	}
	assertContains(t, buf.String(), "Erro")
}
