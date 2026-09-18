package main

import (
	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sqlcleaner",
		Short: i18n.T("root.short"),
	}
	root.CompletionOptions.DisableDefaultCmd = true
	// --lang já foi lido e aplicado em langFlagValue()/i18n.SetLang() antes
	// deste comando ser construído; a flag só existe aqui pra o cobra
	// aceitar/documentar --lang (--help, autocomplete etc.).
	var lang string
	root.PersistentFlags().StringVar(&lang, "lang", string(i18n.Current()), i18n.T("flag.lang"))
	root.AddCommand(newAnalisarCmd())
	root.AddCommand(newRemoverCmd())
	root.AddCommand(newEsvaziarCmd())
	root.AddCommand(newTruncarCmd())
	root.AddCommand(newSetCmd())
	root.AddCommand(newRestaurarCmd())
	return root
}

// Execute roda o comando raiz a partir dos argumentos do processo.
func Execute() error {
	return newRootCmd().Execute()
}
