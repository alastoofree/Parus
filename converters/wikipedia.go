// parus/converters/wikipedia.go

package converters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html")

type wikiInfo struct {
	Domain      string
	ProjectName string}

func convertWikipedia(pageURL string) ([]ConvertPart, error) {
info, err := wikiInfoFromURL(pageURL)
if err != nil {return nil, err}
articleTitle, err := wikiTitleFromURL(pageURL)
if err != nil {return nil, err}
htmlContent, err := wikiGetHTML(articleTitle, info.Domain)
if err != nil {return nil, err}
doc, err := html.Parse(strings.NewReader(htmlContent))
if err != nil {return nil, fmt.Errorf("parse html: %w", err)}
pageTitle := wikiExtractTitle(htmlContent, articleTitle)
tagName   := "wikipedia: " + pageTitle
var parts  []ConvertPart
parts = append(parts, ConvertPart{Title: tagName, IsTag: true})
// 1. НАВБОКСЫ
var allNavboxes []*html.Node
var findNavboxes func(*html.Node)
findNavboxes = func(n *html.Node) {
	if wikiHasClass(n, "navbox") {allNavboxes = append(allNavboxes, n)}
	for c := n.FirstChild; c != nil; c = c.NextSibling {findNavboxes(c)}}
findNavboxes(doc)
isNested := make(map[*html.Node]bool)
for _, outer := range allNavboxes {
	for _, inner := range allNavboxes {
		if outer == inner {continue}
		var isDesc func(*html.Node) bool
		isDesc = func(n *html.Node) bool {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c == inner || isDesc(c) {return true}}
			return false}
		if isDesc(outer) {isNested[inner] = true}}}
var navboxParts []ConvertPart
for i, nb := range allNavboxes {
	if isNested[nb] {continue}
	nbTitle := wikiExtractText(wikiFind(nb, "", "navbox-title"))
	if nbTitle == "" {nbTitle = fmt.Sprintf("Нижний шаблон %d", i+1)}
	fullTitle := pageTitle + ":" + nbTitle
	var b strings.Builder
	wikiAsonFromNode(nb, nb, &b, info, fullTitle, tagName, &navboxParts)
	asonContent := strings.TrimSpace(b.String())
	if asonContent != "" {
		navboxParts = append(navboxParts, ConvertPart{
			Title:   fullTitle,
			Content: strings.ReplaceAll(strings.TrimSpace(asonContent), "\n", " "),
			TagName: tagName,
			IsHTML:  true,})}
	if nb.Parent != nil {nb.Parent.RemoveChild(nb)}}
// 2. Рендерим body после удаления навбоксов
var renderedHTML bytes.Buffer
var findBody func(*html.Node)
findBody = func(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "body" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {html.Render(&renderedHTML, c)}; return}
		for c := n.FirstChild; c != nil; c = c.NextSibling {findBody(c)}}
findBody(doc)
htmlContent = renderedHTML.String()
// 3. ИНФОБОКС
infoboxPattern := `(?s)(<table class="infobox.*?</table>)`
infoboxHTML, fullMatch := wikiExtractMatch(htmlContent, infoboxPattern)
var infoboxPart *ConvertPart
if infoboxHTML != "" {
	cleaned := wikiCleanupHTML(infoboxHTML, pageTitle, info, nil)
	infoboxPart = &ConvertPart{
		Title:   pageTitle + ": Шаблон-карточка",
		Content: strings.ReplaceAll(strings.TrimSpace(cleaned), "\n", " "),
		TagName: tagName,
		IsHTML:  true,}
	htmlContent = strings.Replace(htmlContent, fullMatch, "", 1)}
// 4. ПРИМЕЧАНИЯ
var notesSectionTitle string
notesKeywords := []string{"Примечания", "References", "Сноски", "Notes"}
headerFindRe  := regexp.MustCompile(`(?s)<h[2-4].*?>(.*?)</h[2-4]>`)
for _, hdr := range headerFindRe.FindAllStringSubmatch(htmlContent, -1) {
	title := strings.TrimSpace(regexp.MustCompile("<[^>]*>").ReplaceAllString(hdr[1], ""))
	for _, kw := range notesKeywords {
		if strings.EqualFold(title, kw) {notesSectionTitle = title; break}}}
// 5. Чистим HTML
htmlContent = wikiCleanupHTML(htmlContent, pageTitle, info, &notesSectionTitle)
// 6. РАЗДЕЛЕНИЕ ПО H2-H4
headerRe     := regexp.MustCompile(`(?s)(<(h[2-4]).*?>.*?/h\d>)`)
allHeaders   := headerRe.FindAllStringSubmatch(htmlContent, -1)
splitContent := headerRe.Split(htmlContent, -1)
sourceLink   := `<p><br><i>Источник: <a href="` + pageURL + `" target="_blank" rel="noopener noreferrer">` + pageURL + `</a></i></p>`
// Трансклюзии
var transcluded []string
if infoboxPart != nil {
	parts = append(parts, *infoboxPart)
	transcluded = append(transcluded, "{{"+infoboxPart.Title+"}}")}
// Вступление
introTitle := pageTitle + ": Вступление"
parts = append(parts, ConvertPart{
	Title:   introTitle,
	Content: wikiClean(wikiCleanupHTML(splitContent[0]+sourceLink, pageTitle, info, nil)),
	TagName: tagName,
	IsHTML:  true,})
transcluded = append(transcluded, "{{"+introTitle+"}}")
// Разделы
if len(allHeaders) > 0 {
	var curH2, curH3 string
	for i, sec := range splitContent[1:] {
		hTag  := allHeaders[i][2]
		sTitle := strings.TrimSpace(regexp.MustCompile("<[^>]*>").ReplaceAllString(allHeaders[i][1], ""))
		var tTitle string
		switch hTag {
		case "h2": curH2 = sTitle; curH3 = ""; tTitle = pageTitle + ": " + curH2
		case "h3":
			curH3 = sTitle
			if curH2 != "" {tTitle = pageTitle + ": " + curH2 + " / " + curH3
			} else {tTitle = pageTitle + ": " + curH3}
		case "h4":
			if curH2 != "" && curH3 != "" {tTitle = pageTitle + ": " + curH2 + " / " + curH3 + " / " + sTitle
			} else if curH2 != "" {tTitle = pageTitle + ": " + curH2 + " / " + sTitle
			} else {tTitle = pageTitle + ": " + sTitle}}
		cleaned := wikiClean(wikiCleanupHTML(allHeaders[i][0]+" "+sec, pageTitle, info, nil))
		parts = append(parts, ConvertPart{
			Title:   tTitle,
			Content: cleaned,
			TagName: tagName,
			IsHTML:  true,})
		transcluded = append(transcluded, "{{"+tTitle+"}}")}}
// Навбоксы
for _, np := range navboxParts {
	parts = append(parts, np)
	transcluded = append(transcluded, "{{"+np.Title+"}}")}
// 7. КАТЕГОРИИ
categories, _ := wikiGetCategories(articleTitle, info.Domain)
if len(categories) > 0 {
	var catLinks []string
	for _, cat := range categories {
		name := strings.TrimPrefix(strings.TrimPrefix(cat, "Категория:"), "Category:")
		catLinks = append(catLinks, fmt.Sprintf(`<li><a href="https://%s/wiki/%s" target="_blank" rel="noopener noreferrer">%s</a></li>`,
			info.Domain, strings.ReplaceAll(cat, " ", "_"), name))}
	catTitle := pageTitle + ": Категории"
	parts = append(parts, ConvertPart{
		Title:   catTitle,
		Content: strings.ReplaceAll("<ul>"+strings.Join(catLinks, "")+"</ul>", "\n", " "),
		TagName: tagName,
		IsHTML:  true,})
	transcluded = append(transcluded, "{{"+catTitle+"}}")}
// Головной фрейм
headContent := strings.Join(transcluded, "\n")
parts = append([]ConvertPart{{
	Title:   pageTitle,
	Content: headContent,
	TagName: tagName,
	IsHTML:  true,}}, parts...)
return parts, nil}
// 
func wikiAsonFromNode(node, rootNode *html.Node, b *strings.Builder, info *wikiInfo, parentTitle, baseTag string, results *[]ConvertPart) {
if node == nil {return}
if wikiHasClass(node, "navbar", "mw-collapsible-toggle") || node.Data == "style" || node.Data == "script" {return}
if node.Type == html.TextNode {
	text := strings.TrimSpace(node.Data)
	if text != "" && text != "•" && text != "·" && text != "|" {
		wikiAddSpace(b); b.WriteString(text)}; return}
if node.Type == html.ElementNode && node.Data == "a" {
	href := wikiGetAttr(node, "href")
	if !strings.HasPrefix(href, "/wiki/Template:") && !strings.HasPrefix(href, "/wiki/Шаблон:") {
		text := wikiExtractText(node)
		if text != "" {
			wikiAddSpace(b)
			fullURL := fmt.Sprintf("https://%s%s", info.Domain, href)
			b.WriteString(fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, fullURL, text))}}; return}
isList := wikiHasClass(node, "navbox-list", "navbox-abovebelow")
if isList {wikiAddSpace(b); b.WriteString("<b>[</b> ")}
for c := node.FirstChild; c != nil; c = c.NextSibling {
	if wikiHasClass(c, "navbox-subgroup") && wikiHasClass(c, "mw-collapsible") {
		subTitle := wikiExtractText(wikiFind(c, "", "navbox-title"))
		if subTitle == "" {subTitle = "Вложенный шаблон"}
		fullSub := parentTitle + " / " + subTitle
		wikiAddSpace(b); b.WriteString(fmt.Sprintf(`<b>[[%s]]</b>`, fullSub))
		var subB strings.Builder
		wikiAsonFromNode(c, c, &subB, info, fullSub, baseTag, results)
		*results = append(*results, ConvertPart{
			Title:   fullSub,
			Content: strings.TrimSpace(subB.String()),
			TagName: baseTag,
			IsHTML:  true,})
	} else {wikiAsonFromNode(c, rootNode, b, info, parentTitle, baseTag, results)}}
if isList {b.WriteString(" <b>]</b>")}}
// 
func wikiCleanupHTML(h, pageTitle string, info *wikiInfo, notes *string) string {
h = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`).ReplaceAllString(h, "")
h = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`).ReplaceAllString(h, "")
h = regexp.MustCompile(`(?s)<table[^>]*class="[^"]*mbox[^"]*".*?</table>`).ReplaceAllString(h, "")
h = regexp.MustCompile(`(?s)<div[^>]*id="mp-[^"]*"[^>]*>.*?</div>`).ReplaceAllString(h, "")
h = strings.ReplaceAll(h, `src="//`, `src="https://`)
doc, err := html.Parse(strings.NewReader(h))
if err != nil {return h}
var removals []*html.Node
var walk func(*html.Node)
walk = func(n *html.Node) {
	if n.Type == html.ElementNode {
		cls := wikiGetAttr(n, "class")
		id  := wikiGetAttr(n, "id")
		shouldRemove :=
			strings.Contains(cls, "mw-editsection") ||
			strings.Contains(cls, "mw-jump-link") ||
			strings.Contains(cls, "navbox") ||
			strings.Contains(cls, "mbox") ||
			strings.Contains(cls, "mw-empty-elt") ||
			strings.Contains(cls, "noprint") ||
			strings.Contains(cls, "mw-indicators") ||
			strings.HasPrefix(id, "mp-") ||
            strings.Contains(cls, "mw-cite-backlink") ||
			strings.Contains(cls, "portal-column")
		if shouldRemove {removals = append(removals, n); return}
		for i, a := range n.Attr {
			if a.Key == "href" && strings.HasPrefix(a.Val, "/wiki/") {
				n.Attr[i].Val = fmt.Sprintf("https://%s%s", info.Domain, a.Val)
				n.Attr = wikiAppendAttr(n.Attr, "target", "_blank")}}
		if n.Data == "sup" && strings.Contains(cls, "reference") {
        wikiReplaceWithText(n, "__NOTES__")
        return}}
	for c := n.FirstChild; c != nil; c = c.NextSibling {walk(c)}}
walk(doc)
for _, n := range removals {
	if n.Parent != nil {n.Parent.RemoveChild(n)}}
for pass := 0; pass < 5; pass++ {
	var empty []*html.Node
	var findEmpty func(*html.Node)
	findEmpty = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "div" || n.Data == "span") {
			hasContent := false
			var check func(*html.Node)
			check = func(c *html.Node) {
				if hasContent {return}
				if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {hasContent = true; return}
				if c.Type == html.ElementNode && (c.Data == "a" || c.Data == "img" || c.Data == "table") {hasContent = true; return}
				for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {check(cc)}}
			check(n)
			if !hasContent {empty = append(empty, n); return}}
		for c := n.FirstChild; c != nil; c = c.NextSibling {findEmpty(c)}}
	findEmpty(doc)
	if len(empty) == 0 {break}
	for _, n := range empty {
		if n.Parent != nil {n.Parent.RemoveChild(n)}}}
//     
var buf bytes.Buffer
var extractBody func(*html.Node)
extractBody = func(n *html.Node) {
if n.Type == html.ElementNode && n.Data == "body" {
	for c := n.FirstChild; c != nil; c = c.NextSibling {html.Render(&buf, c)}; return}
	for c := n.FirstChild; c != nil; c = c.NextSibling {extractBody(c)}}
extractBody(doc)
result := buf.String()
result = regexp.MustCompile(`^(\s*</(div|span|ul|li|p|table|tbody|tr|td)>)+`).ReplaceAllString(strings.TrimSpace(result), "")
result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
return strings.TrimSpace(result) }
// 
func wikiInfoFromURL(pageURL string) (*wikiInfo, error) {
parsed, err := url.Parse(pageURL)
if err != nil {return nil, err}
parts := strings.Split(parsed.Host, ".")
projectName := "wikipedia"
if len(parts) >= 2 {projectName = parts[1]}
return &wikiInfo{Domain: parsed.Host, ProjectName: projectName}, nil}
// 
func wikiTitleFromURL(pageURL string) (string, error) {
parsed, err := url.Parse(pageURL)
if err != nil {return "", err}
return url.PathUnescape(strings.TrimPrefix(parsed.Path, "/wiki/"))}
// 
func wikiGetHTML(title, domain string) (string, error) {
apiURL := fmt.Sprintf("https://%s/w/api.php?action=parse&page=%s&prop=text&format=json&disabletoc=true",
domain, url.QueryEscape(title))
req, _ := http.NewRequest("GET", apiURL, nil)
req.Header.Set("User-Agent", "PARUS-Bot/1.0")
resp, err := http.DefaultClient.Do(req)
if err != nil {return "", err}
defer resp.Body.Close()
var res struct {
	Parse struct {
		Text struct{ Body string `json:"*"` } `json:"text"`
	} `json:"parse"`}
json.NewDecoder(resp.Body).Decode(&res)
return res.Parse.Text.Body, nil}
// 
func wikiGetCategories(title, domain string) ([]string, error) {
apiURL := fmt.Sprintf("https://%s/w/api.php?action=query&prop=categories&titles=%s&format=json&cllimit=max&clshow=!hidden",
domain, url.QueryEscape(title))
req, _ := http.NewRequest("GET", apiURL, nil)
req.Header.Set("User-Agent", "PARUS-Bot/1.0")
resp, err := http.DefaultClient.Do(req)
if err != nil {return nil, err}
defer resp.Body.Close()
var res struct {
	Query struct {
		Pages map[string]struct {
			Categories []struct{ Title string `json:"title"` } `json:"categories"`
		} `json:"pages"`
	} `json:"query"`}
json.NewDecoder(resp.Body).Decode(&res)
var cats []string
for _, pg := range res.Query.Pages {
	for _, c := range pg.Categories {cats = append(cats, c.Title)}}
return cats, nil}
// 
func wikiExtractTitle(htmlContent, fallback string) string {
re := regexp.MustCompile(`(?s)<h1.*?>(.*?)</h1>`)
m  := re.FindStringSubmatch(htmlContent)
if len(m) > 1 {
	t := regexp.MustCompile("<[^>]*>").ReplaceAllString(m[1], "")
	if t != "" {return strings.TrimSpace(t)}}
return fallback}
// 
func wikiExtractMatch(h, pattern string) (string, string) {
re := regexp.MustCompile(pattern)
m  := re.FindStringSubmatch(h)
if len(m) > 1 {return m[1], m[0]}
return "", ""}
// 
func wikiHasClass(n *html.Node, names ...string) bool {
if n == nil || n.Type != html.ElementNode {return false}
for _, a := range n.Attr {
	if a.Key == "class" {
		fields := strings.Fields(a.Val)
		for _, w := range names {
			for _, f := range fields {if f == w {return true}}}}}
return false}
// 
func wikiExtractText(n *html.Node) string {
if n == nil {return ""}
if wikiHasClass(n, "navbar", "navbox-toggler", "mw-collapsible-toggle") || n.Data == "style" || n.Data == "script" {return ""}
if n.Type == html.TextNode {
	text := n.Data
	text = strings.ReplaceAll(text, "[", "&#91;")
	text = strings.ReplaceAll(text, "]", "&#93;")
	text = strings.ReplaceAll(text, "{", "&#123;")
	text = strings.ReplaceAll(text, "}", "&#125;")
	return text}
var b bytes.Buffer
for c := n.FirstChild; c != nil; c = c.NextSibling {b.WriteString(wikiExtractText(c))}
return strings.TrimSpace(b.String())}
// 
func wikiFind(n *html.Node, tag, class string) *html.Node {
if n == nil {return nil}
if n.Type == html.ElementNode && (tag == "" || n.Data == tag) && (class == "" || wikiHasClass(n, class)) {return n}
for c := n.FirstChild; c != nil; c = c.NextSibling {
	if res := wikiFind(c, tag, class); res != nil {return res}}
return nil }
// 
func wikiAddSpace(b *strings.Builder) {
if b.Len() > 0 && !strings.HasSuffix(b.String(), " ") && !strings.HasSuffix(b.String(), "<b>[</b> ") {b.WriteString(" ")}}
// 
func wikiGetAttr(n *html.Node, key string) string {
for _, a := range n.Attr {if a.Key == key {return a.Val}}
return ""}
// 
func wikiAppendAttr(attrs []html.Attribute, key, val string) []html.Attribute {
for _, a := range attrs {if a.Key == key {return attrs}}
return append(attrs, html.Attribute{Key: key, Val: val})}
// 
func wikiReplaceWithText(n *html.Node, text string) {
n.Type = html.TextNode
n.Data = text
n.FirstChild = nil
n.LastChild  = nil
n.Attr       = nil}
// 
func wikiClean(s string) string {
s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
return strings.TrimSpace(s) }