package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newAnalisarCmd() *cobra.Command {
	var top int
	cmd := &cobra.Command{
		Use:   "analisar <entrada>",
		Short: "Lista tabelas de um dump SQL ordenadas por tamanho",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalisar(os.Stdout, args[0], top)
		},
	}
	cmd.Flags().IntVarP(&top, "top", "n", 0, "Exibir apenas as N maiores tabelas (padrão: todas)")
	return cmd
}

func runAnalisar(w io.Writer, inputPath string, top int) error {
	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(w, "Erro: Arquivo %s não encontrado.\n", inputPath)
		return nil
	}

	start := time.Now()

	reader, err := sqldump.OpenReader(inputPath)
	if err != nil {
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}
	defer reader.Close()

	fmt.Fprintln(w, "--- SQL Table Analyzer 1.0 ---")
	fmt.Fprintf(w, "Entrada: %s\n", inputPath)
	fmt.Fprintln(w, "------------------------------")

	scanner := sqldump.NewLineScanner(reader)
	tableSizes := map[string]int64{}
	tableRows := map[string]int64{}
	var order []string
	seen := map[string]bool{}

	var currentTable string
	var currentBytes int64
	var currentRows int64
	inCopy := false
	count := 0

	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}
		count++
		if count%5000000 == 0 {
			fmt.Fprintf(w, "Progresso: %dM linhas processadas...\n", count/1000000)
		}

		if inCopy {
			if strings.TrimSpace(line) == `\.` {
				tableSizes[currentTable] += currentBytes
				tableRows[currentTable] += currentRows
				if !seen[currentTable] {
					seen[currentTable] = true
					order = append(order, currentTable)
				}
				inCopy = false
				currentTable = ""
				currentBytes = 0
				currentRows = 0
			} else {
				currentBytes += int64(len(line))
				currentRows++
			}
			continue
		}

		if strings.HasPrefix(line, "COPY ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentTable = strings.ReplaceAll(parts[1], "public.", "")
				currentBytes = 0
				currentRows = 0
				inCopy = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}

	duration := time.Since(start).Seconds()

	type row struct {
		table string
		size  int64
	}
	rows := make([]row, 0, len(order))
	for _, t := range order {
		rows = append(rows, row{t, tableSizes[t]})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].size > rows[j].size })
	if top > 0 && top < len(rows) {
		rows = rows[:top]
	}

	var total int64
	var totalRows int64
	for _, s := range tableSizes {
		total += s
	}
	for _, r := range tableRows {
		totalRows += r
	}

	fmt.Fprintf(w, "\n%-50s %10s  %12s  %6s\n", "Tabela", "Tamanho", "Registros", "%")
	fmt.Fprintln(w, strings.Repeat("-", 84))
	for _, r := range rows {
		pct := 0.0
		if total != 0 {
			pct = float64(r.size) / float64(total) * 100
		}
		fmt.Fprintf(w, "%-50s %10s  %12s  %5.1f%%\n", r.table, sqldump.FormatSize(float64(r.size)), formatThousands(tableRows[r.table]), pct)
	}
	fmt.Fprintln(w, strings.Repeat("-", 84))
	fmt.Fprintf(w, "%-50s %10s  %12s\n", "TOTAL", sqldump.FormatSize(float64(total)), formatThousands(totalRows))
	fmt.Fprintf(w, "\n%d tabelas analisadas em %.1fs (%d linhas)\n", len(tableSizes), duration, count)

	return nil
}
