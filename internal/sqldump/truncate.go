package sqldump

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caiogouveia/sqlcleaner-go/internal/i18n"
)

// TruncateConfig é o limite e a ordem de corte configurados para uma tabela.
type TruncateConfig struct {
	Limit int
	Order string // "ASC" ou "DESC"
}

// TruncateTarget associa uma tabela-alvo (com schema opcional) à sua configuração.
// É mantido como slice (não map) para preservar a ordem de entrada do usuário em
// -t, igual ao dict de inserção-ordenada do Python (usado no relatório de "sem
// correspondência" e para resolver ambiguidades de casamento na mesma ordem).
type TruncateTarget struct {
	Key    QualifiedName
	Config TruncateConfig
}

// ParseTruncateTargets parseia o valor de -t de truncar_tabelas.py:
//
//	"tabela"           -> usa defaultLimit, ordem ASC
//	"tabela:N"         -> usa N como limite, ordem ASC
//	"tabela:ORDEM"     -> usa defaultLimit, ordem ASC ou DESC
//	"tabela:N:ORDEM"   -> usa N como limite, ordem ASC ou DESC
func ParseTruncateTargets(tablesArg string, defaultLimit *int) ([]TruncateTarget, error) {
	var targets []TruncateTarget
	for _, entry := range strings.Split(tablesArg, ",") {
		entry = strings.TrimSpace(entry)
		parts := strings.Split(entry, ":")
		name := parts[0]
		rest := parts[1:]

		order := "ASC"
		limit := -1
		hasLimit := false

		switch len(rest) {
		case 0:
			// nada a fazer
		case 1:
			if isOrder(rest[0]) {
				order = strings.ToUpper(rest[0])
			} else {
				n, err := strconv.Atoi(rest[0])
				if err != nil {
					return nil, fmt.Errorf(i18n.T("err.truncate.invalid_limit"), rest[0], name)
				}
				limit = n
				hasLimit = true
			}
		case 2:
			n, err := strconv.Atoi(rest[0])
			if err != nil {
				return nil, fmt.Errorf(i18n.T("err.truncate.invalid_limit"), rest[0], name)
			}
			limit = n
			hasLimit = true
			if !isOrder(rest[1]) {
				return nil, fmt.Errorf(i18n.T("err.truncate.invalid_order"), rest[1], name)
			}
			order = strings.ToUpper(rest[1])
		default:
			return nil, fmt.Errorf(i18n.T("err.truncate.invalid_format"), entry)
		}

		if !hasLimit {
			if defaultLimit == nil {
				return nil, fmt.Errorf(i18n.T("err.truncate.missing_limit"), name)
			}
			limit = *defaultLimit
		}

		targets = append(targets, TruncateTarget{
			Key:    ParseQualifiedName(name),
			Config: TruncateConfig{Limit: limit, Order: order},
		})
	}
	return targets, nil
}

func isOrder(s string) bool {
	u := strings.ToUpper(s)
	return u == "ASC" || u == "DESC"
}

// FindConfig procura, na ordem, o alvo cujo schema/tabela case com (schema, table).
func FindConfig(schema string, hasSchema bool, table string, targets []TruncateTarget) (QualifiedName, TruncateConfig, bool) {
	for _, t := range targets {
		key := t.Key
		if key.Table != table {
			continue
		}
		if !key.HasSchema {
			if !hasSchema || schema == "public" {
				return key, t.Config, true
			}
		} else if hasSchema && schema == key.Schema {
			return key, t.Config, true
		}
	}
	return QualifiedName{}, TruncateConfig{}, false
}
