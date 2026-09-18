package i18n

// entry guarda a mesma mensagem nos dois idiomas suportados.
type entry struct {
	pt, en string
}

var catalog = map[string]entry{
	// main.go
	"err.fatal": {"❌ Erro: %v\n", "❌ Error: %v\n"},

	// root.go
	"root.short": {
		"Utilitário para processar dumps SQL do PostgreSQL via streaming",
		"Utility to process PostgreSQL SQL dumps via streaming",
	},
	"flag.lang": {
		"Idioma das mensagens: pt ou en (padrão: detectado do sistema)",
		"Message language: pt or en (default: detected from system)",
	},

	// mensagens compartilhadas entre subcomandos
	"flag.output": {"Arquivo SQL de saída", "Output SQL file"},
	"err.file_not_found": {
		"Erro: Arquivo %s não encontrado.\n",
		"Error: file %s not found.\n",
	},
	"err.generic":   {"\n❌ Erro: %v\n", "\n❌ Error: %v\n"},
	"err.parse":     {"Erro: %v\n", "Error: %v\n"},
	"label.entrada": {"Entrada: %s\n", "Input: %s\n"},
	"label.saida":   {"Saída: %s\n", "Output: %s\n"},
	"label.alvos":   {"Alvos: %s\n", "Targets: %s\n"},
	"progress":      {"Progresso: %dM linhas processadas...\n", "Progress: %dM lines processed...\n"},
	"warn.no_match": {"\n⚠️  Aviso: nenhuma correspondência encontrada no dump para: %s\n", "\n⚠️  Warning: no match found in the dump for: %s\n"},
	"warn.check_name": {
		`Verifique nome, schema e uso de aspas (ex.: -t '"Tabela.Com.Ponto"' para nomes com ponto literal).`,
		`Check name, schema and quoting (e.g.: -t '"Table.With.Dot"' for names with a literal dot).`,
	},
	"result.done":           {"\n✅ Concluído em %.1fs!\n", "\n✅ Done in %.1fs!\n"},
	"result.lines":          {"Linhas: %d total / %d removidas\n", "Lines: %d total / %d removed\n"},
	"result.original":       {"Original: %s\n", "Original: %s\n"},
	"result.economia_pct":   {"Economia: %s (%.1f%%)\n", "Saved: %s (%.1f%%)\n"},
	"result.economia_plain": {"Economia: %s\n", "Saved: %s\n"},
	"result.label.limpo":    {"Limpo", "Cleaned"},
	"result.label.truncado": {"Truncado", "Truncated"},

	// analisar.go
	"analisar.short": {
		"Lista tabelas de um dump SQL ordenadas por tamanho",
		"Lists tables in a SQL dump sorted by size",
	},
	"analisar.flag.top": {
		"Exibir apenas as N maiores tabelas (padrão: todas)",
		"Show only the N largest tables (default: all)",
	},
	"analisar.header.tabela":    {"Tabela", "Table"},
	"analisar.header.tamanho":   {"Tamanho", "Size"},
	"analisar.header.registros": {"Registros", "Records"},
	"analisar.summary": {
		"\n%d tabelas analisadas em %.1fs (%d linhas)\n",
		"\n%d tables analyzed in %.1fs (%d lines)\n",
	},

	// remover.go
	"remover.short": {
		"Remove tabelas inteiras (estrutura + dados) de um dump SQL",
		"Removes entire tables (structure + data) from a SQL dump",
	},
	"remover.flag.tables": {
		"Lista de tabelas para remover (separadas por vírgula)",
		"List of tables to remove (comma-separated)",
	},

	// esvaziar.go
	"esvaziar.short": {
		"Remove apenas os dados de tabelas específicas, mantendo a estrutura",
		"Removes only the data of specific tables, keeping the structure",
	},
	"esvaziar.flag.tables": {
		"Lista de tabelas para esvaziar (separadas por vírgula)",
		"List of tables to empty (comma-separated)",
	},

	// truncar.go
	"truncar.short": {
		"Mantém apenas N registros por tabela",
		"Keeps only N records per table",
	},
	"truncar.flag.tables": {
		`Tabelas, limites e ordem: "tabela", "tabela:N", "tabela:ORDEM" ou "tabela:N:ORDEM". ` +
			`ORDEM é ASC (padrão, mantém os primeiros N) ou DESC (mantém os últimos N). ` +
			`Ex: log:1000:DESC,cache:500,other:ASC`,
		`Tables, limits and order: "table", "table:N", "table:ORDER" or "table:N:ORDER". ` +
			`ORDER is ASC (default, keeps the first N) or DESC (keeps the last N). ` +
			`Ex: log:1000:DESC,cache:500,other:ASC`,
	},
	"truncar.flag.limit": {
		"Limite padrão de registros para tabelas sem limite explícito",
		"Default record limit for tables without an explicit limit",
	},
	"truncar.target_header": {
		"  %s: manter %d registros (%s)\n",
		"  %s: keep %d records (%s)\n",
	},

	// set.go
	"set.short": {
		"Remove linhas 'SET <parametro> = ...;' de um dump SQL",
		"Removes 'SET <parameter> = ...;' lines from a SQL dump",
	},
	"set.flag.params": {
		"Parâmetros a remover, separados por vírgula (padrão: transaction_timeout)",
		"Parameters to remove, comma-separated (default: transaction_timeout)",
	},
	"set.label.params": {"Parâmetros: %s\n", "Parameters: %s\n"},

	// internal/sqldump/truncate.go
	"err.truncate.invalid_limit": {
		"limite inválido '%s' para tabela '%s'",
		"invalid limit '%s' for table '%s'",
	},
	"err.truncate.invalid_order": {
		"ordem inválida '%s' para tabela '%s'. Use ASC ou DESC.",
		"invalid order '%s' for table '%s'. Use ASC or DESC.",
	},
	"err.truncate.invalid_format": {
		"formato inválido para '%s'. Use tabela, tabela:N, tabela:ORDEM ou tabela:N:ORDEM.",
		"invalid format for '%s'. Use table, table:N, table:ORDER or table:N:ORDER.",
	},
	"err.truncate.missing_limit": {
		"tabela '%s' sem limite e -n não foi informado.",
		"table '%s' has no limit and -n was not provided.",
	},

	// restaurar.go
	"restaurar.short": {
		"Restaura um dump SQL num servidor PostgreSQL",
		"Restores a SQL dump into a PostgreSQL server",
	},
	"restaurar.flag.dsn": {
		"String de conexão PostgreSQL (padrão: variáveis PGHOST/PGPORT/PGUSER/PGPASSWORD/PGDATABASE, como no psql)",
		"PostgreSQL connection string (default: PGHOST/PGPORT/PGUSER/PGPASSWORD/PGDATABASE env vars, like psql)",
	},
	"restaurar.flag.create_db": {
		"Cria o banco de destino se ele ainda não existir",
		"Creates the target database if it does not exist yet",
	},
	"restaurar.flag.dry_run": {
		"Apenas indexa o dump e mostra o que seria restaurado, sem aplicar nada no banco",
		"Only indexes the dump and shows what would be restored, without applying anything to the database",
	},
	"restaurar.flag.yes": {
		"Pula a confirmação antes de restaurar sobre um banco com tabelas existentes",
		"Skips the confirmation before restoring into a database with existing tables",
	},
	"restaurar.flag.jobs": {
		"Paralelismo da fase de dados (tabelas restauradas simultaneamente)",
		"Data-phase parallelism (tables restored concurrently)",
	},
	"restaurar.summary.tables": {
		"\n%d tabelas encontradas no dump:\n",
		"\n%d tables found in the dump:\n",
	},
	"restaurar.dry_run.done": {
		"\n(--dry-run: nada foi aplicado no banco)",
		"\n(--dry-run: nothing was applied to the database)",
	},
	"restaurar.confirm.prompt": {
		"\n⚠️  O banco '%s' já tem tabelas. Continuar e restaurar por cima mesmo assim? [y/N] ",
		"\n⚠️  Database '%s' already has tables. Continue and restore over it anyway? [y/N] ",
	},
	"restaurar.confirm.aborted": {
		"Operação cancelada.",
		"Operation cancelled.",
	},
	"restaurar.table.done": {
		"  ✅ %-40s %s\n",
		"  ✅ %-40s %s\n",
	},
	"restaurar.table.error": {
		"  ❌ %-40s erro: %v\n",
		"  ❌ %-40s error: %v\n",
	},
}
