BINARY := sqlcleaner

PG_TEST_IMAGE     := postgres:16-alpine
PG_TEST_CONTAINER := sqlcleaner-pg-test
PG_TEST_PORT      := 55432
PG_TEST_DSN       := postgres://postgres:postgres@localhost:$(PG_TEST_PORT)/postgres

.PHONY: all build test vet clean test-integration

all: test build

build:
	go build -o $(BINARY) ./cmd/sqlcleaner

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

# test-integration roda o teste de integração de `restaurar`
# (cmd/sqlcleaner/restaurar_integration_test.go) contra um PostgreSQL real
# descartável, subido via Docker. Some com o container ao final, mesmo se o
# teste falhar. Use PG_TEST_IMAGE=postgres:14-alpine make test-integration
# para rodar contra outra versão.
test-integration:
	@docker rm -f $(PG_TEST_CONTAINER) >/dev/null 2>&1 || true
	docker run -d --rm --name $(PG_TEST_CONTAINER) \
		-e POSTGRES_PASSWORD=postgres -p $(PG_TEST_PORT):5432 $(PG_TEST_IMAGE) >/dev/null
	@echo "Aguardando $(PG_TEST_IMAGE) ficar pronto..."
	@until docker exec $(PG_TEST_CONTAINER) pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
	SQLCLEANER_TEST_DSN=$(PG_TEST_DSN) go test ./cmd/sqlcleaner/... -run Integration -v; \
	status=$$?; \
	docker stop $(PG_TEST_CONTAINER) >/dev/null; \
	exit $$status
