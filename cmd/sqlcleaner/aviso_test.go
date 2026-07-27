package main

import "testing"

// TestAvisoSemCorrespondencia espelha TestAvisoSemCorrespondencia de
// test_scripts.py: aviso no relatório final quando um nome em -t não casa
// com nada no dump.
func TestAvisoSemCorrespondencia(t *testing.T) {
	t.Run("esvaziar_avisa_tabela_inexistente", func(t *testing.T) {
		stdout := runEsvaziarStdoutT(t, dumpSample, "tabela_inexistente")
		assertContains(t, stdout, "Aviso")
		assertContains(t, stdout, "tabela_inexistente")
	})

	t.Run("esvaziar_sem_aviso_quando_tudo_casa", func(t *testing.T) {
		stdout := runEsvaziarStdoutT(t, dumpSample, "logs")
		assertNotContains(t, stdout, "Aviso")
	})

	t.Run("remover_avisa_tabela_inexistente", func(t *testing.T) {
		stdout := runRemoverStdoutT(t, dumpSample, "tabela_inexistente")
		assertContains(t, stdout, "Aviso")
		assertContains(t, stdout, "tabela_inexistente")
	})

	t.Run("remover_sem_aviso_quando_tudo_casa", func(t *testing.T) {
		stdout := runRemoverStdoutT(t, dumpSample, "logs")
		assertNotContains(t, stdout, "Aviso")
	})

	t.Run("truncar_avisa_tabela_inexistente", func(t *testing.T) {
		stdout := runTruncarStdoutT(t, dumpSample, "tabela_inexistente:5", nil)
		assertContains(t, stdout, "Aviso")
		assertContains(t, stdout, "tabela_inexistente")
	})

	t.Run("truncar_sem_aviso_quando_tudo_casa", func(t *testing.T) {
		stdout := runTruncarStdoutT(t, dumpSample, "logs:1", nil)
		assertNotContains(t, stdout, "Aviso")
	})

	t.Run("truncar_avisa_ponto_literal_sem_aspas", func(t *testing.T) {
		// nome com ponto literal, sem aspas internas, é interpretado como
		// schema.tabela e não casa com nada no dump
		stdout := runTruncarStdoutT(t, dumpLiteralDot, "DW.ticketsDW:5", nil)
		assertContains(t, stdout, "Aviso")
		assertContains(t, stdout, "DW.ticketsDW")
	})
}
