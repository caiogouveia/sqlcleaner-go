package main

import "testing"

// TestRemover espelha TestLimparSQL de test_scripts.py:
// remove tabela inteira (dados + estrutura).

func TestRemover(t *testing.T) {
	output := runRemoverT(t, dumpSample, "logs", false, false)

	t.Run("remove_create_table", func(t *testing.T) {
		assertNotContains(t, output, "CREATE TABLE public.logs")
	})
	t.Run("remove_sequence", func(t *testing.T) {
		assertNotContains(t, output, "CREATE SEQUENCE public.logs_id_seq")
	})
	t.Run("remove_alter_table", func(t *testing.T) {
		assertNotContains(t, output, "nextval('public.logs_id_seq")
	})
	t.Run("remove_index", func(t *testing.T) {
		assertNotContains(t, output, "idx_logs_id")
	})
	t.Run("remove_copy_data", func(t *testing.T) {
		assertNotContains(t, output, "startup")
		assertNotContains(t, output, "shutdown")
	})
	t.Run("keeps_other_table_structure", func(t *testing.T) {
		assertContains(t, output, "CREATE TABLE public.users")
		assertContains(t, output, "CREATE SEQUENCE public.users_id_seq")
		assertContains(t, output, "idx_users_name")
	})
	t.Run("keeps_other_table_data", func(t *testing.T) {
		assertContains(t, output, "Alice")
		assertContains(t, output, "Bob")
	})
}
