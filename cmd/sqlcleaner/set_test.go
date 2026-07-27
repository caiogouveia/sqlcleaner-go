package main

import "testing"

// TestSetCommand espelha TestRemoverSet de test_scripts.py:
// remove linhas SET <parametro> = ...;/TO ...; de um dump.

func TestSetCommand(t *testing.T) {
	t.Run("remove_parametro_padrao", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertNotContains(t, output, "SET transaction_timeout = 0;")
	})

	t.Run("case_insensitive_keyword_e_parametro", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertNotContains(t, output, "TRANSACTION_TIMEOUT = 5000")
	})

	t.Run("remove_forma_com_aspas_e_to", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertNotContains(t, output, `"transaction_timeout" TO 3000`)
	})

	t.Run("nao_remove_parametro_com_prefixo_igual", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertContains(t, output, "SET transaction_timeout_like = 1;")
	})

	t.Run("mantem_outros_parametros", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertContains(t, output, "SET statement_timeout = 0;")
		assertContains(t, output, "SET search_path = public;")
	})

	t.Run("mantem_estrutura_e_dados", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertContains(t, output, "CREATE TABLE public.users")
		assertContains(t, output, "COPY public.users (id) FROM stdin;")
		assertContains(t, output, "1\n\\.")
	})

	t.Run("multiplos_parametros", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout,statement_timeout", false, false)
		assertNotContains(t, output, "SET statement_timeout = 0;")
		assertContains(t, output, "SET search_path = public;")
	})

	t.Run("relatorio_conta_linhas_removidas", func(t *testing.T) {
		_, stdout := runSetT(t, dumpSetSample, "transaction_timeout", false, false)
		assertContains(t, stdout, "3 removidas")
	})

	t.Run("gz_in_gz_out", func(t *testing.T) {
		output, _ := runSetT(t, dumpSetSample, "transaction_timeout", true, true)
		assertNotContains(t, output, "transaction_timeout = 0")
		assertContains(t, output, "CREATE TABLE public.users")
	})
}
