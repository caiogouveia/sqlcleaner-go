package sqldump

import "testing"

func TestParseTruncateTargets(t *testing.T) {
	cases := []struct {
		name         string
		tablesArg    string
		defaultLimit *int
		wantErr      bool
		want         []TruncateTarget
	}{
		{
			"apenas_nome_usa_default",
			"logs",
			intPtr(100),
			false,
			[]TruncateTarget{{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 100, Order: "ASC"}}},
		},
		{
			"nome_com_limite",
			"logs:5",
			nil,
			false,
			[]TruncateTarget{{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 5, Order: "ASC"}}},
		},
		{
			"nome_com_ordem",
			"logs:desc",
			intPtr(10),
			false,
			[]TruncateTarget{{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 10, Order: "DESC"}}},
		},
		{
			"nome_com_limite_e_ordem",
			"logs:5:desc",
			nil,
			false,
			[]TruncateTarget{{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 5, Order: "DESC"}}},
		},
		{
			"multiplas_tabelas",
			"logs:5,dw.ticketsdw:10:desc",
			nil,
			false,
			[]TruncateTarget{
				{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 5, Order: "ASC"}},
				{Key: QualifiedName{Schema: "dw", HasSchema: true, Table: "ticketsdw"}, Config: TruncateConfig{Limit: 10, Order: "DESC"}},
			},
		},
		{
			"limite_invalido",
			"logs:abc",
			nil,
			true,
			nil,
		},
		{
			"ordem_invalida",
			"logs:5:LADO",
			nil,
			true,
			nil,
		},
		{
			"formato_invalido_demais_partes",
			"logs:5:desc:extra",
			nil,
			true,
			nil,
		},
		{
			"sem_limite_e_sem_default",
			"logs",
			nil,
			true,
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseTruncateTargets(c.tablesArg, c.defaultLimit)
			if c.wantErr {
				if err == nil {
					t.Fatalf("esperava erro, mas não houve. got=%+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("len(got) = %d, want %d (got=%+v)", len(got), len(c.want), got)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("got[%d] = %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestFindConfig(t *testing.T) {
	targets := []TruncateTarget{
		{Key: QualifiedName{Table: "logs"}, Config: TruncateConfig{Limit: 100, Order: "ASC"}},
		{Key: QualifiedName{Schema: "dw", HasSchema: true, Table: "ticketsdw"}, Config: TruncateConfig{Limit: 10, Order: "DESC"}},
	}

	t.Run("casa_sem_schema_com_public", func(t *testing.T) {
		key, cfg, found := FindConfig("public", true, "logs", targets)
		if !found || key != targets[0].Key || cfg != targets[0].Config {
			t.Errorf("got (%+v, %+v, %v)", key, cfg, found)
		}
	})

	t.Run("casa_schema_especifico", func(t *testing.T) {
		key, cfg, found := FindConfig("dw", true, "ticketsdw", targets)
		if !found || key != targets[1].Key || cfg != targets[1].Config {
			t.Errorf("got (%+v, %+v, %v)", key, cfg, found)
		}
	})

	t.Run("nao_encontrado", func(t *testing.T) {
		_, _, found := FindConfig("outro", true, "inexistente", targets)
		if found {
			t.Errorf("esperava não encontrar, mas encontrou")
		}
	})
}

func intPtr(n int) *int { return &n }
