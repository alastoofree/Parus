// parus/converters/medium.go

package converters

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time")

type mediumRef struct {
	Ref string `json:"__ref"`}
type mediumParagraph struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	Text            string            `json:"text"`
	Markups         []mediumMarkup    `json:"markups"`
	Layout          string            `json:"layout"`
	Metadata        *mediumRefWrapper `json:"metadata"`
	Iframe          *mediumRefWrapper `json:"iframe"`
	MixtapeMetadata *mediumRefWrapper `json:"mixtapeMetadata"`}
type mediumRefWrapper struct {
	Ref           string     `json:"__ref"`
	MediaResource *mediumRef `json:"mediaResource"`}
type mediumMarkup struct {
	Type  string `json:"type"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Href  string `json:"href"`}
type mediumImageResource struct {
	ID string `json:"id"`}
type mediumMediaResource struct {
	IframeSrc string `json:"iframeSrc"`}
type mediumMixtapeResource struct {
	Href string `json:"href"`}
type mediumTagResource struct {
	DisplayTitle string `json:"displayTitle"`
	Slug         string `json:"normalizedTagSlug"`}
type mediumGQLResponse struct {
	Data struct {
		Post struct {
			Responses struct {
				Edges []struct {
					Node struct {
						ID      string `json:"id"`
						Creator struct {Name string `json:"name"`} `json:"creator"`
						Content struct {
							BodyModel struct {
								Paragraphs []mediumParagraph `json:"paragraphs"`
							} `json:"bodyModel"`
						} `json:"content"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"responses"`
		} `json:"post"`
	} `json:"data"`}
// 
func convertMedium(pageURL string) ([]ConvertPart, error) {
tr := &http.Transport{
	TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),}
client := &http.Client{Transport: tr, Timeout: 30 * time.Second}
req, _ := http.NewRequest("GET", pageURL, nil)
req.Header.Set("User-Agent", "Twitterbot/1.0")
resp, err := client.Do(req)
if err != nil {return nil, err}
defer resp.Body.Close()
htmlBytes, _ := io.ReadAll(resp.Body)
htmlStr := string(htmlBytes)
// Извлекаем APOLLO_STATE
//	reState := regexp.MustCompile(`window\.__APOLLO_STATE__\s*=\s*({.+?})</script>`)
reState := regexp.MustCompile(`window\.__APOLLO_STATE__\s*=\s*({.+})</script>`)
matches := reState.FindStringSubmatch(htmlStr)
if len(matches) < 2 {
	return nil, fmt.Errorf("не найден блок данных APOLLO_STATE")}
var rawData map[string]json.RawMessage
if err := json.Unmarshal([]byte(matches[1]), &rawData); err != nil {
	return nil, fmt.Errorf("ошибка разбора JSON: %v", err)}
getObject := func(ref string, target interface{}) bool {
	if raw, ok := rawData[ref]; ok {
		json.Unmarshal(raw, target)
		return true}
	return false}
// Ищем пост
var paragraphs []mediumParagraph
var postTitle string
var postID   string
var postTags []string
foundPost := false
for key, val := range rawData {
	if strings.HasPrefix(key, "Post:") && strings.Contains(string(val), "bodyModel") {
		foundPost = true
		postID = strings.TrimPrefix(key, "Post:")
		var postMap map[string]json.RawMessage
		json.Unmarshal(val, &postMap)
		if t, ok := postMap["title"]; ok {json.Unmarshal(t, &postTitle)}
		if tagsJSON, ok := postMap["tags"]; ok {
			var tagsRefs []mediumRef
			json.Unmarshal(tagsJSON, &tagsRefs)
			for _, tRef := range tagsRefs {
				var tRes mediumTagResource
				if getObject(tRef.Ref, &tRes) {postTags = append(postTags, tRes.DisplayTitle)}}}
		var contentBody struct {
			BodyModel struct {
				Paragraphs []mediumRef `json:"paragraphs"`
			} `json:"bodyModel"`}
		for k, v := range postMap {
			if strings.HasPrefix(k, "content") {
				json.Unmarshal(v, &contentBody); break}}
		for _, pRef := range contentBody.BodyModel.Paragraphs {
			var para mediumParagraph
			if getObject(pRef.Ref, &para) {
				if para.Metadata != nil && para.Metadata.Ref != "" {
					var img mediumImageResource
					if getObject(para.Metadata.Ref, &img) {para.Layout = img.ID}}
				if para.Iframe != nil && para.Iframe.MediaResource != nil && para.Iframe.MediaResource.Ref != "" {
					var vid mediumMediaResource
					if getObject(para.Iframe.MediaResource.Ref, &vid) {
						if para.Type == "IFRAME" {para.Text = vid.IframeSrc}}}
				if para.MixtapeMetadata != nil && para.MixtapeMetadata.Ref != "" {
					var mix mediumMixtapeResource
					if getObject(para.MixtapeMetadata.Ref, &mix) {para.Layout = mix.Href}}
				paragraphs = append(paragraphs, para)}}; break}}
if !foundPost {return nil, fmt.Errorf("структура Post не найдена в JSON")}
if len(paragraphs) == 0 {return nil, fmt.Errorf("список параграфов пуст")}
log.Printf("Medium: пост '%s', параграфов: %d", postTitle, len(paragraphs))
// Теги
tagsList := []string{"medium-" + mediumSlugify(postTitle), "medium-пост", "medium-import"}
for _, t := range postTags {
	tagsList = append(tagsList, "medium-"+mediumSlugify(t))}
var parts []ConvertPart
for _, t := range tagsList {
	parts = append(parts, ConvertPart{Title: t, IsTag: true})}
baseTag    := "medium-" + mediumSlugify(postTitle)
sourceLink := fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, pageURL, pageURL)
// Фрагментация по H3
var mainBuf     strings.Builder
var sectionBuf  strings.Builder
var sectionTitle string
inSection := false
var sectionTitles []string
if len(postTags) > 0 {
	mainBuf.WriteString("<p><b>Теги:</b> " + strings.Join(postTags, ", ") + "</p><hr>")}
for i, para := range paragraphs {
	if i == 0 {continue}
	if para.Type == "H3" {
		if inSection {
			parts = append(parts, ConvertPart{
				Title:   sectionTitle,
				Content: strings.ReplaceAll(strings.TrimSpace(sectionBuf.String()), "\n", " "),
				TagName: baseTag,
				IsHTML:  true,})
			mainBuf.WriteString(fmt.Sprintf("<h3>[[%s]]</h3>\n", sectionTitle))
			sectionBuf.Reset()
		} else {
			inSection = true}
		sectionTitle = para.Text
		sectionTitles = append(sectionTitles, sectionTitle)
		continue}
	chunk := mediumParaToHTML(para)
	if inSection {
		sectionBuf.WriteString(chunk)
	} else {
		mainBuf.WriteString(chunk)}}
if inSection && sectionBuf.Len() > 0 {
	parts = append(parts, ConvertPart{
		Title:   sectionTitle,
		Content: strings.ReplaceAll(strings.TrimSpace(sectionBuf.String()), "\n", " "),
		TagName: baseTag,
		IsHTML:  true,})
	mainBuf.WriteString(fmt.Sprintf("<h3>[[%s]]</h3>\n", sectionTitle))}
mainBuf.WriteString(sourceLink)
parts = append(parts, ConvertPart{
	Title:   postTitle,
	Content: strings.ReplaceAll(strings.TrimSpace(mainBuf.String()), "\n", " "),
	TagName: baseTag,
	IsHTML:  true,})
log.Printf("Medium: создано %d фреймов (включая %d секций)", len(parts), len(sectionTitles))
// Комментарии
if postID != "" {
	comments := mediumFetchComments(postID, client, baseTag)
	parts = append(parts, comments...)}
return parts, nil }
// mediumFetchComments загружает комментарии через GraphQL
func mediumFetchComments(postID string, client *http.Client, baseTag string) []ConvertPart {
query := `{"query":"query PostResponses($postId: String!, $first: Int) { post(id: $postId) { responses(first: $first) { edges { node { id creator { name } content { bodyModel { paragraphs { id type text markups { type start end href } } } } } } } } }","variables":{"postId":"` + postID + `","first":25}}`
req, _ := http.NewRequest("POST", "https://medium.com/_/graphql", strings.NewReader(query))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("User-Agent", "Twitterbot/1.0")
resp, err := client.Do(req)
if err != nil || resp.StatusCode != 200 {return nil}
defer resp.Body.Close()
var gqlResp mediumGQLResponse
if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {return nil}
var parts []ConvertPart
for _, edge := range gqlResp.Data.Post.Responses.Edges {
	node := edge.Node
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", node.Creator.Name))
	b.WriteString("<hr>")
	for _, para := range node.Content.BodyModel.Paragraphs {
		b.WriteString(mediumParaToHTML(para))}
	commentTitle := fmt.Sprintf("%s-comment-%s", baseTag, node.ID)
	parts = append(parts, ConvertPart{
		Title:   commentTitle,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: baseTag,
		Rank:    "0",
		IsHTML:  true,})}
log.Printf("Medium: комментариев: %d", len(parts))
return parts }
// mediumParaToHTML конвертирует параграф в HTML
func mediumParaToHTML(para mediumParagraph) string {
textRunes  := []rune(para.Text)
insertions := make(map[int]string)
for _, m := range para.Markups {
	if m.Start < 0 || m.End > len(textRunes) {continue}
	var open, close string
	switch m.Type {
	case "BOLD", "STRONG":
		open, close = "<b>", "</b>"
	case "ITALIC", "EM":
		open, close = "<i>", "</i>"
	case "A", "LINK":
		open  = fmt.Sprintf(`<a href="%s" target="_blank">`, m.Href)
		close = "</a>"
	default:
		continue}
	insertions[m.Start] = insertions[m.Start] + open
	insertions[m.End]   = close + insertions[m.End]}
var sb strings.Builder
for i := 0; i < len(textRunes); i++ {
	if tag, ok := insertions[i]; ok {sb.WriteString(tag)}
	sb.WriteRune(textRunes[i])}
if tag, ok := insertions[len(textRunes)]; ok {sb.WriteString(tag)}
textHTML := sb.String()
switch para.Type {
case "H4":
	return fmt.Sprintf("<h3>%s</h3>\n", textHTML)
case "P":
	return fmt.Sprintf("<p>%s</p>\n", textHTML)
case "IMG":
	if para.Layout != "" {
		return fmt.Sprintf(`<figure><img src="https://miro.medium.com/max/1400/%s" style="max-width:100%%"></figure>`+"\n", para.Layout)}
case "IFRAME":
	if para.Text != "" {
		return fmt.Sprintf(`<iframe src="%s" frameborder="0" allowfullscreen style="width:100%%;aspect-ratio:16/9"></iframe>`+"\n", para.Text)}
case "MIXTAPE_EMBED":
	if para.Layout != "" {
		return fmt.Sprintf(`<blockquote><a href="%s" target="_blank"><b>%s</b></a></blockquote>`+"\n", para.Layout, textHTML)}
case "BQ":
	return fmt.Sprintf("<blockquote>%s</blockquote>\n", textHTML)
case "CODE", "PRE": 
    return fmt.Sprintf("<pre><code>%s</code></pre>\n", textHTML)
case "ULI", "OLI":
	return fmt.Sprintf("<li>%s</li>\n", textHTML)}
if textHTML != "" {
	return fmt.Sprintf("<p>%s</p>\n", textHTML)}
return ""}
// 
func mediumSlugify(s string) string {
s = strings.ToLower(s)
reg, _ := regexp.Compile("[^a-z0-9]+")
s = reg.ReplaceAllString(s, "-")
return strings.Trim(s, "-") }