package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newEsvaziarCmd() *cobra.Command {
	var output string
	var tables string
	cmd := &cobra.Command{
		Use:   "esvaziar <entrada>",
		Short: "Remove apenas os dados de tabelas específicas, mantendo a estrutura",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEsvaziar(os.Stdout, args[0], output, splitCommaList(tables))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "Arquivo SQL de saída")
	cmd.Flags().StringVarP(&tables, "tables", "t", "", "Lista de tabelas para esvaziar (separadas por vírgula)")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("tables")
	return cmd
}

func runEsvaziar(w io.Writer, inputPath, outputPath string, tableArgs []string) error {
	targets := parseTargetTables(tableArgs)

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
	writer, err := sqldump.OpenWriter(outputPath)
	if err != nil {
		reader.Close()
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}

	fmt.Fprintln(w, "--- SQL Table Emptier 1.0 ---")
	fmt.Fprintf(w, "Entrada: %s\n", inputPath)
	fmt.Fprintf(w, "Saída:   %s\n", outputPath)
	fmt.Fprintf(w, "Alvos:   %s\n", strings.Join(tableArgs, ", "))
	fmt.Fprintln(w, "-----------------------------")

	bw := bufio.NewWriterSize(writer, 1<<20)
	scanner := sqldump.NewLineScanner(reader)

	skipUntilDot := false
	count := 0
	skippedLines := 0
	matched := map[sqldump.QualifiedName]bool{}

	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}
		count++
		if count%5000000 == 0 {
			fmt.Fprintf(w, "Progresso: %dM linhas processadas...\n", count/1000000)
		}

		trimmed := strings.TrimSpace(line)

		if skipUntilDot {
			skippedLines++
			if trimmed == `\.` {
				skipUntilDot = false
				bw.WriteString(line)
			}
			continue
		}

		if strings.HasPrefix(line, "COPY ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				qn := sqldump.ParseQualifiedName(parts[1])
				if target, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targets); found {
					matched[target] = true
					skipUntilDot = true
					bw.WriteString(line)
					continue
				}
			}
		}

		bw.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}
	if err := bw.Flush(); err != nil {
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}
	if err := writer.Close(); err != nil {
		fmt.Fprintf(w, "\n❌ Erro: %v\n", err)
		return nil
	}
	reader.Close()

	duration := time.Since(start).Seconds()
	inSize := fileSize(inputPath)
	outSize := fileSize(outputPath)
	saved := inSize - outSize

	unmatched := unmatchedNames(targets, matched)
	if len(unmatched) > 0 {
		fmt.Fprintf(w, "\n⚠️  Aviso: nenhuma correspondência encontrada no dump para: %s\n", strings.Join(unmatched, ", "))
		fmt.Fprintln(w, `Verifique nome, schema e uso de aspas (ex.: -t '"Tabela.Com.Ponto"' para nomes com ponto literal).`)
	}

	pctSaved := 0.0
	if inSize != 0 {
		pctSaved = float64(saved) / float64(inSize) * 100
	}

	fmt.Fprintf(w, "\n✅ Concluído em %.1fs!\n", duration)
	fmt.Fprintf(w, "Linhas:     %d total / %d removidas\n", count, skippedLines)
	fmt.Fprintf(w, "Original:   %s\n", sqldump.FormatSize(float64(inSize)))
	fmt.Fprintf(w, "Limpo:      %s\n", sqldump.FormatSize(float64(outSize)))
	fmt.Fprintf(w, "Economia:   %s (%.1f%%)\n", sqldump.FormatSize(float64(saved)), pctSaved)

	return nil
}
