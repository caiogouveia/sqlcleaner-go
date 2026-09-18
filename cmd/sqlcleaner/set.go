package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	var output string
	var params string
	cmd := &cobra.Command{
		Use:   "set <entrada>",
		Short: i18n.T("set.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(os.Stdout, args[0], output, params)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", i18n.T("flag.output"))
	cmd.Flags().StringVarP(&params, "params", "p", "transaction_timeout", i18n.T("set.flag.params"))
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
		fmt.Fprintf(w, i18n.T("err.file_not_found"), inputPath)
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
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}
	writer, err := sqldump.OpenWriter(outputPath)
	if err != nil {
		reader.Close()
		fmt.Fprintf(w, i18n.T("err.generic"), err)
		return nil
	}

	fmt.Fprintln(w, "--- SQL SET Remover 1.0 ---")
	fmt.Fprintf(w, i18n.T("label.entrada"), inputPath)
	fmt.Fprintf(w, i18n.T("label.saida"), outputPath)
	fmt.Fprintf(w, i18n.T("set.label.params"), strings.Join(params, ", "))
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

	fmt.Fprintf(w, i18n.T("result.done"), duration)
	fmt.Fprintf(w, i18n.T("result.lines"), count, removed)
	fmt.Fprintf(w, i18n.T("result.original"), sqldump.FormatSize(float64(inSize)))
	fmt.Fprintf(w, i18n.T("label.saida"), sqldump.FormatSize(float64(outSize)))
	fmt.Fprintf(w, i18n.T("result.economia_plain"), sqldump.FormatSize(float64(saved)))

	return nil
}
