package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sqlcleaner",
		Short: "Utilitários para processar dumps SQL do PostgreSQL via streaming",
	}
	root.AddCommand(newAnalisarCmd())
	root.AddCommand(newRemoverCmd())
	root.AddCommand(newEsvaziarCmd())
	root.AddCommand(newTruncarCmd())
	root.AddCommand(newSetCmd())
	return root
}

// Execute roda o comando raiz a partir dos argumentos do processo.
func Execute() error {
	return newRootCmd().Execute()
}
