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

func newRemoverCmd() *cobra.Command {
	var output string
	var tables string
	cmd := &cobra.Command{
		Use:   "remover <entrada>",
		Short: "Remove tabelas inteiras (estrutura + dados) de um dump SQL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemover(os.Stdout, args[0], output, splitCommaList(tables))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "Arquivo SQL de saída")
	cmd.Flags().StringVarP(&tables, "tables", "t", "", "Lista de tabelas para remover (separadas por vírgula)")
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("tables")
	return cmd
}

func runRemover(w io.Writer, inputPath, outputPath string, tableArgs []string) error {
	targets := parseTargetTables(tableArgs)
	targetSequences := make([]sqldump.QualifiedName, len(targets))
	for i, t := range targets {
		targetSequences[i] = sqldump.QualifiedName{Schema: t.Schema, HasSchema: t.HasSchema, Table: t.Table + "_id_seq"}
	}

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

	fmt.Fprintln(w, "--- SQL Cleaner 2.0 ---")
	fmt.Fprintf(w, "Entrada: %s\n", inputPath)
	fmt.Fprintf(w, "Saída:   %s\n", outputPath)
	fmt.Fprintf(w, "Alvos:   %s\n", strings.Join(tableArgs, ", "))
	fmt.Fprintln(w, "-----------------------")

	bw := bufio.NewWriterSize(writer, 1<<20)
	scanner := sqldump.NewLineScanner(reader)

	skipUntilDot := false
	skippingBlock := false
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
			}
			continue
		}
		if skippingBlock {
			skippedLines++
			if strings.HasSuffix(trimmed, ";") {
				skippingBlock = false
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
					skippedLines++
					continue
				}
			}
		}

		shouldSkip := false
		parts := strings.Fields(trimmed)

		switch {
		case strings.HasPrefix(line, "CREATE TABLE "):
			if len(parts) >= 3 {
				qn := sqldump.ParseQualifiedName(parts[2])
				if target, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targets); found {
					matched[target] = true
					shouldSkip = true
				}
			}

		case strings.HasPrefix(line, "CREATE SEQUENCE "):
			if len(parts) >= 3 {
				qn := sqldump.ParseQualifiedName(parts[2])
				if _, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targetSequences); found {
					shouldSkip = true
				}
			}

		case strings.HasPrefix(line, "ALTER TABLE "):
			idx := 2
			if len(parts) > 2 && parts[2] == "ONLY" {
				idx = 3
			}
			if len(parts) > idx {
				qn := sqldump.ParseQualifiedName(parts[idx])
				if target, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targets); found {
					matched[target] = true
					shouldSkip = true
				}
			}

		case strings.HasPrefix(line, "ALTER SEQUENCE "):
			idx := 2
			if len(parts) > 2 && parts[2] == "ONLY" {
				idx = 3
			}
			if len(parts) > idx {
				qn := sqldump.ParseQualifiedName(parts[idx])
				if _, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targetSequences); found {
					shouldSkip = true
				}
			}

		case strings.HasPrefix(line, "CREATE INDEX "):
			onIdx := indexOf(parts, "ON")
			if onIdx >= 0 && len(parts) > onIdx+1 {
				qn := sqldump.ParseQualifiedName(parts[onIdx+1])
				if target, found := sqldump.FindMatchingTarget(qn.Schema, qn.HasSchema, qn.Table, targets); found {
					matched[target] = true
					shouldSkip = true
				}
			}
		}

		if shouldSkip {
			skippedLines++
			if !strings.HasSuffix(trimmed, ";") {
				skippingBlock = true
			}
			continue
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
