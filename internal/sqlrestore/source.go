package sqlrestore

import (
	"io"
	"os"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
)

// PreparePlainFile garante um arquivo .sql plano e com acesso aleatório a
// partir de path (que pode ser .sql ou .sql.gz). Para entrada .sql, devolve
// o próprio path. Para .sql.gz, descomprime uma vez para um arquivo
// temporário: a indexação e a fase de dados paralela precisam de seek por
// offset de byte, que gzip não permite diretamente.
//
// cleanup deve sempre ser chamado pelo chamador (é no-op quando path já era
// um arquivo plano).
func PreparePlainFile(path string) (plainPath string, cleanup func(), err error) {
	if !strings.HasSuffix(path, ".gz") {
		return path, func() {}, nil
	}

	reader, err := sqldump.OpenReader(path)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	tmp, err := os.CreateTemp("", "sqlcleaner-restaurar-*.sql")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.Remove(tmp.Name()) }

	if _, err := io.Copy(tmp, reader); err != nil {
		tmp.Close()
		cleanup()
		return "", nil, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", nil, err
	}

	return tmp.Name(), cleanup, nil
}
