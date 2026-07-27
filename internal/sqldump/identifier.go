package sqldump

// QualifiedName representa um nome "tabela" ou "schema.tabela", com suporte a
// identificadores citados (aspas duplas) como o pg_dump usa para nomes com
// maiúsculas ou caracteres especiais. HasSchema distingue "sem schema" (None
// no Python) de um schema vazio, o que importa para o casamento com "public".
type QualifiedName struct {
	Schema    string
	HasSchema bool
	Table     string
}

// String devolve "schema.tabela" ou apenas "tabela", usado nas mensagens de aviso.
func (q QualifiedName) String() string {
	if q.HasSchema {
		return q.Schema + "." + q.Table
	}
	return q.Table
}

// ReadIdentifier lê um identificador (com ou sem aspas duplas) a partir de s[pos:].
// Opera sobre runas para lidar corretamente com identificadores acentuados.
func ReadIdentifier(s []rune, pos int) (string, int) {
	if s[pos] == '"' {
		end := pos + 1
		for s[end] != '"' {
			end++
		}
		return string(s[pos+1 : end]), end + 1
	}
	end := pos
	for end < len(s) {
		c := s[end]
		if c == '.' || c == ' ' || c == '\t' || c == '(' {
			break
		}
		end++
	}
	return string(s[pos:end]), end
}

// ParseQualifiedName parseia "tabela", "schema.tabela", `"Tabela"`,
// `schema."Tabela"` ou `"Schema"."Tabela"` em um QualifiedName.
func ParseQualifiedName(token string) QualifiedName {
	runes := []rune(token)
	name1, pos := ReadIdentifier(runes, 0)
	if pos < len(runes) && runes[pos] == '.' {
		name2, _ := ReadIdentifier(runes, pos+1)
		return QualifiedName{Schema: name1, HasSchema: true, Table: name2}
	}
	return QualifiedName{Table: name1}
}

// FindMatchingTarget procura, na ordem, um alvo cujo schema/tabela case com
// (schema, table). Um alvo sem schema casa com "public" ou sem schema.
func FindMatchingTarget(schema string, hasSchema bool, table string, targets []QualifiedName) (QualifiedName, bool) {
	for _, t := range targets {
		if t.Table != table {
			continue
		}
		if !t.HasSchema {
			if !hasSchema || schema == "public" {
				return t, true
			}
		} else if hasSchema && schema == t.Schema {
			return t, true
		}
	}
	return QualifiedName{}, false
}
