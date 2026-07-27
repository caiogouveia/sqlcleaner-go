# sqlcleaner-go

Port em Go do [`sqlcleaner`](https://github.com/caiogouveia/sqlcleaner) (Python): utilitários para processar dumps SQL do PostgreSQL via streaming, sem carregar o arquivo inteiro na memória — visando um binário único, estático, sem exigir Python nem `pigz` instalados no destino.

**AVISO:** Este é um software experimental. Desenvolvido principalmente com IA. Use por sua conta e risco.

Consulte `docs/plano-migracao-go.md` para o histórico de decisões arquiteturais deste port.

## Instalação / build

```bash
go build -o sqlcleaner ./cmd/sqlcleaner
```

Requer apenas o toolchain do Go — nenhuma dependência de sistema (o gzip paralelo é feito em Go puro via `pgzip`, sem depender do binário `pigz`).

## Uso

```bash
sqlcleaner analisar dump.sql[.gz] [-n TOP]
sqlcleaner remover  dump.sql[.gz] -o saida.sql[.gz] -t tabela1,tabela2
sqlcleaner esvaziar dump.sql[.gz] -o saida.sql[.gz] -t tabela1,tabela2
sqlcleaner truncar  dump.sql[.gz] -o saida.sql[.gz] -t tabela:N:ORDEM -n LIMITE_PADRAO
sqlcleaner set      dump.sql[.gz] -o saida.sql[.gz] -p parametro1,parametro2
```

### `analisar` — lista tabelas por tamanho

Mede bytes e registros de cada bloco `COPY`, ordena da maior para a menor.

```bash
sqlcleaner analisar dump.sql
sqlcleaner analisar dump.sql.gz -n 20
```

### `remover` — remove tabelas completamente

Remove dados (`COPY`), estrutura (`CREATE TABLE`, `CREATE SEQUENCE`) e objetos relacionados (`ALTER TABLE`/`ALTER SEQUENCE`, `CREATE INDEX`).

```bash
sqlcleaner remover dump.sql -o dump_limpo.sql -t log,cache
```

**Limitação:** não remove tabelas com dependências de chave estrangeira em tabelas que permanecem no dump.

### `esvaziar` — apaga apenas os dados

Mantém `CREATE TABLE`, sequências, índices e constraints; remove só as linhas dentro de `COPY … \.`, preservando cabeçalho e terminador.

```bash
sqlcleaner esvaziar dump.sql -o dump_sem_dados.sql -t log,cache
```

### `truncar` — limita registros por tabela

Mantém estrutura e N registros por tabela, descartando o restante. Limite global (`-n`) e/ou por tabela (`tabela:N`) são combináveis. Formato de cada entrada em `-t`: `tabela`, `tabela:N`, `tabela:ORDEM` ou `tabela:N:ORDEM`. `ORDEM` é `ASC` (padrão, mantém as **primeiras** N linhas do bloco `COPY`) ou `DESC` (mantém as **últimas** N — útil para tabelas tipo log/audit quando se quer os registros mais recentes). A ordem é posicional dentro do `COPY`, não um `ORDER BY` de coluna.

```bash
sqlcleaner truncar dump.sql -o saida.sql -t log:1000,cache -n 500
sqlcleaner truncar dump.sql -o saida.sql -t log:1000:DESC,cache
sqlcleaner truncar dump.sql -o saida.sql -t audit_log:DESC -n 200
```

### `set` — remove parâmetros SET

Remove linhas `SET <parametro> = ...;` (ou `TO ...;`) — útil quando o dump foi gerado por uma versão do PostgreSQL mais nova que o servidor de destino (ex.: `transaction_timeout`).

```bash
sqlcleaner set dump.sql -o dump_limpo.sql
sqlcleaner set dump.sql.gz -o dump_limpo.sql.gz -p transaction_timeout,idle_in_transaction_session_timeout
```

## Identificadores citados / nomes com maiúsculas ou ponto literal

Assim como o `pg_dump`, os identificadores podem vir citados entre aspas duplas (`"Tabela"`, `"Schema"."Tabela"`) para nomes com maiúsculas ou caracteres especiais. Sem schema explícito, `tabela` casa com `public.tabela` ou sem qualificação. Um ponto sem aspas em `-t` é sempre separador schema.tabela — nomes de tabela com ponto literal precisam de aspas (dentro do shell: `-t '"nome.com.ponto"'`).

## Testes

```bash
go test ./...
```

## Arquitetura

O pacote `internal/sqldump` concentra a lógica compartilhada entre os subcomandos: parsing de identificadores (`ParseQualifiedName`/`ReadIdentifier`), casamento de nomes-alvo (`FindMatchingTarget`/`FindConfig`), leitura/escrita de `.sql`/`.sql.gz` (`OpenReader`/`OpenWriter`, via `pgzip`) e formatação de tamanhos (`FormatSize`). Os subcomandos ficam em `cmd/sqlcleaner/`, um por script portado.
