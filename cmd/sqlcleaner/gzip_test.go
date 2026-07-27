package main

import "testing"

// TestGzip espelha TestGzip de test_scripts.py: suporte a .sql.gz na
// entrada e/ou saída.
func TestGzip(t *testing.T) {
	t.Run("remover_gz_in_plain_out", func(t *testing.T) {
		output := runRemoverT(t, dumpSample, "logs", true, false)
		assertNotContains(t, output, "CREATE TABLE public.logs")
		assertContains(t, output, "CREATE TABLE public.users")
	})
	t.Run("remover_plain_in_gz_out", func(t *testing.T) {
		output := runRemoverT(t, dumpSample, "logs", false, true)
		assertNotContains(t, output, "CREATE TABLE public.logs")
		assertContains(t, output, "CREATE TABLE public.users")
	})
	t.Run("remover_gz_in_gz_out", func(t *testing.T) {
		output := runRemoverT(t, dumpSample, "logs", true, true)
		assertNotContains(t, output, "startup")
		assertContains(t, output, "Alice")
	})
	t.Run("esvaziar_gz_in_plain_out", func(t *testing.T) {
		output := runEsvaziarT(t, dumpSample, "logs", true, false)
		assertNotContains(t, output, "startup")
		assertContains(t, output, "CREATE TABLE public.logs")
	})
	t.Run("esvaziar_plain_in_gz_out", func(t *testing.T) {
		output := runEsvaziarT(t, dumpSample, "logs", false, true)
		assertNotContains(t, output, "startup")
		assertContains(t, output, "CREATE TABLE public.logs")
	})
	t.Run("esvaziar_gz_in_gz_out", func(t *testing.T) {
		output := runEsvaziarT(t, dumpSample, "logs", true, true)
		assertNotContains(t, output, "startup")
		assertContains(t, output, "Alice")
	})
}
