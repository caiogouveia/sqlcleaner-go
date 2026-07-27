package main

import (
	"strings"
	"testing"
)

// TestEsvaziar espelha TestEsvaziarTabelas de test_scripts.py:
// apaga dados mas mantém estrutura.

func TestEsvaziar(t *testing.T) {
	output := runEsvaziarT(t, dumpSample, "logs", false, false)

	t.Run("keeps_create_table", func(t *testing.T) {
		assertContains(t, output, "CREATE TABLE public.logs")
	})
	t.Run("keeps_sequence", func(t *testing.T) {
		assertContains(t, output, "CREATE SEQUENCE public.logs_id_seq")
	})
	t.Run("keeps_index", func(t *testing.T) {
		assertContains(t, output, "idx_logs_id")
	})
	t.Run("keeps_copy_header", func(t *testing.T) {
		assertContains(t, output, "COPY public.logs (id, message) FROM stdin;")
	})
	t.Run("keeps_copy_terminator", func(t *testing.T) {
		count := strings.Count(output, `\.`)
		if count != 2 {
			t.Errorf("esperava 2 ocorrências de \\., encontrou %d", count)
		}
	})
	t.Run("removes_data_rows", func(t *testing.T) {
		assertNotContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})
	t.Run("keeps_other_table_data", func(t *testing.T) {
		assertContains(t, output, "Alice")
		assertContains(t, output, "Bob")
	})
	t.Run("keeps_other_table_structure", func(t *testing.T) {
		assertContains(t, output, "CREATE TABLE public.users")
	})
}
