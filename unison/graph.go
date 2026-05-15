// parus/unison/graph.go

package unison

import (
    "fmt"
    "encoding/json")

//   GraphData — данные графа для фрейма
//	 Left   — west  (входящие по оси запад)
//	 Top    — nord  (входящие по оси север)
//	 Right  — east  (исходящие из фрейма центра по оси запад)
//	 Bottom — south (исходящие из фрейма центра по оси север)
type GraphData struct {
	Center   GraphNode   `json:"center"`
	Left     []GraphNode `json:"left"`
	Top      []GraphNode `json:"top"`
	Right    []GraphNode `json:"right"`
	Bottom   []GraphNode `json:"bottom"`
	Links    []GraphNode `json:"links"`
	Tags     []GraphNode `json:"tags"`
	Includes []GraphNode `json:"includes"`
	Product  []GraphNode `json:"product"`}
// GraphNode — узел графа
type GraphNode struct {
	Index     string      `json:"index"`
	Title     string      `json:"title"`
    Hint      string      `json:"hint,omitempty"`
	Neighbors []GraphNode `json:"neighbors,omitempty"`
	Icon      string      `json:"icon,omitempty"`
	Color     string      `json:"color,omitempty"`
	Rank      string      `json:"rank,omitempty"`}
//  GraphSpec — спецификация графа, читается из мета-фрейма базы
//	West   — имя индексного слота мета-фрейма для входящих по оси запад
//	Nord   — имя индексного слота мета-фрейма для входящих по оси север
//	Center — имя слота фрейма для эквивалентов центрального узла
type GraphSpec struct {
	West       string
	Nord       string
	MatchNord  int
	MatchSouth int}
//  Имена слотов спецификации в мета-фрейме базы
const (
	SpecWest       = "west"
	SpecNord       = "nord"
	SpecMatchNord  = "match-nord"
	SpecMatchSouth = "match-south")
//  loadSpec читает  из мета-фрейма базы
func loadSpec(meta MetaFrame) GraphSpec {
return GraphSpec{
	West:       FramePairRight(meta.Frame, SpecWest),
	Nord:       FramePairRight(meta.Frame, SpecNord),
	MatchNord:  parseIntDefault(FramePairRight(meta.Frame, SpecMatchNord), 0),
	MatchSouth: parseIntDefault(FramePairRight(meta.Frame, SpecMatchSouth), 0)}}
//  GetGraph возвращает GraphData для фрейма в любой базе
//  Спецификация читается из мета-фрейма — west, nord, center
//  West и nord — индексные слоты мета-фрейма (входящие связи)
//  East и south — слоты самого фрейма центра (исходящие связи)
//  Соседи каждого узла берутся из слотов его фрейма
//  Для метабазы nord и south резолвятся как базы, для остальных — как фреймы
func GetGraph(projectID, frameIndex string) (GraphData, error) {
p, err := LoadProject(projectID)
if err != nil {return GraphData{}, fmt.Errorf("load project: %w", err)}
spec     := loadSpec(p.Meta)
titleMap := TitleMap(p.Meta)
isMeta   := projectID == MetaBaseName
// Индексы входящих связей из мета-фрейма
westMap := parseMapSlot(p.Meta.Frame, spec.West, p.Meta.FileOrderLen)
nordMap := parseMapSlot(p.Meta.Frame, spec.Nord, p.Meta.FileOrderLen)
// Собираем нужные фреймы — центр и соседи первого уровня
needed := map[string]bool{frameIndex: true}
for _, idx := range westMap[frameIndex] {needed[idx] = true}
if !isMeta {
	for _, idx := range nordMap[frameIndex] {needed[idx] = true}}
frameMap := make(map[string]Frame)
for idx := range needed {
	if f, err2 := LoadFrame(projectID, idx); err2 == nil {frameMap[idx] = f}}
// Соседи второго уровня — из авторских слотов загруженных фреймов
needed2 := map[string]bool{}
for idx := range frameMap {
	f := frameMap[idx]
	for _, n := range ParseTokenList(FramePairRight(f, spec.West)) {needed2[n] = true}
	for _, n := range ParseTokenList(FramePairRight(f, spec.Nord)) {needed2[n] = true}}
for idx := range needed2 {
	if _, ok := frameMap[idx]; !ok {
		if f, err2 := LoadFrame(projectID, idx); err2 == nil {frameMap[idx] = f}}}
// Центральный узел — эквиваленты из авторского слота spec.Center
cf := frameMap[frameIndex]
// 
appendUniqueNode := func(slice []GraphNode, n GraphNode) []GraphNode {
	for _, v := range slice {if v.Index == n.Index {return slice}}
	return append(slice, n)}
// 
productNodes := make([]GraphNode, 0)
for _, satIdx := range ParseTokenList(FramePairRight(cf, "product")) {
	sat := frameMap[satIdx]
	neighbors := make([]GraphNode, 0)
	for _, key := range ParseTokenList(FramePairRight(cf, "show-product")) {
		nt := titleMap[key]
		if nt == "" {nt = key}
		neighbors = append(neighbors, GraphNode{Index: key, Title: nt})}
	title := titleMap[satIdx]
	if title == "" {title = satIdx}
	productNodes = append(productNodes, GraphNode{
		Index:     satIdx,
		Title:     title,
		Neighbors: neighbors,
		Icon:      FramePairRight(sat, "icon"),
		Color:     FramePairRight(sat, "color"),
		Rank:      FramePairRight(sat, "rank"),})}
    if w := FramePairRight(cf, SpecWest); w != "" {spec.West = w}
    if n := FramePairRight(cf, SpecNord); n != "" {spec.Nord = n}
// Локальные настройки графа из самого фрейма — перекрывают базовые
centerTitle := titleMap[frameIndex]
if centerTitle == "" {centerTitle = frameIndex}
centerNeighbors := make([]GraphNode, 0)
for _, ref := range ParseTokenList(FramePairRight(cf, FrameSlotAlias)) {
	t := titleMap[ref]
	if t == "" {t = ref}
	centerNeighbors = append(centerNeighbors, GraphNode{Index: ref, Title: t})}
center := GraphNode{
	Index:     frameIndex,
	Title:     centerTitle,
	Neighbors: centerNeighbors,
	Icon:      FramePairRight(cf, "icon"),
	Color:     FramePairRight(cf, "color"),
	Rank:      FramePairRight(cf, "rank"),}
// Для обычной базы — все четыре позиции
// West/Nord — из индекса мета-фрейма, соседи из слотов их фреймов
// East/South — из слотов фрейма центра, соседи из слотов их фреймов
linksMap    := parseMapSlot(p.Meta.Frame, "links",    p.Meta.FileOrderLen)
tagsMap     := parseMapSlot(p.Meta.Frame, "tags",     p.Meta.FileOrderLen)
includesMap := parseMapSlot(p.Meta.Frame, "includes", p.Meta.FileOrderLen)
productMap  := parseMapSlot(p.Meta.Frame, "product",  p.Meta.FileOrderLen)
productNodes2 := make([]GraphNode, 0)
for _, fIdx := range productMap[frameIndex] {
	if f, ok := frameMap[fIdx]; ok {
		neighbors := make([]GraphNode, 0)
		for _, satIdx := range ParseTokenList(FramePairRight(f, "product")) {
			if sat, err2 := LoadFrame(projectID, satIdx); err2 == nil {
				var prod map[string]string
				if json.Unmarshal([]byte(FramePairRight(sat, "content")), &prod) == nil {
					for _, key := range ParseTokenList(FramePairRight(f, "show-product")) {
						if key == frameIndex {continue}
						if _, ok2 := prod[key]; ok2 {
							nt := titleMap[key]
							if nt == "" {nt = key}
							neighbors = appendUniqueNode(neighbors, GraphNode{Index: key, Title: nt, Hint: prod[key]})}}}}}
		title := titleMap[fIdx]
		if title == "" {title = fIdx}
		productNodes2 = append(productNodes2, GraphNode{
			Index:     fIdx,
			Title:     title,
			Neighbors: neighbors,
			Icon:      FramePairRight(f, "icon"),
			Color:     FramePairRight(f, "color"),
			Rank:      FramePairRight(f, "rank"),})}}
// resolve — строит GraphNode, соседи из дополнительных слотов neighborSlot фрейма узла
resolve := func(idx, neighborSlot string) GraphNode {
for _, pn := range productNodes2 {
	if pn.Index == idx {return pn}}
title := titleMap[idx]
if title == "" {title = idx}
neighbors := make([]GraphNode, 0)
if f, ok := frameMap[idx]; ok {
isSat := false
for _, satIdx := range ParseTokenList(FramePairRight(cf, "product")) {
	if satIdx == idx {isSat = true; break}}
var keys []string
if isSat {
	var prod map[string]string
	if json.Unmarshal([]byte(FramePairRight(f, "content")), &prod) == nil {
		show := ParseTokenList(FramePairRight(cf, "show-product"))
		for _, k := range show {
			if _, ok := prod[k]; ok {keys = append(keys, k)}}}
} else {keys = ParseTokenList(FramePairRight(f, neighborSlot))}
for _, nIdx := range keys {
	nt := titleMap[nIdx]
	if nt == "" {nt = nIdx}
    neighbors = append(neighbors, GraphNode{Index: nIdx, Title: nt})}}
f := frameMap[idx]
return GraphNode{
	Index:     idx,
	Title:     title,
	Neighbors: neighbors,
	Icon:      FramePairRight(f, "icon"),
	Color:     FramePairRight(f, "color"),
	Rank:      FramePairRight(f, "rank"),}}
// 
resolveList := func(ids []string, neighborSlot string) []GraphNode {
result := make([]GraphNode, 0, len(ids))
for _, id := range ids {result = append(result, resolve(id, neighborSlot))}
return result }
//
for _, idx := range append(append(append(linksMap[frameIndex], tagsMap[frameIndex]...), includesMap[frameIndex]...), productMap[frameIndex]...) {
if _, ok := frameMap[idx]; !ok {
	if f, err2 := LoadFrame(projectID, idx); err2 == nil {frameMap[idx] = f}}}
if !isMeta {
return GraphData{
	Center:   center,
	Left:     resolveList(westMap[frameIndex], spec.West),
	Top:      resolveList(nordMap[frameIndex], spec.Nord),
	Right:    resolveList(ParseTokenList(FramePairRight(cf, spec.West)), spec.West),
	Bottom:   resolveList(ParseTokenList(FramePairRight(cf, spec.Nord)), spec.Nord),
	Tags:     resolveList(tagsMap[frameIndex],     "tags"),
	Links:    resolveList(linksMap[frameIndex],    "links"),
	Includes: resolveList(includesMap[frameIndex], "includes"),
	Product:  productNodes2,}, nil}
// Для метабазы — nord и south резолвятся как базы
basesID      := ParseTokenList(FramePairRight(p.Meta.Frame, "bases-id"))
basesTitles  := ParseTokenList(FramePairRight(p.Meta.Frame, "bases-title"))
baseTitleMap := make(map[string]string)
for i, id := range basesID {
	if i < len(basesTitles) {baseTitleMap[id] = basesTitles[i]}}
// 
resolveBase := func(baseID string) GraphNode {
title := baseTitleMap[baseID]
if title == "" {title = baseID}
return GraphNode{Index: "base:" + baseID, Title: title, Neighbors: []GraphNode{}}}
// 
resolveBaseList := func(ids []string) []GraphNode {
result := make([]GraphNode, 0, len(ids))
for _, id := range ids {result = append(result, resolveBase(id))}
return result }
// South — базы с совпадением по титулу
return GraphData{
	Center:   center,
	Left:     resolveList(westMap[frameIndex],                           spec.West),
    Top:      resolveBaseList(ParseTokenList(FramePairRight(cf, "match-nord"))),
	Right:    resolveList(ParseTokenList(FramePairRight(cf, spec.West)), spec.West),
    Bottom:   resolveBaseList(ParseTokenList(FramePairRight(cf, "match-south"))),
	Tags:     resolveList(tagsMap[frameIndex],     "tags"),
	Links:    resolveList(linksMap[frameIndex],    "links"),
	Includes: resolveList(includesMap[frameIndex], "includes"),
	Product:  resolveList(productMap[frameIndex],  "product"),}, nil }