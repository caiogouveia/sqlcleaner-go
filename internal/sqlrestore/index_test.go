package sqlrestore

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleDump = "--\n" +
	"-- PostgreSQL database dump\n" +
	"--\n" +
	"\n" +
	"SET statement_timeout = 0;\n" +
	"\n" +
	"CREATE TABLE public.users (\n" +
	"    id integer NOT NULL,\n" +
	"    name text\n" +
	");\n" +
	"\n" +
	"CREATE TABLE public.logs (\n" +
	"    id integer NOT NULL,\n" +
	"    message text\n" +
	");\n" +
	"\n" +
	"COPY public.users (id, name) FROM stdin;\n" +
	"1\tAlice\n" +
	"2\tBob\n" +
	"\\.\n" +
	"\n" +
	"COPY public.logs (id, message) FROM stdin;\n" +
	"1\tstartup\n" +
	"2\tshutdown\n" +
	"\\.\n" +
	"\n" +
	"ALTER TABLE ONLY public.users\n" +
	"    ADD CONSTRAINT users_pkey PRIMARY KEY (id);\n" +
	"\n" +
	"ALTER TABLE ONLY public.logs\n" +
	"    ADD CONSTRAINT logs_pkey PRIMARY KEY (id);\n"

func writeSample(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dump.sql")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBuildIndexSegments(t *testing.T) {
	path := writeSample(t, sampleDump)

	segments, err := BuildIndex(path)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	var kinds []SegmentKind
	for _, s := range segments {
		kinds = append(kinds, s.Kind)
	}
	// A linha em branco entre os dois blocos COPY (padrão do pg_dump) vira
	// seu próprio segmento SQL "em branco" — o executor o pula por ser vazio,
	// mas a indexação preserva a posição.
	want := []SegmentKind{SegmentSQL, SegmentCopy, SegmentSQL, SegmentCopy, SegmentSQL}
	if len(kinds) != len(want) {
		t.Fatalf("esperava %d segmentos, veio %d: %+v", len(want), len(kinds), segments)
	}
	for i, k := range want {
		if kinds[i] != k {
			t.Errorf("segmento %d: esperava kind %v, veio %v", i, k, kinds[i])
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	users := segments[1]
	if got := string(raw[users.Start:users.End]); got != "1\tAlice\n2\tBob\n" {
		t.Errorf("dados de users incorretos: %q", got)
	}
	if users.Table.String() != "public.users" {
		t.Errorf("tabela incorreta: %q", users.Table.String())
	}
	if users.Header != "COPY public.users (id, name) FROM stdin;" {
		t.Errorf("header incorreto: %q", users.Header)
	}

	logs := segments[3]
	if got := string(raw[logs.Start:logs.End]); got != "1\tstartup\n2\tshutdown\n" {
		t.Errorf("dados de logs incorretos: %q", got)
	}

	lastSQL := segments[4]
	if got := string(raw[lastSQL.Start:lastSQL.End]); got == "" {
		t.Errorf("trecho SQL final não deveria ser vazio")
	}
}

func TestSummarize(t *testing.T) {
	path := writeSample(t, sampleDump)
	segments, err := BuildIndex(path)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	stats := Summarize(segments)
	if len(stats) != 2 {
		t.Fatalf("esperava 2 tabelas, veio %d: %+v", len(stats), stats)
	}
	if stats[0].Table.String() != "public.users" || stats[0].Bytes != int64(len("1\tAlice\n2\tBob\n")) {
		t.Errorf("stat[0] incorreto: %+v", stats[0])
	}
	if stats[1].Table.String() != "public.logs" || stats[1].Bytes != int64(len("1\tstartup\n2\tshutdown\n")) {
		t.Errorf("stat[1] incorreto: %+v", stats[1])
	}
}

func TestBuildIndexUnterminatedCopy(t *testing.T) {
	path := writeSample(t, "COPY public.users (id) FROM stdin;\n1\n")
	if _, err := BuildIndex(path); err == nil {
		t.Fatal("esperava erro para bloco COPY sem terminador")
	}
}

func TestBuildIndexEmptyCopyBlock(t *testing.T) {
	content := "CREATE TABLE public.empty_table (id integer);\n\n" +
		"COPY public.empty_table (id) FROM stdin;\n\\.\n"
	path := writeSample(t, content)

	segments, err := BuildIndex(path)
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	var found bool
	for _, s := range segments {
		if s.Kind == SegmentCopy {
			found = true
			if s.End != s.Start {
				t.Errorf("bloco COPY vazio deveria ter Start == End, veio %d/%d", s.Start, s.End)
			}
		}
	}
	if !found {
		t.Fatal("esperava encontrar um segmento COPY")
	}
}
