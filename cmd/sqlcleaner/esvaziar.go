package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/caiogouveia/sqlcleaner-go/internal/sqldump"
	"github.com/spf13/cobra"
)

func newEsvaziarCmd() *cobra.Command {
	var output string
	var tables string
	cmd := &cobra.Command{
		Use:   "esvaziar <entrada>",
		Short: i18n.T("esvaziar.short"),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEsvaziar(os.Stdout, args[0], output, splitCommaList(tables))
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", i18n.T("flag.output"))
	cmd.Flags().StringVarP(&tables, "tables", "t", "", i18n.T("esvaziar.flag.tables"))
	cmd.MarkFlagRequired("output")
	cmd.MarkFlagRequired("tables")
	return cmd
}

// runEsvaziar esvazia os dados das tabelas alvo mantendo a estrutura, o que
// equivale a um runTruncarCore com limite 0 (mantém 0 registros) para cada
// tabela — por isso delega no motor de streaming de truncar.go.
func runEsvaziar(w io.Writer, inputPath, outputPath string, tableArgs []string) error {
	qns := parseTargetTables(tableArgs)
	targets := make([]sqldump.TruncateTarget, len(qns))
	for i, qn := range qns {
		targets[i] = sqldump.TruncateTarget{Key: qn, Config: sqldump.TruncateConfig{Limit: 0, Order: "ASC"}}
	}

	var header strings.Builder
	fmt.Fprintln(&header, "--- SQL Table Emptier 1.0 ---")
	fmt.Fprintf(&header, i18n.T("label.entrada"), inputPath)
	fmt.Fprintf(&header, i18n.T("label.saida"), outputPath)
	fmt.Fprintf(&header, i18n.T("label.alvos"), strings.Join(tableArgs, ", "))
	fmt.Fprintln(&header, "-----------------------------")

	return runTruncarCore(w, inputPath, outputPath, targets, header.String(), i18n.T("result.label.limpo"))
}
