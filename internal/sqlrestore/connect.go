package sqlrestore

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// ParseConfig resolve a configuração de conexão a partir de dsn; se dsn
// estiver vazio, pgconn lê as variáveis PG* (PGHOST, PGPORT, PGUSER,
// PGPASSWORD, PGDATABASE, ...) do ambiente, igual ao psql.
func ParseConfig(dsn string) (*pgconn.Config, error) {
	cfg, err := pgconn.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn inválida: %w", err)
	}
	return cfg, nil
}

// maintenanceConn conecta ao mesmo servidor de cfg, mas na base de
// manutenção "postgres" (usada para checar/criar o banco de destino, que
// não pode ser feito enquanto conectado nele mesmo).
func maintenanceConn(ctx context.Context, cfg *pgconn.Config) (*pgconn.PgConn, error) {
	maint := cfg.Copy()
	maint.Database = "postgres"
	return pgconn.ConnectConfig(ctx, maint)
}

// DatabaseExists verifica se cfg.Database já existe no servidor.
func DatabaseExists(ctx context.Context, cfg *pgconn.Config) (bool, error) {
	conn, err := maintenanceConn(ctx, cfg)
	if err != nil {
		return false, err
	}
	defer conn.Close(ctx)

	result := conn.ExecParams(ctx, "SELECT 1 FROM pg_database WHERE datname = $1",
		[][]byte{[]byte(cfg.Database)}, nil, nil, nil)
	found := false
	for result.NextRow() {
		found = true
	}
	if _, err := result.Close(); err != nil {
		return false, err
	}
	return found, nil
}

// CreateDatabase cria cfg.Database se ele ainda não existir.
func CreateDatabase(ctx context.Context, cfg *pgconn.Config) error {
	exists, err := DatabaseExists(ctx, cfg)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	conn, err := maintenanceConn(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// CREATE DATABASE não aceita o nome como parâmetro; citamos como
	// identificador (cfg.Database vem de pgconn.ParseConfig/flags do
	// operador, não de entrada não confiável) para suportar maiúsculas e
	// caracteres especiais.
	_, err = conn.Exec(ctx, "CREATE DATABASE "+quoteIdentifier(cfg.Database)).ReadAll()
	return err
}

// HasUserTables verifica se o banco de destino (cfg.Database) já tem
// alguma tabela fora dos schemas internos do Postgres — usado para decidir
// se pede confirmação antes de restaurar por cima de dados existentes.
func HasUserTables(ctx context.Context, cfg *pgconn.Config) (bool, error) {
	conn, err := pgconn.ConnectConfig(ctx, cfg)
	if err != nil {
		return false, err
	}
	defer conn.Close(ctx)

	result := conn.ExecParams(ctx, `
		SELECT 1 FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		LIMIT 1`, nil, nil, nil, nil)
	found := false
	for result.NextRow() {
		found = true
	}
	if _, err := result.Close(); err != nil {
		return false, err
	}
	return found, nil
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
