// parus/unison/global_search.go — поиск по глобальным тегам

package unison

import (
	"path/filepath"
	"strings")

// SearchByGlobalTag возвращает все фреймы во всех базах с данным глобальным тегом
// Читает g-tags из мета-фрейма мета-базы — список баз для каждого тега.
// В каждой базе ищет фреймы чей титул совпадает с титулом тега или его алиасами
func SearchByGlobalTag(tagIndex string) ([]SearchResult, error) {
metaPath := filepath.Join(DataDir, MetaBaseName, globalMetaFileName())
metaFrame, err := ParseMetaFrame(metaPath)
if err != nil {return nil, err}
// Титул тега и его алиасы
tagTitle := ""
var tagAliases []string
if tags, _ := ListMetaFrames(); tags != nil {
    for _, gt := range tags {
        if gt.Index == tagIndex {
            tagTitle = gt.Title
            tagAliases = gt.Aliases; break}}}
if tagTitle == "" {return nil, nil}
// Базы где есть этот тег — из g-tags
gTagsSlot := ""
for _, p := range metaFrame.Frame.Pairs {
    if p.Left == "g-tags" {gTagsSlot = p.Right; break}}
baseIDs := extractBracketContent(gTagsSlot, tagIndex)
if len(baseIDs) == 0 {return []SearchResult{}, nil}
// baseID → baseTitle
baseTitleMap := make(map[string]string)
for _, p := range metaFrame.Frame.Pairs {
    if p.Left == "bases-id" {
        ids := ParseTokenList(p.Right)
        for _, p2 := range metaFrame.Frame.Pairs {
            if p2.Left == "bases-title" {
                titles := ParseTokenList(p2.Right)
                for i, id := range ids {
                    if i < len(titles) {baseTitleMap[id] = titles[i]}}; break}}; break}}
// Строим набор имён для поиска — титул и алиасы в нижнем регистре
matchNames := make(map[string]bool)
matchNames[strings.ToLower(tagTitle)] = true
for _, a := range tagAliases {matchNames[strings.ToLower(a)] = true}
var results []SearchResult
for _, baseID := range baseIDs {
    proj, err := LoadProject(baseID)
    if err != nil {continue}
    frames, err := ListFrames(baseID, proj.Meta)
    if err != nil {continue}
    for _, f := range frames {
        title := FramePairRight(f, "title")
        if matchNames[strings.ToLower(title)] {
            results = append(results, SearchResult{
                BaseID:     baseID,
                BaseTitle:  baseTitleMap[baseID],
                FrameIndex: f.Index,
                FrameTitle: title,
                TagIndex:   tagIndex,
                TagTitle:   tagTitle})}}}
return results, nil }
// SearchGlobal — текстовый поиск по title и aliases
func SearchGlobal(query string) ([]GlobalFrame, error) {
if query == "" { return ListMetaFrames() }
q := strings.ToLower(query)
all, err := ListMetaFrames()
if err != nil { return nil, err }
var result []GlobalFrame
for _, gt := range all {
	if strings.Contains(strings.ToLower(gt.Title), q) {
		result = append(result, gt); continue}
	for _, a := range gt.Aliases {
		if strings.Contains(strings.ToLower(a), q) {
			result = append(result, gt); break}}}
return result, nil }
// SearchInFrame ищет query в контенте фрейма и всех его алиасов
func SearchInFrame(projectID, frameIndex, query string) []map[string]string {
if query == "" {return nil}
query = strings.ToLower(query)
var results []map[string]string
indices := []string{frameIndex}
if f, err := LoadFrame(projectID, frameIndex); err == nil {
	for _, p := range f.Pairs {
		if p.Left == "alias" {
			indices = append(indices, strings.Fields(p.Right)...); break}}}
p2, _ := LoadProject(projectID)
titleMap := TitleMap(p2.Meta)
for _, idx := range indices {
	f, err := LoadFrame(projectID, idx)
	if err != nil {continue}
	content := FramePairRight(f, "content")
	lower := strings.ToLower(content)
	pos := strings.Index(lower, query)
	if pos < 0 {continue}
	start := pos - 80
	if start < 0 {start = 0}
	end := pos + len(query) + 80
	if end > len(content) {end = len(content)}
	excerpt := "..." + content[start:end] + "..."
	title := titleMap[idx]
	if title == "" {title = idx}
	results = append(results, map[string]string{
		"index":   idx,
		"title":   title,
		"excerpt": excerpt})}
return results }
// SearchInProject ищет query по титулам и контентам всех фреймов базы
func SearchInProject(projectID, query string) ([]map[string]string, error) {
if query == "" {return nil, nil}
p, err := LoadProject(projectID)
if err != nil {return nil, err}
frames, err := ListFrames(projectID, p.Meta)
if err != nil {return nil, err}
q := strings.ToLower(query)
var results []map[string]string
titleMap := TitleMap(p.Meta)
for _, f := range frames {
    ct := FramePairRight(f, "content-type")
    if ct == "svg" || ct == "json" {continue}
    title    := FramePairRight(f, "title")
    content  := FramePairRight(f, "content")
    stripped := stripContent(content)
    titleL   := strings.ToLower(title)
    strippedL := strings.ToLower(stripped)
    pos := strings.Index(strippedL, q)
    if !strings.Contains(titleL, q) && pos < 0 {continue}
    excerpt := ""
    if pos >= 0 {
        start := pos - 80
        if start < 0 {start = 0}
        end := pos + len(q) + 80
        if end > len(stripped) {end = len(stripped)}
        excerpt = "..." + stripped[start:end] + "..."}
    t := titleMap[f.Index]
    if t == "" {t = title}
    results = append(results, map[string]string{
        "index":   f.Index,
        "title":   t,
        "excerpt": excerpt})}
return results, nil}
// SearchAllProjects ищет query по всем базам
func SearchAllProjects(query string) ([]SearchResult, error) {
if query == "" {return nil, nil}
projects, err := ListProjects()
if err != nil {return nil, err}
q := strings.ToLower(query)
var results []SearchResult
for _, p := range projects {
    if p.ID == MetaBaseName {continue}
    frames, err := ListFrames(p.ID, p.Meta)
    if err != nil {continue}
    baseTitle := projectTitle(p)
    for _, f := range frames {
        title    := FramePairRight(f, "title")
        content  := FramePairRight(f, "content")
        titleL   := strings.ToLower(title)
        contentL := strings.ToLower(content)
        pos := strings.Index(contentL, q)
        if !strings.Contains(titleL, q) && pos < 0 {continue}
        excerpt := ""
        if pos >= 0 {
            start := pos - 80
            if start < 0 {start = 0}
            end := pos + len(q) + 80
            if end > len(content) {end = len(content)}
            excerpt = "..." + content[start:end] + "..."}
        results = append(results, SearchResult{
            BaseID:     p.ID,
            BaseTitle:  baseTitle,
            FrameIndex: f.Index,
            FrameTitle: title,
            TagIndex:   "",
            TagTitle:   excerpt})}}
return results, nil}