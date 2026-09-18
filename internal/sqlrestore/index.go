package sqlrestore

import (
	"fmt"
	"os"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
)

// SegmentKind distingue um trecho de SQL opaco de um bloco de dados COPY.
type SegmentKind int

const (
	SegmentSQL SegmentKind = iota
	SegmentCopy
)

// Segment é um trecho do dump já indexado por offset de byte (no arquivo
// plano, não comprimido), pronto para ser aplicado: sequencialmente
// (SegmentSQL) ou em lote paralelo (SegmentCopy, quando adjacente a outros
// SegmentCopy).
type Segment struct {
	Kind SegmentKind

	// Start/End delimitam em bytes, para SegmentCopy, apenas as linhas de
	// DADOS entre o cabeçalho "COPY ... FROM stdin;" e o terminador "\.";
	// para SegmentSQL, o trecho de SQL a executar tal como está no arquivo.
	Start, End int64

	// Header é a linha "COPY ... FROM stdin;" original (só SegmentCopy),
	// reaproveitada tal qual como comando para pgconn.CopyFrom.
	Header string

	// Table é o nome da tabela do bloco (só SegmentCopy), usado no relatório
	// de progresso/dry-run.
	Table sqldump.QualifiedName
}

// BuildIndex varre plainPath (um arquivo .sql já descomprimido, com acesso
// aleatório) linha a linha e devolve a lista ordenada de segmentos que
// Apply percorre depois.
func BuildIndex(plainPath string) ([]Segment, error) {
	f, err := os.Open(plainPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := sqldump.NewLineScanner(f)

	var segments []Segment
	var offset int64
	var sqlStart int64
	inCopy := false
	var copyHeader string
	var copyStart int64
	var copyTable sqldump.QualifiedName

	flushSQL := func(end int64) {
		if end > sqlStart {
			segments = append(segments, Segment{Kind: SegmentSQL, Start: sqlStart, End: end})
		}
	}

	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}
		lineLen := int64(len(line))

		if inCopy {
			if strings.TrimSpace(line) == `\.` {
				segments = append(segments, Segment{
					Kind:   SegmentCopy,
					Start:  copyStart,
					End:    offset,
					Header: copyHeader,
					Table:  copyTable,
				})
				inCopy = false
				sqlStart = offset + lineLen
			}
			offset += lineLen
			continue
		}

		if strings.HasPrefix(line, "COPY ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				flushSQL(offset)
				copyHeader = strings.TrimRight(line, "\r\n")
				copyTable = sqldump.ParseQualifiedName(parts[1])
				copyStart = offset + lineLen
				inCopy = true
				offset += lineLen
				continue
			}
		}

		// Meta-comandos do psql (ex.: "\restrict <token>"/"\unrestrict <token>",
		// que o pg_dump recente emite por segurança) não são SQL e o servidor
		// os rejeita com erro de sintaxe — só o psql os entende. "\." já foi
		// tratado acima (só é válido dentro de um bloco COPY); qualquer outra
		// linha começando com "\" fora de um COPY é pulada, nunca enviada.
		if strings.HasPrefix(strings.TrimSpace(line), `\`) {
			flushSQL(offset)
			offset += lineLen
			sqlStart = offset
			continue
		}

		offset += lineLen
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inCopy {
		return nil, fmt.Errorf("bloco COPY %s sem terminador '\\.'", copyTable.String())
	}
	flushSQL(offset)

	return segments, nil
}

// TableStat resume o tamanho (em bytes de dados) de uma tabela indexada,
// usado no relatório de --dry-run.
type TableStat struct {
	Table sqldump.QualifiedName
	Bytes int64
}

// Summarize agrega os segmentos COPY por tabela, na ordem em que aparecem
// pela primeira vez no dump.
func Summarize(segments []Segment) []TableStat {
	var order []string
	stats := map[string]*TableStat{}
	for _, seg := range segments {
		if seg.Kind != SegmentCopy {
			continue
		}
		key := seg.Table.String()
		st, ok := stats[key]
		if !ok {
			st = &TableStat{Table: seg.Table}
			stats[key] = st
			order = append(order, key)
		}
		st.Bytes += seg.End - seg.Start
	}
	out := make([]TableStat, 0, len(order))
	for _, k := range order {
		out = append(out, *stats[k])
	}
	return out
}
