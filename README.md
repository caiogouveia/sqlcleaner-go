# sqlcleaner-go

*[Read in English](README.en.md)*

Port em Go do [`sqlcleaner`](https://github.com/caiogouveia/sqlcleaner) (Python): utilitários para processar dumps SQL do PostgreSQL via streaming, sem carregar o arquivo inteiro na memória — visando um binário único, estático, sem exigir Python nem `pigz` instalados no destino. `analisar`/`remover`/`esvaziar`/`truncar`/`set` operam 100% localmente, sem nunca abrir conexão com um banco; a exceção é `restaurar`, que aplica o dump (já limpo pelos outros comandos, se for o caso) diretamente num servidor PostgreSQL via driver nativo (`pgx`), também sem depender de `psql`/`pg_restore` instalados.

**AVISO:** Este é um software experimental. Desenvolvido principalmente com IA. Use por sua conta e risco.

Consulte `docs/plano-migracao-go.md` para o histórico de decisões arquiteturais deste port.

## Instalação

### Via Homebrew (macOS/Linux)

```bash
brew tap caiogouveia/sqlcleaner
brew install --cask sqlcleaner-go
```

### Build a partir do código-fonte

```bash
go build -o sqlcleaner ./cmd/sqlcleaner
# ou, via Makefile:
make build
```

Requer apenas o toolchain do Go — nenhuma dependência de sistema (o gzip paralelo é feito em Go puro via `pgzip`, sem depender do binário `pigz`).

### Binário pré-compilado

Também dá pra baixar o binário direto da [página de releases](https://github.com/caiogouveia/sqlcleaner-go/releases) (linux/darwin/windows, amd64/arm64), sem precisar do Homebrew nem do toolchain Go.

## Idioma

Mensagens, ajuda (`--help`) e erros saem em português por padrão, com inglês disponível via `--lang=en`:

```bash
sqlcleaner --lang=en analisar dump.sql
# ou, pra fixar sem precisar repetir a flag:
export SQLCLEANER_LANG=en
```

Ordem de resolução: `--lang` > `SQLCLEANER_LANG` > `LC_ALL`/`LANG` do sistema > português (padrão). Os nomes dos subcomandos (`analisar`, `remover`, `esvaziar`, `truncar`, `set`) não mudam de idioma.

## Uso

```bash
sqlcleaner analisar  dump.sql[.gz] [-n TOP]
sqlcleaner remover   dump.sql[.gz] -o saida.sql[.gz] -t tabela1,tabela2
sqlcleaner esvaziar  dump.sql[.gz] -o saida.sql[.gz] -t tabela1,tabela2
sqlcleaner truncar   dump.sql[.gz] -o saida.sql[.gz] -t tabela:N:ORDEM -n LIMITE_PADRAO
sqlcleaner set       dump.sql[.gz] -o saida.sql[.gz] -p parametro1,parametro2
sqlcleaner restaurar dump.sql[.gz] --dsn postgres://usuario:senha@host:5432/banco [--create-db] [--dry-run] [-y] [--jobs N]
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

### `restaurar` — aplica um dump num servidor PostgreSQL

Restaura um dump `.sql`/`.sql.gz` (o mesmo formato produzido/editado pelos outros comandos) num servidor PostgreSQL de destino, via driver nativo (`pgx`) — não depende de `psql`/`pg_restore` instalados. Indexa o dump uma vez (trechos de SQL e blocos `COPY`, por offset de byte) e aplica em ordem: SQL sequencial, com corridas de `COPY` adjacentes em paralelo entre si (`--jobs`).

```bash
sqlcleaner restaurar dump_limpo.sql.gz --dsn postgres://usuario:senha@host:5432/banco
sqlcleaner restaurar dump.sql --dsn postgres://usuario:senha@host:5432/banco_novo --create-db
sqlcleaner restaurar dump.sql --dsn postgres://usuario:senha@host:5432/banco --dry-run
sqlcleaner restaurar dump.sql --dsn postgres://usuario:senha@host:5432/banco -y --jobs 8
```

- `--dsn`: string de conexão (URL ou `chave=valor`). Se omitida, usa as variáveis `PGHOST`/`PGPORT`/`PGUSER`/`PGPASSWORD`/`PGDATABASE` do ambiente, como o `psql`.
- `--create-db`: cria o banco de destino se ele ainda não existir.
- `--dry-run`: só indexa o dump e lista as tabelas/tamanhos que seriam restaurados, sem tocar no banco.
- `-y`/`--yes`: pula a confirmação interativa pedida quando o banco de destino já tem tabelas (evita sobrescrever dados por engano).
- `--jobs`: paralelismo da fase de dados (padrão: 4).

**Limitação:** assume um dump `pg_dump --format=plain` bem formado (não o formato custom/directory). Erros do próprio PostgreSQL durante a restauração (ex.: tabela já existe, violação de constraint) são reportados como vieram do servidor.

**Versões do PostgreSQL suportadas:** o driver `pgx/v5` usado internamente segue o mesmo range oficial de suporte do time do PostgreSQL — major releases dos últimos 5 anos, hoje **PostgreSQL 14 em diante**. Testado manualmente de ponta a ponta (dump real via `pg_dump` → `truncar` → `restaurar`) contra `postgres:14-alpine`, `postgres:16-alpine` e `postgres:18-alpine`, todos passando. Versões ≤13 provavelmente funcionam (o protocolo usado — simple/extended query e `COPY` — é estável há décadas), mas não são testadas nem suportadas oficialmente pelo `pgx`.

## Identificadores citados / nomes com maiúsculas ou ponto literal

Assim como o `pg_dump`, os identificadores podem vir citados entre aspas duplas (`"Tabela"`, `"Schema"."Tabela"`) para nomes com maiúsculas ou caracteres especiais. Sem schema explícito, `tabela` casa com `public.tabela` ou sem qualificação. Um ponto sem aspas em `-t` é sempre separador schema.tabela — nomes de tabela com ponto literal precisam de aspas (dentro do shell: `-t '"nome.com.ponto"'`).

## Testes

```bash
go test ./...
```

O teste de integração de `restaurar` (`cmd/sqlcleaner/restaurar_integration_test.go`) roda contra um PostgreSQL real: cria um banco descartável, restaura um dump nele, confere os dados e testa a confirmação de segurança. Ele só roda se a variável `SQLCLEANER_TEST_DSN` estiver definida — sem ela, é pulado (`go test ./...` funciona normalmente sem Postgres instalado).

```bash
make test-integration                                  # sobe postgres:16-alpine via Docker, roda o teste, derruba o container
make test-integration PG_TEST_IMAGE=postgres:14-alpine # testar contra outra versão
```

Ou manualmente, contra qualquer Postgres já disponível:

```bash
SQLCLEANER_TEST_DSN=postgres://postgres:postgres@localhost:5432/postgres go test ./cmd/sqlcleaner/... -run Integration -v
```

## Arquitetura

O pacote `internal/sqldump` concentra a lógica compartilhada entre os subcomandos de edição de dumps: parsing de identificadores (`ParseQualifiedName`/`ReadIdentifier`), casamento de nomes-alvo (`FindMatchingTarget`/`FindConfig`), leitura/escrita de `.sql`/`.sql.gz` (`OpenReader`/`OpenWriter`, via `pgzip`) e formatação de tamanhos (`FormatSize`). O pacote `internal/sqlrestore` cuida só de `restaurar`: indexação do dump em segmentos SQL/COPY por offset de byte, resolução de conexão (`pgconn`) e a aplicação em si (sequencial para SQL, paralela para lotes de `COPY`). Os subcomandos ficam em `cmd/sqlcleaner/`, um por script portado (mais `restaurar`, que não tem equivalente no `sqlcleaner` Python original).
