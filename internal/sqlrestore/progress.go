package sqlrestore

import (
	"fmt"
	"io"
	"sync"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
)

// PrintProgress é uma implementação de Progress que imprime, em w, uma
// linha por tabela concluída (ou com erro), no mesmo estilo de resumo dos
// outros subcomandos.
type PrintProgress struct {
	w  io.Writer
	mu sync.Mutex
}

func NewPrintProgress(w io.Writer) *PrintProgress {
	return &PrintProgress{w: w}
}

func (p *PrintProgress) TableStarted(table string) {}

func (p *PrintProgress) TableDone(table string, bytes int64, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		fmt.Fprintf(p.w, i18n.T("restaurar.table.error"), table, err)
		return
	}
	fmt.Fprintf(p.w, i18n.T("restaurar.table.done"), table, sqldump.FormatSize(float64(bytes)))
}
