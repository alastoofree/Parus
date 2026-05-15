// parus/converters/habr.go

package converters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time")

type habrArticle struct {
	ID            string `json:"id"`
	TimePublished string `json:"timePublished"`
	TitleHtml     string `json:"titleHtml"`
	TextHtml      string `json:"textHtml"`
	Hubs          []struct {Title string `json:"title"`} `json:"hubs"`
	Tags          []struct {TitleHtml string `json:"titleHtml"`} `json:"tags"`
	Author        struct {Alias string `json:"alias"`} `json:"author"`}
type habrCommentsResp struct {
	Comments map[string]habrComment `json:"comments"`}
type habrComment struct {
	ID            string `json:"id"`
	ParentID      string `json:"parentId"`
	TimePublished string `json:"timePublished"`
	Message       string `json:"message"`
	Author        struct {Alias string `json:"alias"`} `json:"author"`}

func convertHabr(pageURL string) ([]ConvertPart, error) {
articleID, err := habrExtractID(pageURL)
if err != nil {return nil, err}
article, err := habrFetchArticle(articleID)
if err != nil {return nil, fmt.Errorf("ошибка получения статьи: %w", err)}
comments, _ := habrFetchComments(articleID)
importTag := "habr-" + articleID
authorTag := "habr-author-" + habrSlugify(article.Author.Alias)
tagNames := []string{importTag, authorTag}
for _, hub := range article.Hubs {
	tagNames = append(tagNames, "habr-hub-"+habrSlugify(hub.Title))}
for _, tag := range article.Tags {
	if s := habrSlugify(tag.TitleHtml); s != "" {
		tagNames = append(tagNames, "habr-tag-"+s)}}
var parts []ConvertPart
uniqueTags := make(map[string]bool)
for _, tn := range tagNames {
	if tn != "" && !uniqueTags[tn] {
		parts = append(parts, ConvertPart{Title: tn, IsTag: true})
		uniqueTags[tn] = true}}
mainTitle  := habrStripTags(article.TitleHtml)
sourceLink := fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, pageURL, pageURL)
fullHTML   := article.TextHtml
// Фрагментация по заголовкам H2/H3
re := regexp.MustCompile(`(?i)(<h[23][^>]*>.*?</h[23]>|<p><b>[^<]+</b></p>)`)
headerIndices := re.FindAllStringIndex(fullHTML, -1)
if len(headerIndices) == 0 {
	parts = append(parts, ConvertPart{
		Title:   mainTitle,
		Content: habrNormalize(fullHTML + sourceLink),
		TagName: importTag,
        Rank: "0",
		IsHTML:  true,})
} else {
	// Вступление
	introContent := fullHTML[:headerIndices[0][0]]
	var level1Links strings.Builder
	for i, idx := range headerIndices {
		headerHTML    := fullHTML[idx[0]:idx[1]]
		sectionTitle  := habrStripTags(headerHTML)
		_ = i
		level1Links.WriteString(fmt.Sprintf("[[%s]] ", sectionTitle))}
	mainContent := habrNormalize(introContent + sourceLink)
	if level1Links.Len() > 0 {
		mainContent += "<p><b>Разделы:</b> " + level1Links.String() + "</p>"}
	parts = append(parts, ConvertPart{
		Title:   mainTitle,
		Content: mainContent,
		TagName: importTag,
        Rank: "0",
		IsHTML:  true,})
	// Секции
	for i, idx := range headerIndices {
		start        := idx[1]
		end          := len(fullHTML)
		if i+1 < len(headerIndices) {end = headerIndices[i+1][0]}
		headerHTML   := fullHTML[idx[0]:idx[1]]
		sectionTitle := habrStripTags(headerHTML)
		parts = append(parts, ConvertPart{
			Title:   sectionTitle,
			Content: habrNormalize(fullHTML[start:end]),
			TagName: mainTitle,
			IsHTML:  true,})}}
// Комментарии
//	if comments != nil && len(comments.Comments) > 0 {
//		idToTitle := make(map[string]string)
//		for id := range comments.Comments {
//			idToTitle[id] = fmt.Sprintf("%s-comment-%s", importTag, id)}
//		var sortedIDs []string
//		for id := range comments.Comments {sortedIDs = append(sortedIDs, id)}
//		sort.Strings(sortedIDs)
		// childMap для ссылок
//		childMap := make(map[string][]string)
//		for _, c := range comments.Comments {
//			if c.ParentID != "" && c.ParentID != "0" {
//				childMap[c.ParentID] = append(childMap[c.ParentID], c.ID)}}
//		for _, id := range sortedIDs {
//			c := comments.Comments[id]
//			if c.Message == "" {continue}
//			commentTitle := idToTitle[id]
//			parentTag    := importTag
//			if c.ParentID != "" && c.ParentID != "0" {
//				if pt, ok := idToTitle[c.ParentID]; ok {parentTag = pt}}
//			var b strings.Builder
//			b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", c.Author.Alias))
//			b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", habrFormatDate(c.TimePublished)))
//			b.WriteString("<hr>")
//			b.WriteString(c.Message)
//			if children, ok := childMap[id]; ok {
//				b.WriteString("<p><b>Ответы:</b> ")
//				for _, childID := range children {
//					b.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[childID]))}
//				b.WriteString("</p>")}
//			parts = append(parts, ConvertPart{
//				Title:   commentTitle,
//				Content: habrNormalize(b.String()),
//				TagName: parentTag,
//				Rank:    "0",
//				IsHTML:  true,})}}
//	return parts, nil}
if comments != nil && len(comments.Comments) > 0 {
    idToTitle := make(map[string]string)
    for id := range comments.Comments {
        idToTitle[id] = fmt.Sprintf("%s-comment-%s", importTag, id)}
    var sortedIDs []string
    for id := range comments.Comments {sortedIDs = append(sortedIDs, id)}
    sort.Strings(sortedIDs)
    // childMap для ссылок
    childMap := make(map[string][]string)
    for _, c := range comments.Comments {
        if c.ParentID != "" && c.ParentID != "0" {
            childMap[c.ParentID] = append(childMap[c.ParentID], c.ID)}}
    // Ссылки на комментарии первого уровня — добавляем в последний фрейм поста
    var level1Links strings.Builder
    for _, id := range sortedIDs {
        c := comments.Comments[id]
        if c.ParentID == "" || c.ParentID == "0" {
            level1Links.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[id]))}}
    if level1Links.Len() > 0 {
        for i := range parts {
            if parts[i].Title == mainTitle {
                parts[i].Content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"
                break}}}
    for _, id := range sortedIDs {
        c := comments.Comments[id]
        if c.Message == "" {continue}
        commentTitle := idToTitle[id]
        parentTag    := importTag
        if c.ParentID != "" && c.ParentID != "0" {
            if pt, ok := idToTitle[c.ParentID]; ok {parentTag = pt}}
        var b strings.Builder
        b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", c.Author.Alias))
        b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", habrFormatDate(c.TimePublished)))
        b.WriteString("<hr>")
        b.WriteString(c.Message)
        if children, ok := childMap[id]; ok {
            b.WriteString("<p><b>Ответы:</b> ")
            for _, childID := range children {
                b.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[childID]))}
            b.WriteString("</p>")}
        parts = append(parts, ConvertPart{
            Title:   commentTitle,
            Content: habrNormalize(b.String()),
            TagName: parentTag,
            Rank:    "0",
            IsHTML:  true,})}}
return parts, nil }
// 
func habrExtractID(pageURL string) (string, error) {
re := regexp.MustCompile(`/(?:articles|post|news|blog|companies/[^/]+/articles|ru|en)/(\d+)`)
m  := re.FindStringSubmatch(pageURL)
if len(m) < 2 {return "", fmt.Errorf("не удалось извлечь ID из URL: %s", pageURL)}
return m[1], nil }
// 
func habrFetchArticle(id string) (*habrArticle, error) {
resp, err := http.Get(fmt.Sprintf("https://habr.com/kek/v2/articles/%s?fl=ru&hl=ru", id))
if err != nil {return nil, err}
defer resp.Body.Close()
if resp.StatusCode != 200 {return nil, fmt.Errorf("статус %d", resp.StatusCode)}
var art habrArticle
if err := json.NewDecoder(resp.Body).Decode(&art); err != nil {return nil, err}
return &art, nil }
// 
func habrFetchComments(id string) (*habrCommentsResp, error) {
resp, err := http.Get(fmt.Sprintf("https://habr.com/kek/v2/articles/%s/comments?fl=ru&hl=ru", id))
if err != nil {return nil, err}
defer resp.Body.Close()
var comm habrCommentsResp
if err := json.NewDecoder(resp.Body).Decode(&comm); err != nil {return nil, err}
return &comm, nil }
// 
func habrStripTags(h string) string {
return regexp.MustCompile("<[^>]*>").ReplaceAllString(h, "")}
// 
func habrNormalize(h string) string {
h = regexp.MustCompile(`[\r\n\t]+`).ReplaceAllString(h, " ")
h = regexp.MustCompile(`\s+`).ReplaceAllString(h, " ")
return strings.TrimSpace(h) }
// 
func habrSlugify(s string) string {
s = habrStripTags(s)
s = strings.ToLower(s)
s = strings.ReplaceAll(s, " ", "-")
reg := regexp.MustCompile(`[^a-z0-9а-яё-]`)
s = reg.ReplaceAllString(s, "")
return strings.Trim(s, "-") }
// 
func habrFormatDate(iso string) string {
t, err := time.Parse(time.RFC3339, iso)
if err != nil {return iso}
return t.Format("02.01.2006 15:04") }