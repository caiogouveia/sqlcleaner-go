package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
)

func main() {
	i18n.SetLang(i18n.Detect(langFlagValue(os.Args[1:])))
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("err.fatal"), err)
		os.Exit(1)
	}
}

// langFlagValue procura --lang=X ou --lang X nos argumentos crus, antes do
// parsing do cobra: o idioma precisa estar resolvido antes de montar a
// árvore de comandos, já que Short/Long/descrições de flag são strings
// fixas definidas na construção dos comandos.
func langFlagValue(args []string) string {
	for i, a := range args {
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return v
		}
		if a == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
