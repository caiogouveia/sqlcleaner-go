package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var output string
	var params string
	cmd := &cobra.Command{
		Use:   "set <entrada>",
		Short: `Remove linhas 'SET <parametro> = ...;' de um dump SQL`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(os.Stdout, args[0], output, params)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "Arquivo SQL de saída")
	cmd.Flags().StringVarP(&params, "params", "p", "transaction_timeout",
		"Parâmetros a remover, separados por vírgula (padrão: transaction_timeout)")
	cmd.MarkFlagRequired("output")
	return cmd
}

func buildSetPatterns(params []string) []*regexp.Regexp {
	pats := make([]*regexp.Regexp, len(params))
	for i, p := range params {
		pats[i] = regexp.MustCompile(`(?i)^SET\s+"?` + regexp.QuoteMeta(p) + `"?\s*(=|TO)(\s|$)`)
	}
	return pats
}

func runSet(w io.Writer, inputPath, outputPath, paramsArg string) error {
	if _, err := os.Stat(inputPath); err != nil {
		fmt.Fprintf(w, "Erro: Arquivo %s não encontrado.\n", inputPath)
		return nil
	}

	var params []string
	for _, p := range strings.Split(paramsArg, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			params = append(params, p)
		}
	}
	patterns := buildSetPatterns(params)

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

	fmt.Fprintln(w, "--- SQL SET Remover 1.0 ---")
	fmt.Fprintf(w, "Entrada: %s\n", inputPath)
	fmt.Fprintf(w, "Saída:   %s\n", outputPath)
	fmt.Fprintf(w, "Parâmetros: %s\n", strings.Join(params, ", "))
	fmt.Fprintln(w, "---------------------------")

	bw := bufio.NewWriterSize(writer, 1<<20)
	scanner := sqldump.NewLineScanner(reader)

	count := 0
	removed := 0

	for {
		line, ok := scanner.Next()
		if !ok {
			break
		}
		count++
		trimmed := strings.TrimSpace(line)

		matchedAny := false
		for _, p := range patterns {
			if p.MatchString(trimmed) {
				matchedAny = true
				break
			}
		}
		if matchedAny {
			removed++
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

	fmt.Fprintf(w, "\n✅ Concluído em %.1fs!\n", duration)
	fmt.Fprintf(w, "Linhas:   %d total / %d removidas\n", count, removed)
	fmt.Fprintf(w, "Original: %s\n", sqldump.FormatSize(float64(inSize)))
	fmt.Fprintf(w, "Saída:    %s\n", sqldump.FormatSize(float64(outSize)))
	fmt.Fprintf(w, "Economia: %s\n", sqldump.FormatSize(float64(saved)))

	return nil
}
