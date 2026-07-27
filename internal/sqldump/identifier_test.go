package sqldump

import "testing"

func TestQualifiedNameString(t *testing.T) {
	cases := []struct {
		name string
		q    QualifiedName
		want string
	}{
		{"sem_schema", QualifiedName{Table: "logs"}, "logs"},
		{"com_schema", QualifiedName{Schema: "public", HasSchema: true, Table: "logs"}, "public.logs"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.q.String(); got != c.want {
				t.Errorf("String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestReadIdentifier(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		pos     int
		want    string
		wantPos int
	}{
		{"simples", "tabela", 0, "tabela", 6},
		{"ate_ponto", "schema.tabela", 0, "schema", 6},
		{"apos_ponto", "schema.tabela", 7, "tabela", 13},
		{"citado", `"Tabela Com Espaço"`, 0, "Tabela Com Espaço", 19},
		{"ate_espaco", "tabela extra", 0, "tabela", 6},
		{"ate_tab", "tabela\textra", 0, "tabela", 6},
		{"ate_parenteses", "tabela(col)", 0, "tabela", 6},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, gotPos := ReadIdentifier([]rune(c.s), c.pos)
			if got != c.want || gotPos != c.wantPos {
				t.Errorf("ReadIdentifier(%q, %d) = (%q, %d), want (%q, %d)", c.s, c.pos, got, gotPos, c.want, c.wantPos)
			}
		})
	}
}

func TestParseQualifiedName(t *testing.T) {
	cases := []struct {
		name  string
		token string
		want  QualifiedName
	}{
		{"sem_schema", "tabela", QualifiedName{Table: "tabela"}},
		{"com_schema", "schema.tabela", QualifiedName{Schema: "schema", HasSchema: true, Table: "tabela"}},
		{"tabela_citada", `"Tabela"`, QualifiedName{Table: "Tabela"}},
		{"schema_citado_tabela_citada", `"Schema"."Tabela"`, QualifiedName{Schema: "Schema", HasSchema: true, Table: "Tabela"}},
		{"schema_simples_tabela_citada", `schema."Tabela"`, QualifiedName{Schema: "schema", HasSchema: true, Table: "Tabela"}},
		{"ponto_literal_citado", `"DW.ticketsDW"`, QualifiedName{Table: "DW.ticketsDW"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseQualifiedName(c.token); got != c.want {
				t.Errorf("ParseQualifiedName(%q) = %+v, want %+v", c.token, got, c.want)
			}
		})
	}
}

func TestFindMatchingTarget(t *testing.T) {
	targets := []QualifiedName{
		{Table: "logs"},
		{Schema: "dw", HasSchema: true, Table: "ticketsdw"},
	}

	cases := []struct {
		name      string
		schema    string
		hasSchema bool
		table     string
		wantFound bool
		wantTgt   QualifiedName
	}{
		{"sem_schema_casa_alvo_sem_schema", "", false, "logs", true, targets[0]},
		{"schema_public_casa_alvo_sem_schema", "public", true, "logs", true, targets[0]},
		{"outro_schema_nao_casa_alvo_sem_schema", "outro", true, "logs", false, QualifiedName{}},
		{"schema_especifico_casa", "dw", true, "ticketsdw", true, targets[1]},
		{"schema_especifico_nao_casa_schema_diferente", "outro", true, "ticketsdw", false, QualifiedName{}},
		{"tabela_inexistente", "", false, "inexistente", false, QualifiedName{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, found := FindMatchingTarget(c.schema, c.hasSchema, c.table, targets)
			if found != c.wantFound || got != c.wantTgt {
				t.Errorf("FindMatchingTarget(%q, %v, %q) = (%+v, %v), want (%+v, %v)",
					c.schema, c.hasSchema, c.table, got, found, c.wantTgt, c.wantFound)
			}
		})
	}
}
