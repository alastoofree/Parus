// parus/converters/lj.go

package converters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html")

type ljPost struct {
	Title string
	URL   string
	Body  string
	Tags  []string}

type ljComment struct {
	ID        float64
	Level     int
	Parent    float64
	ThreadURL string}

func convertLJ(pageURL string) ([]ConvertPart, error) {
	isPost, _ := regexp.MatchString(`/\d+\.html$`, pageURL)
	if !isPost {
		return nil, fmt.Errorf("ожидается URL поста ЖЖ вида /123456.html, получено: %s", pageURL)}
	client := &http.Client{Timeout: 30 * time.Second}
	blogName  := ljExtractBlogName(pageURL)
	importTag := "livejournal-" + blogName
	var parts []ConvertPart
	tagNames := []string{importTag, "livejournal-пост", "livejournal-комментарий", "livejournal-import"}
	uniqueTags := make(map[string]bool)
	for _, tn := range tagNames {
		if !uniqueTags[tn] {
			parts = append(parts, ConvertPart{Title: tn, IsTag: true})
			uniqueTags[tn] = true}}
	post, postBodyBytes, err := ljFetchPost(pageURL, client)
	if err != nil {return nil, fmt.Errorf("ошибка загрузки поста: %w", err)}
	for _, ljTag := range post.Tags {
		tagKey := "lj-" + strings.ToLower(strings.ReplaceAll(ljTag, " ", "-"))
		if !uniqueTags[tagKey] {
			parts = append(parts, ConvertPart{Title: tagKey, IsTag: true})
			uniqueTags[tagKey] = true}}
	postSlug    := Slugify(post.Title)
	postTagName := importTag + "-" + postSlug
	if !uniqueTags[postTagName] {
		parts = append(parts, ConvertPart{Title: postTagName, IsTag: true})
		uniqueTags[postTagName] = true}
	sourceLink := ""
	if post.URL != "" {
		sourceLink = fmt.Sprintf(` <p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, post.URL, post.URL)}
	parts = append(parts, ConvertPart{
		Title:   post.Title,
		Content: strings.ReplaceAll(strings.TrimSpace(post.Body+sourceLink), "\n", " "),
		TagName: postTagName,
		IsHTML:  true,})
	commentParts := ljFetchComments(pageURL, postTagName, client, postBodyBytes)
    // Добавляем ссылки на комментарии первого уровня в контент поста
    var level1Links strings.Builder
    for _, cp := range commentParts {
        if cp.TagName == postTagName && !cp.IsTag {
            level1Links.WriteString(fmt.Sprintf("[[%s]] ", cp.Title))}}
    if level1Links.Len() > 0 {
        // Добавляем к последнему ConvertPart поста
        for i, p := range parts {
            if p.Title == post.Title {
                parts[i].Content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"; break}}}
	parts = append(parts, commentParts...)
	log.Printf("LJ: пост '%s' импортирован. Всего частей: %d", post.Title, len(parts))
	return parts, nil}
// 
func ljFetchPost(pageURL string, client *http.Client) (*ljPost, []byte, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {return nil, nil, err}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {return nil, nil, err}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("статус %d при запросе поста", resp.StatusCode)}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {return nil, nil, err}
	post, err := ljParsePost(bodyBytes)
	if err != nil {return nil, nil, err}
	return post, bodyBytes, nil}
// 
func ljParsePost(htmlBody []byte) (*ljPost, error) {
doc, err := html.Parse(bytes.NewReader(htmlBody))
if err != nil {return nil, err}
post := &ljPost{Tags: make([]string, 0)}
var bodyNode *html.Node
var traverse func(*html.Node)
traverse = func(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "meta" {
		property := ljGetAttr(n, "property")
		content  := ljGetAttr(n, "content")
		switch property {
		case "og:title": if post.Title == "" {post.Title = content}
		case "og:url": if post.URL == "" {post.URL = content}
		case "article:tag": if content != "" {post.Tags = append(post.Tags, content)}}}
	if bodyNode == nil && n.Type == html.ElementNode && n.Data == "div" {
		class := ljGetAttr(n, "class")
		if strings.Contains(class, "entry-content") ||
			strings.Contains(class, "aentry-post__text") ||
			strings.Contains(class, "asset-body") {
			bodyNode = n; return}}
	if bodyNode == nil {
		for c := n.FirstChild; c != nil; c = c.NextSibling {traverse(c)}}}
traverse(doc)
if post.Title == "" {
	if titleNode := ljFindNode(doc, "title"); titleNode != nil {
		post.Title = ljGetTitleText(titleNode)}}
if post.Title == "" {return nil, fmt.Errorf("не удалось найти заголовок поста")}
if bodyNode != nil {
	var b bytes.Buffer
	for c := bodyNode.FirstChild; c != nil; c = c.NextSibling {html.Render(&b, c)}
	post.Body = b.String()
} else {
	post.Body = "<p>Тело поста не найдено.</p>"}
return post, nil }
// 
func ljFetchComments(pageURL, postTagName string, client *http.Client, postBodyBytes []byte) []ConvertPart {
// Шаг 1: получаем replycount со страницы поста
replyCount := ljExtractReplyCount(postBodyBytes)
log.Printf("LJ: всего комментариев по replycount: %d", replyCount)
// Шаг 2: собираем все комментарии по страницам
seen := make(map[float64]bool)
var allComments []ljComment
page := 1
for {
	var pageBytes []byte
	if page == 1 {
        b, err := ljGet(fmt.Sprintf("%s?page=%d", pageURL, page), client)
        if err != nil {break}
        pageBytes = b
	} else {
		b, err := ljGet(fmt.Sprintf("%s?page=%d", pageURL, page), client)
		if err != nil {break}
		pageBytes = b}
	comments := ljExtractComments(pageBytes, pageURL)
	added := 0
	for _, c := range comments {
		if !seen[c.ID] {
			seen[c.ID] = true
			allComments = append(allComments, c)
			added++}}
	log.Printf("LJ: страница %d — найдено %d комментариев (всего %d)", page, added, len(allComments))
	if len(allComments) >= replyCount || added == 0 {break}
	page++}
// Шаг 2.б: для каждого комментария level=1 идём на его страницу
    var level1 []ljComment
    for _, c := range allComments {
        if c.Level == 1 {level1 = append(level1, c)}}
    for _, c := range level1 {
    log.Printf("LJ: обрабатываем ветку id=%.0f", c.ID)
    threadURL := c.ThreadURL
    if idx := strings.Index(threadURL, "#"); idx != -1 {
        threadURL = threadURL[:idx]}
    if threadBytes, err := ljGet(threadURL+"?view=comments", client); err == nil {
        children := ljExtractComments(threadBytes, pageURL)
        for _, child := range children {
            if !seen[child.ID] {
                seen[child.ID] = true
                allComments = append(allComments, child)
                log.Printf("LJ: вложенный id=%.0f parent=%.0f level=%d", child.ID, child.Parent, child.Level)}}}}
    log.Printf("LJ: итого в карте: %d комментариев", len(allComments))
	if len(allComments) == 0 {
	log.Printf("LJ: комментарии не найдены")
    log.Printf("LJ: карта построена: %d комментариев", len(allComments))
	return nil}
// Шаг 3: строим иерархию и создаём фреймы
var parts []ConvertPart
idToTitle := make(map[float64]string)
idToTitle[0] = postTagName
log.Printf("LJ: загружаем контент...")
childMap := make(map[float64][]float64)
for _, c := range allComments {
    childMap[c.Parent] = append(childMap[c.Parent], c.ID)}
for _, c := range allComments {
    idToTitle[c.ID] = fmt.Sprintf("%s-comment-%.0f", postTagName, c.ID)}
for _, c := range allComments {
	commentTitle := fmt.Sprintf("%s-comment-%.0f", postTagName, c.ID)
    parentTag := idToTitle[c.Parent]
    if parentTag == "" {parentTag = postTagName}
	threadURL := c.ThreadURL
	if idx := strings.Index(threadURL, "#"); idx != -1 {
		threadURL = threadURL[:idx]}
	articleText := ""
	if threadURL != "" {
		if threadBytes, err := ljGet(threadURL+"?view=comments", client); err == nil {
			articleText = ljExtractArticleByID(threadBytes, c.ID)}}
	if articleText == "" {
		articleText = "<p><i>Комментарий скрыт или удалён.</i></p>"}
	var b strings.Builder
	b.WriteString(articleText)
    if children, ok := childMap[c.ID]; ok {
        b.WriteString("<p><b>Ответы:</b> ")
        for _, childID := range children {
            b.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[childID]))}
        b.WriteString("</p>")}
	if c.ThreadURL != "" {
		b.WriteString(fmt.Sprintf(`<p><b>Ссылка:</b> <a href="%s" target="_blank">%s</a></p>`, c.ThreadURL, c.ThreadURL))}
	parts = append(parts, ConvertPart{
		Title:   commentTitle,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: parentTag,
		Rank:    "0",
		IsHTML:  true,})}
log.Printf("LJ: создано %d фреймов комментариев", len(parts))
return parts}
// ljExtractReplyCount извлекает общее число комментариев из "replycount":N
func ljExtractReplyCount(htmlBody []byte) int {
re := regexp.MustCompile(`"replycount":(\d+)`)
m  := re.FindSubmatch(htmlBody)
if len(m) < 2 {return 0}
n := 0
fmt.Sscanf(string(m[1]), "%d", &n)
return n }
// ljExtractComments извлекает все комментарии со страницы по маркерам thread/level/thread_url
func ljExtractComments(htmlBody []byte, postURL string) []ljComment {
reParent := regexp.MustCompile(`"parent":(\d+)`)
reThread := regexp.MustCompile(`"thread":(\d+)`)
reLevel  := regexp.MustCompile(`"level":(\d+)`)
var comments []ljComment
pos := 0
for {
	pm := reParent.FindSubmatchIndex(htmlBody[pos:])
	if pm == nil {break}
	var parent float64
	fmt.Sscanf(string(htmlBody[pos+pm[2]:pos+pm[3]]), "%f", &parent)
	pos += pm[1]
	tm := reThread.FindSubmatchIndex(htmlBody[pos:])
	if tm == nil {break}
	var id float64
	fmt.Sscanf(string(htmlBody[pos+tm[2]:pos+tm[3]]), "%f", &id)
	pos += tm[1]
	lm := reLevel.FindSubmatchIndex(htmlBody[pos:])
	if lm == nil {break}
	var level int
	fmt.Sscanf(string(htmlBody[pos+lm[2]:pos+lm[3]]), "%d", &level)
	pos += lm[1]
	if id == 0 {continue}
	comments = append(comments, ljComment{ID: id, Parent: parent, Level: level,
        ThreadURL: ljBuildThreadURL(postURL, id)})}
return comments}
// ljExtractArticleByID извлекает article конкретного комментария со страницы
func ljExtractArticleByID(htmlBody []byte, commentID float64) string {
largestJSON := ljExtractJSON(htmlBody)
if largestJSON == nil {return ""}
var sitePage map[string]interface{}
if err := json.Unmarshal(largestJSON, &sitePage); err != nil {return ""}
commentsData, ok := sitePage["comments"].([]interface{})
if !ok {return ""}
for _, comm := range commentsData {
	commentMap, _ := comm.(map[string]interface{})
	var cID float64
	if id, ok := commentMap["thread"].(float64); ok {cID = id
	} else if id, ok := commentMap["dtalkid"].(float64); ok {cID = id}
	if cID == commentID {
		article, _  := commentMap["article"].(string)
		author, _   := commentMap["dname"].(string)
		datetime, _ := commentMap["ctime"].(string)
		var b strings.Builder
		if author != "" {b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", author))}
		if datetime != "" {b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", datetime))}
		if article != "" {b.WriteString(article)}
		return b.String()}}
return ""}
// ljExtractJSON находит Site.page JSON балансируя скобки
func ljExtractJSON(htmlBody []byte) []byte {
marker := []byte("Site.page = {")
idx := bytes.Index(htmlBody, marker)
if idx < 0 {return nil}
start := idx + len(marker) - 1
depth := 0
inString := false
escaped := false
for i := start; i < len(htmlBody); i++ {
	c := htmlBody[i]
	if escaped {escaped = false; continue}
	if c == '\\' && inString {escaped = true; continue}
	if c == '"' {inString = !inString; continue}
	if inString {continue}
	if c == '{' {depth++}
	if c == '}' {
		depth--
		if depth == 0 {return htmlBody[start : i+1]}}}
return nil }
// 
func ljGet(pageURL string, client *http.Client) ([]byte, error) {
req, err := http.NewRequest("GET", pageURL, nil)
if err != nil {return nil, err}
req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
resp, err := client.Do(req)
if err != nil {return nil, err}
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return nil, fmt.Errorf("статус %d", resp.StatusCode)}
return io.ReadAll(resp.Body) }
// 
func ljExtractBlogName(pageURL string) string {
u, err := url.Parse(pageURL)
if err != nil {return "unknown"}
parts := strings.SplitN(u.Hostname(), ".", 2)
if len(parts) > 0 && parts[0] != "" {return strings.ToLower(parts[0])}
return "unknown"}
// 
func ljGetAttr(n *html.Node, key string) string {
for _, attr := range n.Attr {
	if attr.Key == key {return attr.Val}}
return "" }
// 
func ljFindNode(n *html.Node, tagName string) *html.Node {
if n.Type == html.ElementNode && n.Data == tagName {return n}
for c := n.FirstChild; c != nil; c = c.NextSibling {
	if result := ljFindNode(c, tagName); result != nil {return result}}
return nil }
// 
func ljGetTitleText(n *html.Node) string {
if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
	return strings.TrimSuffix(strings.TrimSpace(n.FirstChild.Data), " — ЖЖ")}
return "" }
// 
func ljBuildThreadURL(postURL string, id float64) string {
base := postURL
if idx := strings.Index(base, "?"); idx != -1 {base = base[:idx]}
return fmt.Sprintf("%s?thread=%.0f#t%.0f", base, id, id) }