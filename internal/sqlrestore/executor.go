package sqlrestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
)

// Progress reporta o andamento da restauração.
type Progress interface {
	TableStarted(table string)
	TableDone(table string, bytes int64, err error)
}

// Options controla a execução de Apply.
type Options struct {
	// Jobs é o paralelismo da fase de dados (COPY); valores <= 1 rodam
	// sequencialmente.
	Jobs int
}

// Apply aplica os segmentos de plainPath (arquivo .sql já plano, gerado por
// PreparePlainFile) no banco descrito por cfg, na ordem em que aparecem no
// dump: trechos SQL rodam sequencialmente (servem de barreira), e corridas
// de blocos COPY adjacentes rodam em paralelo entre si, respeitando o
// paralelismo de Options.Jobs.
func Apply(ctx context.Context, cfg *pgconn.Config, plainPath string, segments []Segment, opts Options, prog Progress) error {
	f, err := os.Open(plainPath)
	if err != nil {
		return err
	}
	defer f.Close()

	conn, err := pgconn.ConnectConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("conectando ao destino: %w", err)
	}
	defer conn.Close(ctx)

	jobs := opts.Jobs
	if jobs < 1 {
		jobs = 1
	}

	i := 0
	for i < len(segments) {
		seg := segments[i]
		if seg.Kind == SegmentSQL {
			if err := execSQLSegment(ctx, conn, f, seg); err != nil {
				return err
			}
			i++
			continue
		}

		j := i
		for j < len(segments) && segments[j].Kind == SegmentCopy {
			j++
		}
		if err := runCopyBatch(ctx, cfg, plainPath, segments[i:j], jobs, prog); err != nil {
			return err
		}
		i = j
	}

	return nil
}

func execSQLSegment(ctx context.Context, conn *pgconn.PgConn, f *os.File, seg Segment) error {
	buf := make([]byte, seg.End-seg.Start)
	if _, err := f.ReadAt(buf, seg.Start); err != nil {
		return err
	}
	sql := string(buf)
	if isBlank(sql) {
		return nil
	}
	_, err := conn.Exec(ctx, sql).ReadAll()
	return err
}

// isBlank evita um round-trip ao servidor para trechos só com linhas em
// branco (ex.: entre um bloco COPY e o próximo).
func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func runCopyBatch(ctx context.Context, cfg *pgconn.Config, plainPath string, batch []Segment, jobs int, prog Progress) error {
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	errs := make([]error, len(batch))

	for idx, seg := range batch {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, seg Segment) {
			defer wg.Done()
			defer func() { <-sem }()
			errs[idx] = copySegment(ctx, cfg, plainPath, seg, prog)
		}(idx, seg)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func copySegment(ctx context.Context, cfg *pgconn.Config, plainPath string, seg Segment, prog Progress) error {
	tableName := seg.Table.String()
	if prog != nil {
		prog.TableStarted(tableName)
	}

	conn, err := pgconn.ConnectConfig(ctx, cfg)
	if err != nil {
		if prog != nil {
			prog.TableDone(tableName, 0, err)
		}
		return err
	}
	defer conn.Close(ctx)

	f, err := os.Open(plainPath)
	if err != nil {
		if prog != nil {
			prog.TableDone(tableName, 0, err)
		}
		return err
	}
	defer f.Close()

	r := io.NewSectionReader(f, seg.Start, seg.End-seg.Start)
	_, err = conn.CopyFrom(ctx, r, seg.Header)
	if prog != nil {
		prog.TableDone(tableName, seg.End-seg.Start, err)
	}
	return err
}
