package main

import "testing"

// TestSchemaVariants espelha TestSchemaVariants de test_scripts.py: ambos os
// comandos devem reconhecer tabelas com e sem prefixo public.
func TestSchemaVariants(t *testing.T) {
	t.Run("remover_unqualified_name", func(t *testing.T) {
		output := runRemoverT(t, dumpSample, "logs", false, false)
		assertNotContains(t, output, "CREATE TABLE public.logs")
	})
	t.Run("esvaziar_unqualified_name", func(t *testing.T) {
		output := runEsvaziarT(t, dumpSample, "logs", false, false)
		assertNotContains(t, output, "startup")
		assertContains(t, output, "CREATE TABLE public.logs")
	})
}

// TestCaseSensitiveEQuoted espelha TestCaseSensitiveEQuoted de
// test_scripts.py: tabelas com maiúsculas/mistas (aspas duplas no dump) e
// schemas não-públicos.
func TestCaseSensitiveEQuoted(t *testing.T) {
	t.Run("esvaziar_nome_maiusculo_sem_schema", func(t *testing.T) {
		output := runEsvaziarT(t, dumpCaseSensitive, "ticketsDW", false, false)
		assertContains(t, output, `CREATE TABLE public."ticketsDW"`)
		assertNotContains(t, output, "1\n2\n\\.")
		assertContains(t, output, "10\n\\.")
	})

	t.Run("esvaziar_nome_maiusculo_com_schema", func(t *testing.T) {
		output := runEsvaziarT(t, dumpCaseSensitive, "DW.ticketsDW", false, false)
		assertContains(t, output, `CREATE TABLE "DW"."ticketsDW"`)
		assertNotContains(t, output, "10\n\\.")
		assertContains(t, output, "1\n2\n\\.")
	})

	t.Run("remover_nome_maiusculo_sem_schema", func(t *testing.T) {
		output := runRemoverT(t, dumpCaseSensitive, "ticketsDW", false, false)
		assertNotContains(t, output, `CREATE TABLE public."ticketsDW"`)
		assertNotContains(t, output, `CREATE SEQUENCE public."ticketsDW_id_seq"`)
		assertNotContains(t, output, "idx_ticketsdw_id")
		assertNotContains(t, output, "1\n2\n\\.")
		assertContains(t, output, `CREATE TABLE "DW"."ticketsDW"`)
		assertContains(t, output, "10\n\\.")
	})

	t.Run("truncar_nome_maiusculo_sem_schema", func(t *testing.T) {
		output := runTruncarT(t, dumpCaseSensitive, "ticketsDW:1", nil, false, false)
		assertContains(t, output, "1\n\\.")
		assertNotContains(t, output, "2\n\\.")
		assertContains(t, output, "10\n\\.")
	})

	t.Run("nome_com_ponto_literal_sem_aspas_nao_bate", func(t *testing.T) {
		// sem aspas, "DW.ticketsDW" é interpretado como schema.tabela (não existe) — não deve casar
		output := runEsvaziarT(t, dumpLiteralDot, "DW.ticketsDW", false, false)
		assertContains(t, output, "1\n2\n\\.")
	})

	t.Run("nome_com_ponto_literal_com_aspas_bate", func(t *testing.T) {
		// com aspas, o ponto faz parte do nome da tabela (schema public)
		output := runEsvaziarT(t, dumpLiteralDot, `"DW.ticketsDW"`, false, false)
		assertNotContains(t, output, "1\n2\n\\.")
		assertContains(t, output, `CREATE TABLE public."DW.ticketsDW"`)
		assertContains(t, output, "9\n\\.")
	})
}
