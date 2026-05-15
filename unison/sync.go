// parus/unison/sync.go

package unison

import (
	"fmt"
    "sync"
	"strings"
    "encoding/json"
	"regexp")

var metaMu sync.Mutex
// EnsureSlots добавляет отсутствующие слоты во фрейм
func EnsureSlots(frame Frame, slots []string) Frame {
for _, left := range slots {
	found := false
	for _, p := range frame.Pairs {
		if p.Left == left { found = true; break }}
	if !found {
		letter := slotLetter[left]
		if letter == "" {letter = "x"}
		frame.Pairs = append(frame.Pairs, SlotPair{
			TypeIndex: letter,
			Left:      left,
			Right:     "",})}}
return frame }
// ResolveAlias находит фрейм по title или alias
func ResolveAlias(query string) (GlobalFrame, bool) {
result, _ := ListMetaFrames()
q := strings.ToLower(query)
for _, gf := range result {
	if strings.ToLower(gf.Title) == q {return gf, true}
	for _, a := range gf.Aliases {
		if strings.ToLower(a) == q {return gf, true}}}
return GlobalFrame{}, false }
// ResolveContent возвращает контент фрейма с резолвенными трансклюзиями
func ResolveContent(projectID, frameIndex string) (string, string, error) {
frame, err := LoadFrame(projectID, frameIndex)
if err != nil {
	return "", "", fmt.Errorf("load frame: %w", err)}
content     := FramePairRight(frame, "content")
contentType := FramePairRight(frame, "content-type")
if contentType == "" || contentType == "text" {
contentType = "markdown"}
// Строим карту индекс->титул
p2, err2 := LoadProject(projectID)
if err2 != nil {
	return content, contentType, nil}
titleMap := TitleMap(p2.Meta)
if contentType == "json" {
    for idx, title := range titleMap {
        content = strings.ReplaceAll(content, `"`+idx+`"`, `"`+title+`"`)}}
result := regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllStringFunc(content, func(match string) string {
	idx := strings.TrimSpace(match[2 : len(match)-2])
	title := titleMap[idx]
	if title == "" { title = idx }
	return `<a href="frame.html?id=` + projectID + `&frame=` + idx + `">` + title + `</a>`})
// Резолвим {{idx}} — трансклюзии
result = regexp.MustCompile(`\{\{([^}]+)\}\}`).ReplaceAllStringFunc(result, func(match string) string {
	idx := strings.TrimSpace(match[2 : len(match)-2])
    resolved, _, _ := ResolveContent(projectID, idx)
    return resolved})
return result, contentType, nil }
// SyncContentLinks парсит [[idx]] и {{idx}} из слота content
// и добавляет найденные индексы в links и includes
func SyncContentLinks(pairs []SlotPair) []SlotPair {
var content, contentType string
for _, p := range pairs {
    if p.Left == "content"      { content = p.Right }
    if p.Left == "content-type" { contentType = p.Right }}
if contentType == "json" {return pairs}
if content == "" {
    // Контент пустой — очищаем links и includes
    for i, p := range pairs {
        if p.Left == "links"    { pairs[i].Right = "" }
        if p.Left == "includes" { pairs[i].Right = "" }}
    return pairs}
// Собираем [[idx]] → links
linksSet := make(map[string]bool)
for _, m := range regexp.MustCompile(`\[\[([^\]]+)\]\]`).FindAllStringSubmatch(content, -1) {
	linksSet[strings.TrimSpace(m[1])] = true}
// Собираем {{idx}} → includes
includesSet := make(map[string]bool)
for _, m := range regexp.MustCompile(`\{\{([^}]+)\}\}`).FindAllStringSubmatch(content, -1) {
	includesSet[strings.TrimSpace(m[1])] = true}
// Обновляем слоты
for i, p := range pairs {
	if p.Left == "links" {
		var list []string
		for k := range linksSet { list = append(list, k) }
		pairs[i].Right = strings.Join(list, " ")}
	if p.Left == "includes" {
		var list []string
		for k := range includesSet { list = append(list, k) }
		pairs[i].Right = strings.Join(list, " ")}}
return pairs }
// SyncTagRemoval — при трансформации тег→фрейм чистит индекс
// из слота tags у всех фреймов проекта
func SyncTagRemoval(projectID, tagIndex string) error {
p, err := LoadProject(projectID)
if err != nil {
    return fmt.Errorf("load project: %w", err)}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {
    return fmt.Errorf("list frames: %w", err)}
for _, f := range frames {
    for i, pair := range f.Pairs {
        if pair.Left == "tags" {
            tokens := ParseTokenList(pair.Right)
            var clean []string
            for _, t := range tokens {
                if t != tagIndex {
                    clean = append(clean, t)}}
            if len(clean) != len(tokens) {
                f.Pairs[i].Right = strings.Join(clean, " ")
                if err := WriteFrame(projectID, f); err != nil {
                    return fmt.Errorf("sync tag removal %s: %w", f.Index, err)}}; break}}}
return nil }
// SyncAliasGroup синхронизирует алиас-группу фрейма
// newAliases — новый список алиасов из слота alias после редактирования
func SyncAliasGroup(projectID, frameIndex string, newAliases []string) error {
// Собираем всю группу — текущий фрейм + его алиасы
group := AppendUnique(newAliases, frameIndex)
// Для каждого члена группы записываем всех остальных
for _, idx := range group {
	f, err := LoadFrame(projectID, idx)
	if err != nil {continue}
	for i, p := range f.Pairs {
		if p.Left == "alias" {
			var members []string
			for _, m := range group {
				if m != idx {members = append(members, m)}}
			f.Pairs[i].Right = strings.Join(members, " "); break}}
	if err := WriteFrame(projectID, f); err != nil {
		return fmt.Errorf("sync alias group %s: %w", idx, err)}}
return nil }
// 
func ResolveProduct(projectID string, frame Frame) string {
satIdxs := ParseTokenList(FramePairRight(frame, "product"))
if len(satIdxs) == 0 {return ""}
showList := FramePairRight(frame, "show-product")
if showList == "" {return ""}
p2, _ := LoadProject(projectID)
titleMap := TitleMap(p2.Meta)
var sb strings.Builder
for _, satIdx := range satIdxs {
	sat, err := LoadFrame(projectID, satIdx)
	if err != nil {continue}
	satTitle := titleMap[satIdx]
	if satTitle == "" {satTitle = satIdx}
	var prod map[string]string
	if json.Unmarshal([]byte(FramePairRight(sat, "content")), &prod) != nil {continue}
	satLink := `<a href="frame.html?id=` + projectID + `&frame=` + satIdx + `">` + satTitle + `</a>`
    sb.WriteString("<strong>" + satLink + "</strong> [ ")
    first := true
    for _, key := range ParseTokenList(showList) {
        if val, ok := prod[key]; ok {
            if !first {sb.WriteString(" · ")}
        first = false
        title := titleMap[key]
        if title == "" {title = key}
        safeVal := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(val, `$1`)
        safeVal = strings.ReplaceAll(safeVal, `"`, `&quot;`)
        link := `<a href="frame.html?id=` + projectID + `&frame=` + key + `" title="` + safeVal + `">` + title + `</a>`
        sb.WriteString(link)}}
    sb.WriteString(" ] ")}
return sb.String() }