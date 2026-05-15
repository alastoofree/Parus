// parus/unison/terms/term-graph.go

package terms

import (
	"fmt"
	"sort"
	"strings"
	"parus/unison")

// GetTermGraph строит граф для термина
func GetTermGraph(path string) (map[string]interface{}, error) {
f, err := LoadTerm(path)
if err != nil {return nil, fmt.Errorf("load term: %w", err)}
allTerms, err := ListTerms()
if err != nil {return nil, fmt.Errorf("list terms: %w", err)}
nameMap := make(map[string]string)
for _, t := range allTerms {
    name := TermPairRight(t, "name")
    if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
// 
resolve := func(names []string, slotName string) []map[string]interface{} {
    result := []map[string]interface{}{}
    for _, ref := range names {
        if p, ok := nameMap[strings.ToLower(ref)]; ok {
            neighbor, err2 := LoadTerm(p)
            if err2 != nil {continue}
            nd, _ := parseTermSlotLists(TermPairRight(neighbor, slotName))
            if slotName == "" {nd = nil}
            var neighbors []map[string]interface{}
            for _, nref := range nd {
                if np, ok2 := nameMap[strings.ToLower(nref)]; ok2 {
                    neighbors = append(neighbors, map[string]interface{}{"index": np, "title": nref})}}
            if neighbors == nil {neighbors = []map[string]interface{}{}}
            result = append(result, map[string]interface{}{
                "index":     p,
                "title":     ref,
                "icon":      p,
                "color":     TermPairRight(neighbor, "color"),
                "neighbors": neighbors,})}}
    return result }
name  := TermPairRight(f, "name")
td, ti := parseTermSlotLists(TermPairRight(f, "tags"))
ld, li := parseTermSlotLists(TermPairRight(f, "links"))
id, ii := parseTermSlotLists(TermPairRight(f, "includes"))
pd, pi := parseTermSlotLists(TermPairRight(f, "product"))
ad, ai := parseTermSlotLists(TermPairRight(f, "attribute"))
rd, ri := parseTermSlotLists(TermPairRight(f, "preposition"))
xd, xi := parseTermSlotLists(TermPairRight(f, "prefix"))
sd, si := parseTermSlotLists(TermPairRight(f, "suffix"))
westSlot := TermPairRight(f, "west")
nordSlot := TermPairRight(f, "nord")
if westSlot == "" {westSlot = "tags"}
if nordSlot == "" {nordSlot = "links"}
slotDirect := map[string][]string{
    "tags": td, "links": ld, "includes": id, "product": pd,
    "attribute": ad, "preposition": rd, "prefix": xd, "suffix": sd}
slotInverse := map[string][]string{
    "tags": ti, "links": li, "includes": ii, "product": pi,
    "attribute": ai, "preposition": ri, "prefix": xi, "suffix": si}
center := map[string]interface{}{
    "index":     path,
    "title":     name,
    "icon":      path,
    "color":     TermPairRight(f, "color"),
    "neighbors": resolve(unison.ParseTokenList(TermPairRight(f, "alias")), ""),}
return map[string]interface{}{
    "center":           center,
    "left":             resolve(slotInverse[westSlot], westSlot),
    "top":              resolve(slotInverse[nordSlot], nordSlot),
    "right":            resolve(slotDirect[westSlot],  westSlot),
    "bottom":           resolve(slotDirect[nordSlot],  nordSlot),
    "tags":             resolve(td, "tags"),
    "links":            resolve(ld, "links"),
    "includes":         resolve(id, "includes"),
    "product":          resolve(pd, "product"),
    "attribute":        resolve(ad, "attribute"),
    "preposition":      resolve(rd, "preposition"),
    "prefix":           resolve(xd, "prefix"),
    "suffix":           resolve(sd, "suffix"),
    "interTags":        resolve(ti, "tags"),
    "interLinks":       resolve(li, "links"),
    "interIncludes":    resolve(ii, "includes"),
    "interProduct":     resolve(pi, "product"),
    "interAttribute":   resolve(ai, "attribute"),
    "interPreposition": resolve(ri, "preposition"),
    "interPrefix":      resolve(xi, "prefix"),
    "interSuffix":      resolve(si, "suffix"),}, nil }
// GetTermChildren возвращает дочерние термы тега
func GetTermChildren(path, iSort, aSort string) ([]map[string]string, error) {
f, err := LoadTerm(path)
if err != nil {return nil, fmt.Errorf("load term: %w", err)}
allTerms, err := ListTerms()
if err != nil {return nil, fmt.Errorf("list terms: %w", err)}
nameMap := make(map[string]string)
for _, t := range allTerms {
	name := TermPairRight(t, "name")
	if name != "" {nameMap[strings.ToLower(name)] = t.Path}}
children := []map[string]string{}
_, ti := parseTermSlotLists(TermPairRight(f, "tags"))
for _, ref := range ti {
	if p, ok := nameMap[strings.ToLower(ref)]; ok {
		children = append(children, map[string]string{"index": p, "name": ref})}}
if iSort != "" {
	order := map[string]int{}
	for i, name := range unison.ParseTokenList(iSort) {order[strings.ToLower(name)] = i}
	sort.Slice(children, func(i, j int) bool {
		oi, ok1 := order[strings.ToLower(children[i]["name"])]
		oj, ok2 := order[strings.ToLower(children[j]["name"])]
		if !ok1 {oi = 9999}
		if !ok2 {oj = 9999}
		return oi < oj})
} else if aSort != "" {
	charOrder := unison.BuildCharOrder(aSort)
	sort.Slice(children, func(i, j int) bool {
		return unison.CompareByASort(children[i]["name"], children[j]["name"], charOrder) < 0})}
return children, nil }