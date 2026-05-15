// parus/unison/terms/term-search.go

package terms

import "strings"

// SearchTerms ищет query по именам алиасам и контенту термов
func SearchTerms(query string) ([]map[string]string, error) {
q := strings.ToLower(query)
allTerms, err := ListTerms()
if err != nil {return nil, err}
var results []map[string]string
for _, t := range allTerms {
    name    := TermPairRight(t, "name")
    alias   := TermPairRight(t, "alias")
    content := TermPairRight(t, "content")
    nameL    := strings.ToLower(name)
    aliasL   := strings.ToLower(alias)
    contentL := strings.ToLower(content)
    pos := strings.Index(contentL, q)
    if !strings.Contains(nameL, q) && !strings.Contains(aliasL, q) && pos < 0 {continue}
    excerpt := ""
    if pos >= 0 {
        start := pos - 80
        if start < 0 {start = 0}
        end := pos + len(q) + 80
        if end > len(content) {end = len(content)}
        excerpt = "..." + content[start:end] + "..."}
    results = append(results, map[string]string{
        "path":    t.Path,
        "name":    name,
        "excerpt": excerpt})}
return results, nil }