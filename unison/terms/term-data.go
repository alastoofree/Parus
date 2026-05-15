// parus/unison/terms/term-data.go

package terms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"parus/unison")

// LoadTermSatellite загружает JSON-спутник терма
func LoadTermSatellite(termPath string) (map[string]string, error) {
jsonPath := strings.TrimSuffix(termPath, filepath.Ext(termPath)) + ".json"
data, err := os.ReadFile(jsonPath)
if err != nil {return nil, err}
var result map[string]string
if err := json.Unmarshal(data, &result); err != nil {
    return nil, fmt.Errorf("parse satellite %s: %w", jsonPath, err)}
return result, nil }
// WriteTermSatellite записывает JSON-спутник терма
func WriteTermSatellite(termPath string, data map[string]string) error {
jsonPath := strings.TrimSuffix(termPath, filepath.Ext(termPath)) + ".json"
b, err := json.MarshalIndent(data, "", "  ")
if err != nil {return fmt.Errorf("marshal satellite: %w", err)}
return os.WriteFile(jsonPath, b, 0644) }
// LoadTermSatelliteSvg загружает SVG-спутник терма
func LoadTermSatelliteSvg(termPath string) (string, error) {
svgPath := strings.TrimSuffix(termPath, filepath.Ext(termPath)) + ".svg"
data, err := os.ReadFile(svgPath)
if err != nil {return "", err}
return string(data), nil }
// WriteTermSatelliteSvg записывает SVG-спутник терма
func WriteTermSatelliteSvg(termPath string, svg string) error {
svgPath := strings.TrimSuffix(termPath, filepath.Ext(termPath)) + ".svg"
return os.WriteFile(svgPath, []byte(svg), 0644) }
// FindTermByName находит путь терма по его имени
func FindTermByName(name string) (string, bool) {
terms, err := ListTerms()
if err != nil {return "", false}
for _, t := range terms {
    if strings.EqualFold(TermPairRight(t, "name"), name) {return t.Path, true}}
return "", false }
// ResolveTermContent возвращает контент терма с резолвенными ссылками и трансклюзиями
func ResolveTermContent(path string) (string, string, error) {
f, err := LoadTerm(path)
if err != nil {return "", "", fmt.Errorf("load term: %w", err)}
content  := TermPairRight(f, "content")
allTerms, err := ListTerms()
if err != nil {return content, content, nil}
nameMap := make(map[string]string)
for _, t := range allTerms {
    name := TermPairRight(t, "name")
    if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
raw := content
result := regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllStringFunc(content, func(match string) string {
    name := strings.TrimSpace(match[2 : len(match)-2])
    tp, ok := nameMap[strings.ToLower(name)]
    if !ok {return name}
    return `<a href="term.html?path=` + tp + `">` + name + `</a>`})
result = regexp.MustCompile(`\{\{([^}]+)\}\}`).ReplaceAllStringFunc(result, func(match string) string {
    name := strings.TrimSpace(match[2 : len(match)-2])
    tp, ok := nameMap[strings.ToLower(name)]
    if !ok {return ""}
    resolved, _, _ := ResolveTermContent(tp)
    return resolved})
return result, raw, nil }
// ResolveTermProduct возвращает продукции терма для отображения
func ResolveTermProduct(path string) (map[string]string, error) {
f, err := LoadTerm(path)
if err != nil {return nil, fmt.Errorf("load term: %w", err)}
showProduct := TermPairRight(f, "show-product")
if showProduct == "" {return map[string]string{}, nil}
sat, err := LoadTermSatellite(path)
if err != nil {return map[string]string{}, nil}
result := make(map[string]string)
for _, key := range unison.ParseTokenList(showProduct) {
    if val, ok := sat[key]; ok {result[key] = val}}
return result, nil }