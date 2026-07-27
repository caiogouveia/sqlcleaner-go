package main

import (
	"strings"
	"testing"
)

// payloadLines extrai a coluna de payload (e1..e5) das linhas com tab,
// preservando a ordem — usado para verificar preservação de ordem relativa.
func payloadLines(output string) []string {
	var out []string
	for _, l := range strings.Split(output, "\n") {
		if !strings.Contains(l, "\t") {
			continue
		}
		parts := strings.SplitN(l, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[1] {
		case "e1", "e2", "e3", "e4", "e5":
			out = append(out, parts[1])
		}
	}
	return out
}

func assertPayloadEquals(t *testing.T, output string, want []string) {
	t.Helper()
	got := payloadLines(output)
	if len(got) != len(want) {
		t.Fatalf("esperava %v, obteve %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("esperava %v, obteve %v", want, got)
		}
	}
}

// TestTruncar espelha TestTruncarTabelas de test_scripts.py: mantém apenas
// N registros por tabela.
func TestTruncar(t *testing.T) {
	t.Run("limite_unico_mantem_n_registros", func(t *testing.T) {
		// logs tem 2 registros; manter 1
		output := runTruncarT(t, dumpSample, "logs", intPtr(1), false, false)
		assertContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})

	t.Run("limite_unico_zero_remove_todos", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs", intPtr(0), false, false)
		assertNotContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})

	t.Run("limite_maior_que_total_mantem_tudo", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs", intPtr(999), false, false)
		assertContains(t, output, "startup")
		assertContains(t, output, "shutdown")
	})

	t.Run("limite_por_tabela", func(t *testing.T) {
		// users tem 2 registros; manter 1 de users e 1 de logs
		output := runTruncarT(t, dumpSample, "users:1,logs:1", nil, false, false)
		assertContains(t, output, "Alice")
		assertNotContains(t, output, "Bob")
		assertContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})

	t.Run("mistura_explicito_e_padrao", func(t *testing.T) {
		// logs com limite explícito, users usa padrão -n
		output := runTruncarT(t, dumpSample, "logs:1,users", intPtr(1), false, false)
		assertContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
		assertContains(t, output, "Alice")
		assertNotContains(t, output, "Bob")
	})

	t.Run("estrutura_preservada", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs", intPtr(1), false, false)
		assertContains(t, output, "CREATE TABLE public.logs")
		assertContains(t, output, "CREATE INDEX idx_logs_id")
	})

	t.Run("tabela_nao_alvo_intacta", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs", intPtr(1), false, false)
		assertContains(t, output, "Alice")
		assertContains(t, output, "Bob")
	})

	t.Run("gz_in_gz_out", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs", intPtr(1), true, true)
		assertContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})

	t.Run("desc_mantem_ultimos_registros", func(t *testing.T) {
		// logs tem 2 registros; DESC com limite 1 mantém o último (shutdown)
		output := runTruncarT(t, dumpSample, "logs:1:DESC", nil, false, false)
		assertContains(t, output, "shutdown")
		assertNotContains(t, output, "startup")
	})

	t.Run("asc_e_desc_combinados", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "users:1,logs:1:DESC", nil, false, false)
		assertContains(t, output, "Alice")
		assertNotContains(t, output, "Bob")
		assertContains(t, output, "shutdown")
		assertNotContains(t, output, "startup")
	})

	t.Run("desc_sem_limite_explicito_usa_padrao", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs:DESC", intPtr(1), false, false)
		assertContains(t, output, "shutdown")
		assertNotContains(t, output, "startup")
	})

	t.Run("desc_limite_zero_remove_todos", func(t *testing.T) {
		output := runTruncarT(t, dumpSample, "logs:0:DESC", nil, false, false)
		assertNotContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})

	t.Run("ordem_invalida_erro", func(t *testing.T) {
		stdout := runTruncarStdoutT(t, dumpSample, "logs:1:XYZ", nil)
		assertContains(t, stdout, "Erro")
	})

	t.Run("sem_limite_e_sem_n_erro", func(t *testing.T) {
		stdout := runTruncarStdoutT(t, dumpSample, "logs", nil)
		assertContains(t, stdout, "Erro")
	})

	t.Run("asc_preserva_ordem_relativa", func(t *testing.T) {
		// events tem 5 linhas (e1..e5); ASC com limite 3 mantém e1,e2,e3 nessa ordem
		output := runTruncarT(t, dumpEvents, "events:3", nil, false, false)
		assertPayloadEquals(t, output, []string{"e1", "e2", "e3"})
	})

	t.Run("desc_preserva_ordem_relativa", func(t *testing.T) {
		// events tem 5 linhas (e1..e5); DESC com limite 3 mantém e3,e4,e5 nessa ordem
		output := runTruncarT(t, dumpEvents, "events:3:DESC", nil, false, false)
		assertPayloadEquals(t, output, []string{"e3", "e4", "e5"})
	})

	t.Run("desc_limite_maior_que_total_mantem_tudo_em_ordem", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "events:999:DESC", nil, false, false)
		assertPayloadEquals(t, output, []string{"e1", "e2", "e3", "e4", "e5"})
	})

	t.Run("desc_limite_igual_ao_total_mantem_tudo", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "events:5:DESC", nil, false, false)
		assertPayloadEquals(t, output, []string{"e1", "e2", "e3", "e4", "e5"})
	})

	t.Run("desc_ordem_case_insensitive", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "events:2:desc", nil, false, false)
		assertPayloadEquals(t, output, []string{"e4", "e5"})
	})

	t.Run("asc_explicito_case_insensitive", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "events:2:asc", nil, false, false)
		assertPayloadEquals(t, output, []string{"e1", "e2"})
	})

	t.Run("desc_com_prefixo_public", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "public.events:2:DESC", nil, false, false)
		assertPayloadEquals(t, output, []string{"e4", "e5"})
	})

	t.Run("desc_tabela_vazia_nao_trava", func(t *testing.T) {
		// empty_table não tem nenhuma linha de dado; DESC não deve travar nem gerar erro
		output := runTruncarT(t, dumpEvents, "empty_table:5:DESC", nil, false, false)
		assertContains(t, output, "COPY public.empty_table (id) FROM stdin;")
		assertContains(t, output, `\.`)
	})

	t.Run("desc_gz_in_gz_out", func(t *testing.T) {
		output := runTruncarT(t, dumpEvents, "events:3:DESC", nil, true, true)
		assertPayloadEquals(t, output, []string{"e3", "e4", "e5"})
	})

	t.Run("desc_nao_afeta_outras_tabelas", func(t *testing.T) {
		// DESC em events não deve alterar users/logs (ASC padrão) no mesmo comando
		output := runTruncarT(t, dumpSample+dumpEvents, "events:2:DESC,users:1", nil, false, false)
		assertContains(t, output, "Alice")
		assertNotContains(t, output, "Bob")
		assertPayloadEquals(t, output, []string{"e4", "e5"})
	})

	t.Run("multiplas_tabelas_desc_com_limites_diferentes", func(t *testing.T) {
		// duas tabelas em DESC, cada uma com seu próprio limite — buffers não podem vazar entre blocos
		content := dumpSample + dumpEvents
		output := runTruncarT(t, content, "logs:1:DESC,events:2:DESC", nil, false, false)
		assertContains(t, output, "shutdown")
		assertNotContains(t, output, "startup")
		assertPayloadEquals(t, output, []string{"e4", "e5"})
	})

	t.Run("formato_com_dois_pontos_demais_erro", func(t *testing.T) {
		stdout := runTruncarStdoutT(t, dumpSample, "logs:1:DESC:extra", nil)
		assertContains(t, stdout, "Erro")
	})

	t.Run("relatorio_conta_linhas_removidas_desc", func(t *testing.T) {
		// events tem 5 linhas; DESC com limite 2 deve reportar 3 linhas removidas
		stdout := runTruncarStdoutT(t, dumpEvents, "events:2:DESC", nil)
		assertContains(t, stdout, "3 removidas")
	})
}
