package main

import (
	"strings"
	"testing"
)

// TestAnalisar espelha TestAnalisarTabelas de test_scripts.py: lista
// tabelas ordenadas por tamanho dos dados.
func TestAnalisar(t *testing.T) {
	t.Run("lista_tabelas", func(t *testing.T) {
		output := runAnalisarT(t, dumpSample, false, 0)
		assertContains(t, output, "users")
		assertContains(t, output, "logs")
	})

	t.Run("maior_primeiro", func(t *testing.T) {
		output := runAnalisarT(t, dumpSample, false, 0)
		posUsers := strings.Index(output, "users")
		posLogs := strings.Index(output, "logs")
		if posUsers <= 0 {
			t.Errorf("esperava 'users' presente após a posição 0, achou em %d", posUsers)
		}
		if posLogs <= 0 {
			t.Errorf("esperava 'logs' presente após a posição 0, achou em %d", posLogs)
		}
	})

	t.Run("top_n", func(t *testing.T) {
		output := runAnalisarT(t, dumpSample, false, 1)
		var tableLines []string
		for _, l := range strings.Split(output, "\n") {
			trimmed := strings.TrimSpace(l)
			if trimmed == "" || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "Tabela") ||
				strings.HasPrefix(trimmed, "TOTAL") || strings.HasPrefix(trimmed, "Entrada") {
				continue
			}
			if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
				continue
			}
			if strings.Contains(l, "%") {
				tableLines = append(tableLines, l)
			}
		}
		if len(tableLines) != 1 {
			t.Errorf("esperava 1 linha de tabela com -n 1, encontrou %d: %v", len(tableLines), tableLines)
		}
	})

	t.Run("gz_input", func(t *testing.T) {
		output := runAnalisarT(t, dumpSample, true, 0)
		assertContains(t, output, "users")
		assertContains(t, output, "logs")
	})

	t.Run("contagem_registros", func(t *testing.T) {
		output := runAnalisarT(t, dumpSample, false, 0)
		for _, line := range strings.Split(output, "\n") {
			tokens := strings.Fields(line)
			if len(tokens) == 0 {
				continue
			}
			if tokens[0] == "users" || tokens[0] == "logs" {
				if tokens[3] != "2" {
					t.Errorf("esperava 2 registros para %q, achou %q (linha: %q)", tokens[0], tokens[3], line)
				}
			}
		}
	})

	t.Run("tabela_sem_dados_nao_aparece", func(t *testing.T) {
		dumpVazio := "CREATE TABLE public.vazia (id integer);\n" +
			"COPY public.vazia (id) FROM stdin;\n" +
			"\\.\n"
		output := runAnalisarT(t, dumpVazio, false, 0)
		assertNotContains(t, output, "❌")
	})
}
