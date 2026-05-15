// parus/converters/convert.go

package converters

import (
	"fmt"
	"strings"
	"parus/unison")

// ConvertPart — один фрейм результата конвертации
type ConvertPart struct {
	Title   string
	Content string
	IsTag   bool
	TagName string
    Rank string
	IsHTML  bool}
// Convert определяет платформу и запускает нужный парсер
func Convert(pageURL string) ([]ConvertPart, error) {
u := strings.ToLower(pageURL)
switch {
	case strings.Contains(u, "wikipedia.org") ||
		strings.Contains(u, "wikibooks.org") ||
		strings.Contains(u, "wikisource.org"):
		return convertWikipedia(pageURL)
	case strings.Contains(u, "livejournal.com"):
		return convertLJ(pageURL)
	case strings.Contains(u, "blogspot.com") ||
		strings.Contains(u, "blogger.com"):
		return convertBlogger(pageURL)
	case strings.Contains(u, "medium.com"):
		return convertMedium(pageURL)
	case strings.Contains(u, "habr.com"):
		return convertHabr(pageURL)
	case strings.Contains(u, "wordpress.com"):
		return convertWordpress(pageURL)
	case strings.Contains(u, "hashnode.dev") ||
		strings.Contains(u, "hashnode.com") ||
		strings.Contains(u, "hashnode.space"):
		return convertHashnode(pageURL)
	default:
		// Пробуем Blogger API для кастомных доменов
		if parts, err := convertBlogger(pageURL); err == nil {
			return parts, nil}
return nil, fmt.Errorf("unsupported platform: %s", pageURL)} }
// InjectParts создаёт фреймы из результатов конвертации
func InjectParts(projectID string, meta unison.MetaFrame, parts []ConvertPart) (int, error) {
tagMap   := make(map[string]string) // tagName → index
titleMap := make(map[string]string) // title → index
// Проход 1 — теги
for _, part := range parts {
	if !part.IsTag {continue}
	f, err := unison.CreateFrame(projectID, meta, part.Title)
	if err != nil {continue}
	for i, p := range f.Pairs {
		if p.Left == "rank" {f.Pairs[i].Right = "0"; break}}
	_ = unison.WriteFrame(projectID, f)
    _ = unison.Count(projectID, 0, 1)
	tagMap[part.Title]   = f.Index
	titleMap[part.Title] = f.Index}
// Проход 2 — фреймы с контентом
type created struct{ index string; part ConvertPart }
var createdList []created
for _, part := range parts {
	if part.IsTag {continue}
	f, err := unison.CreateFrame(projectID, meta, part.Title)
	if err != nil {continue}
	for i, p := range f.Pairs {
		if p.Left == "content-type" {
            if part.IsHTML {
                f.Pairs[i].Right = "html"
            } else {
                f.Pairs[i].Right = "markdown"}}
		if p.Left == "tags" {
            if idx, ok := tagMap[part.TagName]; ok {
                f.Pairs[i].Right = idx
        } else if idx, ok := titleMap[part.TagName]; ok {
            f.Pairs[i].Right = idx}}
        if p.Left == "rank" && part.Rank != "" {
            f.Pairs[i].Right = part.Rank}}
	_ = unison.WriteFrame(projectID, f)
	titleMap[part.Title] = f.Index
	createdList = append(createdList, created{f.Index, part})}
// Проход 2б — обновляем tags для фреймов с родителями-фреймами
for _, c := range createdList {
    if c.part.TagName == "" {continue}
    if _, ok := tagMap[c.part.TagName]; ok {continue} // уже обработан
    if idx, ok := titleMap[c.part.TagName]; ok {
        f, err := unison.LoadFrame(projectID, c.index)
        if err != nil {continue}
        for i, p := range f.Pairs {
            if p.Left == "tags" {f.Pairs[i].Right = idx; break}}
    _ = unison.WriteFrame(projectID, f)}}
// Проход 3 — подставляем реальные индексы в трансклюзии
var count int
for _, c := range createdList {
	content := c.part.Content
	for title, idx := range titleMap {
		content = strings.ReplaceAll(content, "{{"+title+"}}", "{{"+idx+"}}")}
        for title, idx := range titleMap {
            content = strings.ReplaceAll(content, "[["+title+"]]",
                fmt.Sprintf(`<a href="frame.html?id=%s&frame=%s">%s</a>`, projectID, idx, title))}
notesIdx := ""
for title, idx := range titleMap {
    if strings.HasSuffix(title, ": Примечания") ||
    strings.HasSuffix(title, ": References") ||
    strings.HasSuffix(title, ": Сноски") ||
    strings.HasSuffix(title, ": Notes") {
    notesIdx = idx; break}}
noteLink := "<sup>*</sup>"
if notesIdx != "" {
    noteLink = fmt.Sprintf(` <sup><a href="frame.html?id=%s&frame=%s">*</a></sup>`, projectID, notesIdx)}
content = strings.ReplaceAll(content, "__NOTES__", noteLink)
	f, err := unison.LoadFrame(projectID, c.index)
	if err != nil {continue}
	for i, p := range f.Pairs {
		if p.Left == "content" {f.Pairs[i].Right = content; break}}
	_ = unison.WriteFrame(projectID, f)
	_ = unison.BuildIndex(projectID, f.Index)
	count++}
return count, nil }
// 
// Slugify преобразует строку в slug: нижний регистр, пробелы и точки → дефис
func Slugify(s string) string {
s = strings.ToLower(s)
s = strings.ReplaceAll(s, " ", "-")
s = strings.ReplaceAll(s, ".", "-")
return s }