package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestRestaurarIntegration roda `restaurar` de ponta a ponta contra um
// PostgreSQL real: cria um banco descartável, restaura o dump nele via
// runRestaurar (a mesma função usada pelo comando), confere os dados linha
// a linha e testa a confirmação de segurança sobre um banco não-vazio.
//
// Requer a variável de ambiente SQLCLEANER_TEST_DSN, uma URL de conexão
// (ex.: postgres://postgres:postgres@localhost:5432/postgres) apontando
// para um servidor de testes onde é seguro criar/derrubar bancos. Sem essa
// variável o teste é pulado, então `go test ./...` continua funcionando sem
// um Postgres disponível.
//
// Para rodar localmente:
//
//	docker run -d --rm -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
//	SQLCLEANER_TEST_DSN=postgres://postgres:postgres@localhost:5432/postgres go test ./cmd/sqlcleaner/... -run Integration -v
func TestRestaurarIntegration(t *testing.T) {
	baseDSN := os.Getenv("SQLCLEANER_TEST_DSN")
	if baseDSN == "" {
		t.Skip("SQLCLEANER_TEST_DSN não definida; pulando teste de integração com Postgres real")
	}

	u, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("SQLCLEANER_TEST_DSN inválida (precisa ser uma URL postgres://...): %v", err)
	}

	dbName := fmt.Sprintf("sqlcleaner_restaurar_test_%d", time.Now().UnixNano())
	u.Path = "/" + dbName
	targetDSN := u.String()

	ctx := context.Background()
	t.Cleanup(func() {
		adminCfg, err := pgconn.ParseConfig(baseDSN)
		if err != nil {
			return
		}
		conn, err := pgconn.ConnectConfig(ctx, adminCfg)
		if err != nil {
			t.Logf("aviso: não consegui reconectar para limpar %s: %v", dbName, err)
			return
		}
		defer conn.Close(ctx)
		if _, err := conn.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName).ReadAll(); err != nil {
			t.Logf("aviso: falha ao remover banco de teste %s: %v", dbName, err)
		}
	})

	in := writeTempFile(t, dumpSample, false)

	// 1) restauração real, com criação do banco e paralelismo na fase de dados.
	var out strings.Builder
	err = runRestaurar(&out, strings.NewReader(""), in, restaurarOptions{
		dsn:      targetDSN,
		createDB: true,
		yes:      true,
		jobs:     4,
	})
	if err != nil {
		t.Fatalf("runRestaurar: %v", err)
	}
	assertContains(t, out.String(), "Concluído")

	if got := countRows(t, targetDSN, "users"); got != 2 {
		t.Errorf("esperava 2 linhas em users, veio %d", got)
	}
	if got := countRows(t, targetDSN, "logs"); got != 2 {
		t.Errorf("esperava 2 linhas em logs, veio %d", got)
	}

	// 2) rodando de novo sobre o banco não-vazio, sem --yes: deve pedir
	// confirmação e abortar quando a resposta for "n".
	out.Reset()
	err = runRestaurar(&out, strings.NewReader("n\n"), in, restaurarOptions{dsn: targetDSN})
	if err != nil {
		t.Fatalf("runRestaurar (confirmação): %v", err)
	}
	assertContains(t, out.String(), "cancelada")
	if got := countRows(t, targetDSN, "users"); got != 2 {
		t.Errorf("restauração cancelada não deveria ter alterado os dados; users=%d", got)
	}
}

func countRows(t *testing.T, dsn, table string) int {
	t.Helper()
	cfg, err := pgconn.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("dsn inválida: %v", err)
	}
	ctx := context.Background()
	conn, err := pgconn.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("conectando para verificar %s: %v", table, err)
	}
	defer conn.Close(ctx)

	result := conn.ExecParams(ctx, "SELECT count(*) FROM "+table, nil, nil, nil, nil)
	var count int
	for result.NextRow() {
		vals := result.Values()
		count, err = strconv.Atoi(string(vals[0]))
		if err != nil {
			t.Fatalf("resultado inesperado de count(*): %v", err)
		}
	}
	if _, err := result.Close(); err != nil {
		t.Fatalf("consultando %s: %v", table, err)
	}
	return count
}
