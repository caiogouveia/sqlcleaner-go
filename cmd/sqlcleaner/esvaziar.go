package main

import (
	"fmt"
	"io"
	"os"
	"strings"

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
	fmt.Fprintf(&header, "Entrada: %s\n", inputPath)
	fmt.Fprintf(&header, "Saída:   %s\n", outputPath)
	fmt.Fprintf(&header, "Alvos:   %s\n", strings.Join(tableArgs, ", "))
	fmt.Fprintln(&header, "-----------------------------")

	return runTruncarCore(w, inputPath, outputPath, targets, header.String(), "Limpo")
}
