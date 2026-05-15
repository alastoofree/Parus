// parus/unison/terms/term-index.go

package terms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"sort"
	"parus/unison")

// ListTerms рекурсивно обходит KATER и возвращает все термы
func ListTerms() ([]TermFrame, error) {
var terms []TermFrame
err := filepath.Walk(TermBaseDir, func(path string, info os.FileInfo, err error) error {
	if err != nil {return err}
	if info.IsDir() {return nil}
	if !strings.HasSuffix(path, ".term") {return nil}
	f, err := LoadTerm(path)
	if err != nil {return nil}
	terms = append(terms, f)
	return nil})
return terms, err }
// parseTermSlotLists разбирает слот с двумя списками [ прямые ] [ обратные ]
func parseTermSlotLists(raw string) (direct []string, inverse []string) {
tokens := unison.ParseTokenList(raw)
i := 0
listIdx := 0
for i < len(tokens) {
    if tokens[i] == "[" {
        i++
        var list []string
        for i < len(tokens) && tokens[i] != "]" {
            list = append(list, tokens[i])
            i++}
        if i < len(tokens) {i++}
        if listIdx == 0 {direct = list} else {inverse = list}
        listIdx++
    } else {i++}}
return }
// fmtTermSlotLists форматирует два списка в строку слота
func fmtTermSlotLists(direct, inverse []string) string {
return "[ " + strings.Join(direct, " ") + " ] [ " + strings.Join(inverse, " ") + " ]" }
// SyncTermAliasGroup синхронизирует алиас-группу термина
func SyncTermAliasGroup(currentPath string, newAliases []string) error {
allTerms, err := ListTerms()
if err != nil {return fmt.Errorf("list terms: %w", err)}
nameMap := make(map[string]string)
for _, t := range allTerms {
	name := TermPairRight(t, "name")
	if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
group := []string{currentPath}
for _, alias := range newAliases {
	if p, ok := nameMap[strings.ToLower(alias)]; ok {
		group = unison.AppendUnique(group, p)}}
for _, path := range group {
	f, err := LoadTerm(path)
	if err != nil {continue}
	var members []string
	for _, m := range group {
		if m != path {
			mt, err2 := LoadTerm(m)
			if err2 != nil {continue}
			members = append(members, TermPairRight(mt, "name"))}}
	for i, p := range f.Pairs {
		if p.Left == "alias" {
			f.Pairs[i].Right = strings.Join(members, " "); break}}
	if err := WriteTerm(f); err != nil {
		return fmt.Errorf("sync term alias group %s: %w", path, err)}}
return nil }
// SyncTermContentLinks синхронизирует ссылки из контента
func SyncTermContentLinks(pairs []unison.SlotPair) []unison.SlotPair {
var content string
for _, p := range pairs {
    if p.Left == "content" {content = p.Right}}
linksSet := make(map[string]bool)
for _, m := range regexp.MustCompile(`\[\[([^\]]+)\]\]`).FindAllStringSubmatch(content, -1) {
    linksSet[strings.TrimSpace(m[1])] = true}
includesSet := make(map[string]bool)
for _, m := range regexp.MustCompile(`\{\{([^}]+)\}\}`).FindAllStringSubmatch(content, -1) {
    includesSet[strings.TrimSpace(m[1])] = true}
for i, p := range pairs {
    if p.Left == "links" {
        var list []string
        for k := range linksSet {list = append(list, k)}
        _, inverse := parseTermSlotLists(p.Right)
        pairs[i].Right = fmtTermSlotLists(list, inverse)}
    if p.Left == "includes" {
        var list []string
        for k := range includesSet {list = append(list, k)}
        _, inverse := parseTermSlotLists(p.Right)
        pairs[i].Right = fmtTermSlotLists(list, inverse)}}
return pairs }
// findTermByDepth ищет первый терм чьё имя совпадает с токеном по глубине deep
func findTermByDepth(token string, deep int, sortedNames []string, nameMap map[string]string) string {
t := strings.ToLower(token)
if deep <= 0 || deep > len(t) {deep = len(t)}
prefix := t[:deep]
for _, name := range sortedNames {
	if len(name) >= deep && name[:deep] == prefix {
		return nameMap[name]}}
return "" }
// SyncTermProduct вычисляет прямые и обратные связи product из жсон-спутника
func SyncTermProduct(currentPath string, allTerms []TermFrame) error {
f, err := LoadTerm(currentPath)
if err != nil {return err}
showProduct := TermPairRight(f, "show-product")
if showProduct == "" {return nil}
deepStr := TermPairRight(f, "deep")
deep := 0
fmt.Sscanf(deepStr, "%d", &deep)
sat, err := LoadTermSatellite(currentPath)
if err != nil {return nil}
nameMap := make(map[string]string)
for _, t := range allTerms {
	name := TermPairRight(t, "name")
	if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
sortedNames := make([]string, 0, len(nameMap))
for name := range nameMap {sortedNames = append(sortedNames, name)}
sort.Slice(sortedNames, func(i, j int) bool {return len(sortedNames[i]) < len(sortedNames[j])})
found := make(map[string]bool)
for _, key := range unison.ParseTokenList(showProduct) {
	val, ok := sat[key]
	if !ok {continue}
	for _, token := range strings.Fields(val) {
		tp := findTermByDepth(token, deep, sortedNames, nameMap)
		if tp != "" && tp != currentPath {
			for _, t := range allTerms {
				if t.Path == tp {
					name := TermPairRight(t, "name")
					if name != "" {found[name] = true}; break}}}}}
direct := make([]string, 0, len(found))
for name := range found {direct = append(direct, name)}
inverse := []string{}
for _, t := range allTerms {
	if t.Path == currentPath {continue}
	sat2, err2 := LoadTermSatellite(t.Path)
	if err2 != nil {continue}
	tName := TermPairRight(t, "name")
	if tName == "" {continue}
	for _, key := range unison.ParseTokenList(showProduct) {
		val, ok := sat2[key]
		if !ok {continue}
		found2 := false
		for _, token := range strings.Fields(val) {
			tp := findTermByDepth(token, deep, sortedNames, nameMap)
			if strings.TrimPrefix(tp, "./") == strings.TrimPrefix(currentPath, "./") {found2 = true; break}}
		if found2 {inverse = unison.AppendUnique(inverse, tName); break}}}
for i, p := range f.Pairs {
	if p.Left == "product" {
		f.Pairs[i].Right = fmtTermSlotLists(direct, inverse); break}}
return WriteTerm(f) }
// BuildTermIndex строит индекс для терма
func BuildTermIndex(currentPath string) error {
f, err := LoadTerm(currentPath)
if err != nil {return fmt.Errorf("load term: %w", err)}
allTerms, err := ListTerms()
if err != nil {return fmt.Errorf("list terms: %w", err)}
nameMap := make(map[string]string)
for _, t := range allTerms {
    name := TermPairRight(t, "name")
    if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
currentName := TermPairRight(f, "name")
for _, t := range allTerms {
    changed := false
    for i, p := range t.Pairs {
        if p.Left == "tags" || p.Left == "links" || p.Left == "includes" ||
            p.Left == "product" || p.Left == "attribute" || p.Left == "preposition" ||
            p.Left == "prefix" || p.Left == "suffix" {
            direct, inverse := parseTermSlotLists(p.Right)
            cleaned := unison.RemoveVal(inverse, currentName)
            if len(cleaned) != len(inverse) {
                t.Pairs[i].Right = fmtTermSlotLists(direct, cleaned)
                changed = true}}}
    if changed {_ = WriteTerm(t)}}
addToInverse := func(slotName string) {
    direct, _ := parseTermSlotLists(TermPairRight(f, slotName))
    for _, ref := range direct {
        if tp, ok := nameMap[strings.ToLower(ref)]; ok && tp != currentPath {
            target, err2 := LoadTerm(tp)
            if err2 != nil {continue}
            for i, p := range target.Pairs {
                if p.Left == slotName {
                    td, ti := parseTermSlotLists(p.Right)
                    ti = unison.AppendUnique(ti, currentName)
                    target.Pairs[i].Right = fmtTermSlotLists(td, ti); break}}
            _ = WriteTerm(target)}}}
addToInverse("tags")
addToInverse("links")
addToInverse("includes")
addToInverse("product")
addToInverse("attribute")
addToInverse("preposition")
addToInverse("prefix")
addToInverse("suffix")
if err := SyncTermProduct(currentPath, allTerms); err != nil {
    return fmt.Errorf("sync product: %w", err)}
f2, _ := LoadTerm(currentPath)
name2    := TermPairRight(f2, "name")
ln       := len([]rune(name2))
na       := len(strings.Fields(TermPairRight(f2, "alias")))
td2, ti2 := parseTermSlotLists(TermPairRight(f2, "tags"))
ld2, li2 := parseTermSlotLists(TermPairRight(f2, "links"))
id2, ii2 := parseTermSlotLists(TermPairRight(f2, "includes"))
pd2, pi2 := parseTermSlotLists(TermPairRight(f2, "product"))
ad2, ai2 := parseTermSlotLists(TermPairRight(f2, "attribute"))
rd2, ri2 := parseTermSlotLists(TermPairRight(f2, "preposition"))
xd2, xi2 := parseTermSlotLists(TermPairRight(f2, "prefix"))
sd2, si2 := parseTermSlotLists(TermPairRight(f2, "suffix"))
extraSlots := 0
for _, p := range f2.Pairs {
    if len(p.TypeIndex) == 4 {extraSlots++}}
countsVal := fmt.Sprintf("[ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ %d %d ] [ 0 %d ] [ 0 0 ] [ 0 0 ] [ 0 0 ]",
    ln, na,
    len(ti2), len(td2),
    len(li2), len(ld2),
    len(ii2), len(id2),
    len(pi2), len(pd2),
    len(ai2), len(ad2),
    len(ri2), len(rd2),
    len(xi2), len(xd2),
    len(si2), len(sd2),
    extraSlots)
for i, p := range f2.Pairs {
    if p.Left == "counts" {
        f2.Pairs[i].Right = countsVal; break}}
_ = WriteTerm(f2)
return nil }