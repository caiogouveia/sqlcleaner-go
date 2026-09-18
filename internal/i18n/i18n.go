// Package i18n resolve o idioma das mensagens do sqlcleaner (pt/en).
//
// O pacote nasce sempre em PT e só muda de idioma se algo chamar SetLang
// explicitamente — a detecção via variáveis de ambiente (LANG/LC_ALL) só
// acontece dentro de main(), nunca em tempo de import. Isso mantém os
// testes determinísticos: eles chamam as funções runXxx diretamente, sem
// passar por main(), então nunca disparam a detecção do ambiente.
package i18n

import (
	"fmt"
	"os"
	"strings"
)

type Lang string

const (
	PT Lang = "pt"
	EN Lang = "en"
)

var current Lang = PT

// SetLang define o idioma corrente. Qualquer valor diferente de PT/EN cai
// para PT.
func SetLang(l Lang) {
	if l != PT && l != EN {
		l = PT
	}
	current = l
}

// Current devolve o idioma corrente.
func Current() Lang {
	return current
}

// Detect resolve o idioma a partir de, em ordem de prioridade: flagValue
// (tipicamente o --lang explícito do usuário), a variável SQLCLEANER_LANG,
// e por fim LC_ALL/LANG do sistema. Qualquer valor que não comece com "en"
// cai para PT (padrão do projeto).
func Detect(flagValue string) Lang {
	if flagValue != "" {
		return normalize(flagValue)
	}
	if v := os.Getenv("SQLCLEANER_LANG"); v != "" {
		return normalize(v)
	}
	if v := os.Getenv("LC_ALL"); v != "" {
		return normalize(v)
	}
	if v := os.Getenv("LANG"); v != "" {
		return normalize(v)
	}
	return PT
}

func normalize(v string) Lang {
	v = strings.ToLower(strings.TrimSpace(v))
	if strings.HasPrefix(v, "en") {
		return EN
	}
	return PT
}

// T devolve a mensagem `key` no idioma corrente, aplicando fmt.Sprintf com
// args quando houver. Chave desconhecida devolve a própria chave (visível
// o bastante em teste/uso pra ser notada e corrigida).
func T(key string, args ...any) string {
	e, ok := catalog[key]
	if !ok {
		return key
	}
	msg := e.pt
	if current == EN {
		msg = e.en
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
