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

func newRemoverCmd() *cobra.Command {
	var output string
	var tables string
	cmd := &cobra.Command{
		Use:   "remover <entrada>",
		Short: i18n.T("remover.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemover(os.Stdout, args[0], output, splitCommaList(tables))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", i18n.T("flag.output"))
	cmd.Flags().StringVarP(&tables, "tables", "t", "", i18n.T("remover.flag.tables"))
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
		fmt.Fprintf(w, i18n.T("err.file_not_found"), inputPath)
		return nil
	}

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

	fmt.Fprintln(w, "--- SQL Cleaner 2.0 ---")
	fmt.Fprintf(w, i18n.T("label.entrada"), inputPath)
	fmt.Fprintf(w, i18n.T("label.saida"), outputPath)
	fmt.Fprintf(w, i18n.T("label.alvos"), strings.Join(tableArgs, ", "))
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
			fmt.Fprintf(w, i18n.T("progress"), count/1000000)
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

	unmatched := unmatchedNames(targets, matched)
	if len(unmatched) > 0 {
		fmt.Fprintf(w, i18n.T("warn.no_match"), strings.Join(unmatched, ", "))
		fmt.Fprintln(w, i18n.T("warn.check_name"))
	}

	pctSaved := 0.0
	if inSize != 0 {
		pctSaved = float64(saved) / float64(inSize) * 100
	}

	fmt.Fprintf(w, i18n.T("result.done"), duration)
	fmt.Fprintf(w, i18n.T("result.lines"), count, skippedLines)
	fmt.Fprintf(w, i18n.T("result.original"), sqldump.FormatSize(float64(inSize)))
	fmt.Fprintf(w, "%s: %s\n", i18n.T("result.label.limpo"), sqldump.FormatSize(float64(outSize)))
	fmt.Fprintf(w, i18n.T("result.economia_pct"), sqldump.FormatSize(float64(saved)), pctSaved)

	return nil
}
