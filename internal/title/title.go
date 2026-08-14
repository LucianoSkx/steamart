// Pacote title normaliza títulos de jogos para comparação de match,
// portado do Decky-Metadata (backend/matching.py, GPL-3.0).
package title

import (
	"html"
	"regexp"
	"strings"
)

var (
	reBrackets  = regexp.MustCompile(`\[[^\]]+\]|\([^\)]*\)`)
	reArticles  = regexp.MustCompile(`\b(the|a|an)\b`)
	reEditions  = regexp.MustCompile(`\b(remaster(ed)?|hd|definitive|ultimate|complete|goty|edition)\b`)
	reRegionVer = regexp.MustCompile(`\b(usa|europe|eur|japan|jp|world|rev|revision|beta|proto|prototype|demo|sample|en|fr|de|es|it|pt|br|v\d+(?:\.\d+)*)\b`)
	reNonAlpha  = regexp.MustCompile(`[^a-z0-9]+`)
	reSpaces    = regexp.MustCompile(`\s+`)
)

// NormaliseTitle normaliza um título para comparação: remove marcações
// (™/®/©), colchetes e parênteses, artigos, edições (remaster/hd/goty...),
// regiões e versões.
func NormaliseTitle(s string) string {
	text := html.UnescapeString(s)
	text = strings.ToLower(text)
	text = strings.NewReplacer("™", "", "®", "", "©", "").Replace(text)
	text = reBrackets.ReplaceAllString(text, " ")
	text = reArticles.ReplaceAllString(text, " ")
	text = reEditions.ReplaceAllString(text, " ")
	text = reRegionVer.ReplaceAllString(text, " ")
	text = reNonAlpha.ReplaceAllString(text, " ")
	return strings.TrimSpace(reSpaces.ReplaceAllString(text, " "))
}

// CleanTitle limpa um título para exibição/cache: unescape de HTML e
// remoção de marcações, preservando o resto.
func CleanTitle(s string) string {
	text := html.UnescapeString(s)
	text = strings.NewReplacer("™", "", "®", "", "©", "").Replace(text)
	return strings.TrimSpace(reSpaces.ReplaceAllString(text, " "))
}

var nonPrimaryPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bdemo\b`),
	regexp.MustCompile(`(?i)\bbeta\b`),
	regexp.MustCompile(`(?i)\bplaytest\b`),
	regexp.MustCompile(`(?i)\bprototype\b`),
	regexp.MustCompile(`(?i)\bsoundtrack\b`),
	regexp.MustCompile(`(?i)\bost\b`),
	regexp.MustCompile(`(?i)\bseason\s+pass\b`),
	regexp.MustCompile(`(?i)\bdlc\b`),
	regexp.MustCompile(`(?i)\bpack\b`),
	regexp.MustCompile(`(?i)\bbundle\b`),
	regexp.MustCompile(`(?i)\bartbook\b`),
	regexp.MustCompile(`(?i)\bart\s+book\b`),
	regexp.MustCompile(`(?i)\btrailer\b`),
	regexp.MustCompile(`(?i)\bdedicated\s+server\b`),
	regexp.MustCompile(`(?i)\bserver\b`),
	regexp.MustCompile(`(?i)\btest\b`),
}

// IsNonPrimaryTitle indica se o nome é uma variante não-primária
// (demo, beta, soundtrack, DLC, bundle, etc.) que deve perder prioridade
// no match.
func IsNonPrimaryTitle(name string) bool {
	for _, re := range nonPrimaryPatterns {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

// DistinctiveTokensPresent verifica se os tokens distintivos da consulta
// (3+ caracteres ou numéricos) estão todos presentes no candidato.
func DistinctiveTokensPresent(query, candidate string) bool {
	q := strings.Fields(query)
	c := strings.Fields(candidate)
	set := map[string]bool{}
	for _, t := range c {
		set[t] = true
	}
	for _, t := range q {
		if len(t) < 3 && !isAllDigits(t) {
			continue
		}
		if !set[t] {
			return false
		}
	}
	return true
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

// LevRatio devolve a similaridade Levenshtein normalizada (1 = idênticos),
// aproximação do difflib.SequenceMatcher.ratio do Decky-Metadata.
func LevRatio(a, b string) float64 {
	if a == b {
		return 1
	}
	la, lb := len(a), len(b)
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			m := prev[j] + 1
			if cur[j-1]+1 < m {
				m = cur[j-1] + 1
			}
			if prev[j-1]+cost < m {
				m = prev[j-1] + cost
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	max := la
	if lb > max {
		max = lb
	}
	if max == 0 {
		return 0
	}
	return 1 - float64(prev[lb])/float64(max)
}
