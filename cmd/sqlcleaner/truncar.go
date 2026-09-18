package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newTruncarCmd() *cobra.Command {
	var output string
	var tables string
	var limit int
	cmd := &cobra.Command{
		Use:   "truncar <entrada>",
		Short: i18n.T("truncar.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var defaultLimit *int
			if cmd.Flags().Changed("limit") {
				defaultLimit = &limit
			}
			return runTruncar(os.Stdout, args[0], output, tables, defaultLimit)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", i18n.T("flag.output"))
	cmd.Flags().StringVarP(&tables, "tables", "t", "", i18n.T("truncar.flag.tables"))
	cmd.Flags().IntVarP(&limit, "limit", "n", 0, i18n.T("truncar.flag.limit"))
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("tables")
	return cmd
}

func runTruncar(w io.Writer, inputPath, outputPath, tablesArg string, defaultLimit *int) error {
	targets, err := sqldump.ParseTruncateTargets(tablesArg, defaultLimit)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.parse"), err)
		return nil
	}

	var header strings.Builder
	fmt.Fprintln(&header, "--- SQL Table Truncator 1.0 ---")
	fmt.Fprintf(&header, i18n.T("label.entrada"), inputPath)
	fmt.Fprintf(&header, i18n.T("label.saida"), outputPath)
	for _, t := range targets {
		fmt.Fprintf(&header, i18n.T("truncar.target_header"), t.Key.String(), t.Config.Limit, t.Config.Order)
	}
	fmt.Fprintln(&header, "-------------------------------")

	return runTruncarCore(w, inputPath, outputPath, targets, header.String(), i18n.T("result.label.truncado"))
}

// runTruncarCore é o motor de streaming compartilhado por `truncar` e
// `esvaziar` (que é apenas um truncar com limite 0 para todas as tabelas).
func runTruncarCore(w io.Writer, inputPath, outputPath string, targets []sqldump.TruncateTarget, header, resultLabel string) error {
	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(w, i18n.T("err.file_not_found"), inputPath)
		return nil
	}

	fmt.Fprint(w, header)

	start := time.Now()

	reader, err := sqldump.OpenReader(inputPath)
	if err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	writer, err := sqldump.OpenWriter(outputPath)
	if err != nil {
		reader.Close()
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}

	bw := bufio.NewWriterSize(writer, 1<<20)
	scanner := sqldump.NewLineScanner(reader)

	inCopy := false
	currentLimit := 0
	currentOrder := "ASC"
	currentCount := 0
	var currentBuffer *ringBuffer
	skippedLines := 0
	totalLines := 0
	matched := map[sqldump.QualifiedName]bool{}

	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}
		totalLines++
		if totalLines%5000000 == 0 {
			fmt.Fprintf(w, i18n.T("progress"), totalLines/1000000)
		}

		trimmed := strings.TrimSpace(line)

		if inCopy {
			if trimmed == `\.` {
				inCopy = false
				if currentOrder == "DESC" {
					kept := currentBuffer.Lines()
					if d := currentCount - len(kept); d > 0 {
						skippedLines += d
					}
					for _, l := range kept {
						bw.WriteString(l)
					}
					currentBuffer = nil
				}
				bw.WriteString(line)
				continue
			}
			if currentOrder == "DESC" {
				currentCount++
				currentBuffer.Append(line)
			} else if currentCount < currentLimit {
				currentCount++
				bw.WriteString(line)
			} else {
				skippedLines++
			}
			continue
		}

		if strings.HasPrefix(line, "COPY ") {
			parts := strings.Fields(line)
			var cfg sqldump.TruncateConfig
			var found bool
			var key sqldump.QualifiedName
			if len(parts) >= 2 {
				qn := sqldump.ParseQualifiedName(parts[1])
				key, cfg, found = sqldump.FindConfig(qn.Schema, qn.HasSchema, qn.Table, targets)
				if found {
					matched[key] = true
				}
			}
			if found {
				inCopy = true
				currentLimit = cfg.Limit
				currentOrder = cfg.Order
				currentCount = 0
				if currentOrder == "DESC" {
					currentBuffer = newRingBuffer(currentLimit)
				} else {
					currentBuffer = nil
				}
			}
		}

		bw.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	if err := bw.Flush(); err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	if err := writer.Close(); err != nil {
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	reader.Close()

	duration := time.Since(start).Seconds()
	inSize := fileSize(inputPath)
	outSize := fileSize(outputPath)
	saved := inSize - outSize

	var unmatchedNames []string
	for _, t := range targets {
		if !matched[t.Key] {
			unmatchedNames = append(unmatchedNames, t.Key.String())
		}
	}
	if len(unmatchedNames) > 0 {
		fmt.Fprintf(w, i18n.T("warn.no_match"), strings.Join(unmatchedNames, ", "))
		fmt.Fprintln(w, i18n.T("warn.check_name"))
	}

	pctSaved := 0.0
	if inSize != 0 {
		pctSaved = float64(saved) / float64(inSize) * 100
	}

	fmt.Fprintf(w, i18n.T("result.done"), duration)
	fmt.Fprintf(w, i18n.T("result.lines"), totalLines, skippedLines)
	fmt.Fprintf(w, i18n.T("result.original"), sqldump.FormatSize(float64(inSize)))
	fmt.Fprintf(w, "%s: %s\n", resultLabel, sqldump.FormatSize(float64(outSize)))
	fmt.Fprintf(w, i18n.T("result.economia_pct"), sqldump.FormatSize(float64(saved)), pctSaved)

	return nil
}

// ringBuffer é um buffer circular de tamanho fixo usado para manter os N
// últimos registros de um bloco COPY (ordem DESC), equivalente a
// collections.deque(maxlen=N) do Python.
type ringBuffer struct {
	limit  int
	buf    []string
	idx    int
	filled bool
}

func newRingBuffer(limit int) *ringBuffer {
	if limit < 0 {
		limit = 0
	}
	return &ringBuffer{limit: limit, buf: make([]string, limit)}
}

func (r *ringBuffer) Append(line string) {
	if r.limit == 0 {
		return
	}
	r.buf[r.idx] = line
	r.idx = (r.idx + 1) % r.limit
	if r.idx == 0 {
		r.filled = true
	}
}

// Lines devolve as linhas mantidas, na ordem relativa original.
func (r *ringBuffer) Lines() []string {
	if r.limit == 0 {
		return nil
	}
	if !r.filled {
		out := make([]string, r.idx)
		copy(out, r.buf[:r.idx])
		return out
	}
	out := make([]string, 0, r.limit)
	out = append(out, r.buf[r.idx:]...)
	out = append(out, r.buf[:r.idx]...)
	return out
}
