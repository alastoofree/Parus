// parus/converters/wordpress.go

package converters

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time")

// --- Структуры для wordpress.com API ---
type wpComPostAuthor struct {Name string `json:"name"`}
type wpComPostTags   map[string]struct {Name string `json:"name"`}
type wpComPost struct {
	ID      int             `json:"ID"`
	URL     string          `json:"URL"`
	Title   string          `json:"title"`
	Content string          `json:"content"`
	Author  wpComPostAuthor `json:"author"`
	Tags    wpComPostTags   `json:"tags"`}
type wpComComment struct {
	ID      int         `json:"ID"`
	URL     string      `json:"URL"`
	Author  struct {Name string `json:"name"`} `json:"author"`
	Date    string      `json:"date"`
	Content string      `json:"content"`
	Parent  interface{} `json:"parent"`}

// --- Структуры для самохостинга WP REST API v2 ---
type wpRenderedField struct {Rendered string `json:"rendered"`}
type wpSelfPost struct {
	ID       int    `json:"id"`
	Title    wpRenderedField `json:"title"`
	Content  wpRenderedField `json:"content"`
	Link     string          `json:"link"`
	Embedded struct {
		Author []struct {Name string `json:"name"`} `json:"author"`
		WpTerm [][]struct {Name string `json:"name"`} `json:"wp:term"`
	} `json:"_embedded"`}
type wpSelfComment struct {
	ID         int             `json:"id"`
	Post       int             `json:"post"`
	Parent     int             `json:"parent"`
	AuthorName string          `json:"author_name"`
	Date       string          `json:"date"`
	Content    wpRenderedField `json:"content"`
	Link       string          `json:"link"`}

func convertWordpress(pageURL string) ([]ConvertPart, error) {
parsed, err := url.Parse(pageURL)
if err != nil {return nil, fmt.Errorf("некорректный URL: %w", err)}
host     := parsed.Host
isWpCom  := strings.HasSuffix(host, ".wordpress.com")
blogName := Slugify(host)
importTag := "wordpress-" + blogName
var parts []ConvertPart
tagNames := []string{importTag, "wordpress-пост", "wordpress-комментарий", "wordpress-import"}
uniqueTags := make(map[string]bool)
for _, tn := range tagNames {
	if !uniqueTags[tn] {
		parts = append(parts, ConvertPart{Title: tn, IsTag: true})
		uniqueTags[tn] = true}}
postSlug := wpExtractSlug(parsed.Path)
if isWpCom {
	log.Printf("WordPress: режим wordpress.com, хост: %s", host)
	if postSlug != "" {
		p, err := wpFetchComSinglePost(host, postSlug, importTag, uniqueTags)
		if err != nil {return nil, err}
		parts = append(parts, p...)
	} else {
		p, err := wpFetchComAll(host, importTag, uniqueTags)
		if err != nil {return nil, err}
		parts = append(parts, p...)}
} else {
	log.Printf("WordPress: режим самохостинга, хост: %s", host)
	if postSlug != "" {
		p, err := wpFetchSelfSinglePost(host, postSlug, importTag, uniqueTags)
		if err != nil {return nil, err}
		parts = append(parts, p...)
	} else {
		p, err := wpFetchSelfAll(host, importTag, uniqueTags)
		if err != nil {return nil, err}
		parts = append(parts, p...)}}
log.Printf("WordPress: импорт завершён, всего частей: %d", len(parts))
return parts, nil }
// wpFetchComSinglePost загружает один пост wordpress.com по slug
func wpFetchComSinglePost(host, slug, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
apiURL := fmt.Sprintf("https://public-api.wordpress.com/rest/v1.1/sites/%s/posts/slug:%s", host, slug)
body, err := wpGet(apiURL)
if err != nil {return nil, err}
var post wpComPost
if err := json.Unmarshal(body, &post); err != nil {return nil, err}
comments, _ := wpGetComComments(host, post.ID)
hierarchy := wpBuildComHierarchy(comments)
return wpConvertComPost(post, importTag, uniqueTags, comments, hierarchy), nil}
// wpFetchComAll загружает все посты wordpress.com
func wpFetchComAll(host, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
var parts []ConvertPart
page := 1
for {
	apiURL := fmt.Sprintf("https://public-api.wordpress.com/rest/v1.1/sites/%s/posts?page=%d&fields=ID,URL,title,content,author,tags,slug", host, page)
	body, err := wpGet(apiURL)
	if err != nil {break}
	var resp struct {Posts []wpComPost `json:"posts"`}
	if err := json.Unmarshal(body, &resp); err != nil || len(resp.Posts) == 0 {break}
	for _, post := range resp.Posts {
		comments, _ := wpGetComComments(host, post.ID)
		hierarchy := wpBuildComHierarchy(comments)
		parts = append(parts, wpConvertComPost(post, importTag, uniqueTags, comments, hierarchy)...)
		time.Sleep(250 * time.Millisecond)}
	page++}
return parts, nil }
// wpFetchSelfSinglePost загружает один пост самохостинга по slug
func wpFetchSelfSinglePost(host, slug, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
apiURL := fmt.Sprintf("https://%s/wp-json/wp/v2/posts?slug=%s&_embed=author,wp:term", host, slug)
body, err := wpGet(apiURL)
if err != nil {return nil, err}
var posts []wpSelfPost
if err := json.Unmarshal(body, &posts); err != nil || len(posts) == 0 {
	return nil, fmt.Errorf("пост с slug '%s' не найден", slug)}
post := posts[0]
commentsURL := fmt.Sprintf("https://%s/wp-json/wp/v2/comments?post=%d&per_page=100&order=asc", host, post.ID)
var comments []wpSelfComment
if body2, err := wpGet(commentsURL); err == nil {
	json.Unmarshal(body2, &comments)}
hierarchy := wpBuildSelfHierarchy(comments)
return wpConvertSelfPost(post, importTag, uniqueTags, comments, hierarchy), nil}
// wpFetchSelfAll загружает все посты самохостинга
func wpFetchSelfAll(host, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
var parts []ConvertPart
// Посты
var allPosts []wpSelfPost
page := 1
for {
	apiURL := fmt.Sprintf("https://%s/wp-json/wp/v2/posts?page=%d&_embed=author,wp:term", host, page)
	body, err := wpGet(apiURL)
	if err != nil || body == nil {break}
	var posts []wpSelfPost
	if err := json.Unmarshal(body, &posts); err != nil || len(posts) == 0 {break}
	allPosts = append(allPosts, posts...)
	page++
	time.Sleep(250 * time.Millisecond)}
// Комментарии
var allComments []wpSelfComment
page = 1
for {
	apiURL := fmt.Sprintf("https://%s/wp-json/wp/v2/comments?page=%d&per_page=100&order=asc", host, page)
	body, err := wpGet(apiURL)
	if err != nil || body == nil {break}
	var comments []wpSelfComment
	if err := json.Unmarshal(body, &comments); err != nil || len(comments) == 0 {break}
	allComments = append(allComments, comments...)
	page++
	time.Sleep(250 * time.Millisecond)}
hierarchy := wpBuildSelfHierarchy(allComments)
postMap := make(map[int]wpSelfPost)
for _, p := range allPosts {postMap[p.ID] = p}
for _, post := range allPosts {
	var postComments []wpSelfComment
	for _, c := range allComments {
		if c.Post == post.ID {postComments = append(postComments, c)}}
	parts = append(parts, wpConvertSelfPost(post, importTag, uniqueTags, postComments, hierarchy)...)}
return parts, nil }
// wpConvertComPost конвертирует один пост wordpress.com
func wpConvertComPost(post wpComPost, importTag string, uniqueTags map[string]bool, comments []wpComComment, hierarchy map[int]int) []ConvertPart {
var parts []ConvertPart
cleanTitle := html.UnescapeString(post.Title)
postTagName := importTag + "-" + Slugify(cleanTitle)
if !uniqueTags[postTagName] {
	parts = append(parts, ConvertPart{Title: postTagName, IsTag: true})
	uniqueTags[postTagName] = true}
for _, tag := range post.Tags {
	labelTag := "wp-tag-" + Slugify(tag.Name)
	if !uniqueTags[labelTag] {
		parts = append(parts, ConvertPart{Title: labelTag, IsTag: true})
		uniqueTags[labelTag] = true}}
content := post.Content
if post.Author.Name != "" {
	content += fmt.Sprintf("<p><b>Автор:</b> %s</p>", post.Author.Name)}
if post.URL != "" {
	content += fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, post.URL, post.URL)}
// Ссылки на комментарии первого уровня
idToTitle := make(map[int]string)
for _, c := range comments {
	idToTitle[c.ID] = fmt.Sprintf("%s-comment-%d", postTagName, c.ID)}
var level1Links strings.Builder
for _, c := range comments {
	if _, hasParent := hierarchy[c.ID]; !hasParent {
		level1Links.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[c.ID]))}}
if level1Links.Len() > 0 {
	content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"}
parts = append(parts, ConvertPart{
	Title:   cleanTitle,
	Content: strings.ReplaceAll(strings.TrimSpace(content), "\n", " "),
	TagName: postTagName,
	IsHTML:  true,})
// childMap для ссылок
childMap := make(map[int][]int)
for id, parentID := range hierarchy {
	childMap[parentID] = append(childMap[parentID], id)}
for _, c := range comments {
	commentTitle := idToTitle[c.ID]
	parentTag := postTagName
	if parentID, ok := hierarchy[c.ID]; ok {
		if pt, ok := idToTitle[parentID]; ok {parentTag = pt}}
	var b strings.Builder
	if c.Author.Name != "" {b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", c.Author.Name))}
	if c.Date != "" {b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", c.Date))}
	b.WriteString("<hr>")
	b.WriteString(c.Content)
	if children, ok := childMap[c.ID]; ok {
		b.WriteString("<p><b>Ответы:</b> ")
		for _, childID := range children {
			b.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[childID]))}
		b.WriteString("</p>")}
	parts = append(parts, ConvertPart{
		Title:   commentTitle,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: parentTag,
		Rank:    "0",
		IsHTML:  true,})}
return parts }
// wpConvertSelfPost конвертирует один пост самохостинга
func wpConvertSelfPost(post wpSelfPost, importTag string, uniqueTags map[string]bool, comments []wpSelfComment, hierarchy map[int]int) []ConvertPart {
var parts []ConvertPart
cleanTitle := html.UnescapeString(post.Title.Rendered)
postTagName := importTag + "-" + Slugify(cleanTitle)
if !uniqueTags[postTagName] {
	parts = append(parts, ConvertPart{Title: postTagName, IsTag: true})
	uniqueTags[postTagName] = true}
if len(post.Embedded.WpTerm) > 0 {
	for _, termList := range post.Embedded.WpTerm {
		for _, term := range termList {
			labelTag := "wp-tag-" + Slugify(term.Name)
			if !uniqueTags[labelTag] {
				parts = append(parts, ConvertPart{Title: labelTag, IsTag: true})
				uniqueTags[labelTag] = true}}}}
content := html.UnescapeString(post.Content.Rendered)
if len(post.Embedded.Author) > 0 {
	content += fmt.Sprintf("<p><b>Автор:</b> %s</p>", html.UnescapeString(post.Embedded.Author[0].Name))}
if post.Link != "" {
	content += fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, post.Link, post.Link)}
// Ссылки на комментарии первого уровня
idToTitle := make(map[int]string)
for _, c := range comments {
	idToTitle[c.ID] = fmt.Sprintf("%s-comment-%d", postTagName, c.ID)}
var level1Links strings.Builder
for _, c := range comments {
	if _, hasParent := hierarchy[c.ID]; !hasParent {
		level1Links.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[c.ID]))}}
if level1Links.Len() > 0 {
	content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"}
parts = append(parts, ConvertPart{
	Title:   cleanTitle,
	Content: strings.ReplaceAll(strings.TrimSpace(content), "\n", " "),
	TagName: postTagName,
	IsHTML:  true,})
// childMap для ссылок
childMap := make(map[int][]int)
for id, parentID := range hierarchy {
	childMap[parentID] = append(childMap[parentID], id)}
for _, c := range comments {
	commentTitle := idToTitle[c.ID]
	parentTag := postTagName
	if parentID, ok := hierarchy[c.ID]; ok {
		if pt, ok := idToTitle[parentID]; ok {parentTag = pt}}
	var b strings.Builder
	if c.AuthorName != "" {b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", c.AuthorName))}
	if c.Date != "" {b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", c.Date))}
	b.WriteString("<hr>")
	b.WriteString(html.UnescapeString(c.Content.Rendered))
	if children, ok := childMap[c.ID]; ok {
		b.WriteString("<p><b>Ответы:</b> ")
		for _, childID := range children {
			b.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[childID]))}
		b.WriteString("</p>")}
	parts = append(parts, ConvertPart{
		Title:   commentTitle,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: parentTag,
		Rank:    "0",
		IsHTML:  true,})}
return parts }
// wpBuildComHierarchy строит карту id→parentID для комментариев wordpress.com
func wpBuildComHierarchy(comments []wpComComment) map[int]int {
hierarchy := make(map[int]int)
for _, c := range comments {
	if parentMap, ok := c.Parent.(map[string]interface{}); ok {
		if parentID, ok := parentMap["ID"].(float64); ok && int(parentID) != 0 {
			hierarchy[c.ID] = int(parentID)}}}
return hierarchy }
// wpBuildSelfHierarchy строит карту id→parentID для комментариев самохостинга
func wpBuildSelfHierarchy(comments []wpSelfComment) map[int]int {
hierarchy := make(map[int]int)
for _, c := range comments {
	if c.Parent != 0 {hierarchy[c.ID] = c.Parent}}
return hierarchy }
// wpGetComComments загружает комментарии поста wordpress.com
func wpGetComComments(host string, postID int) ([]wpComComment, error) {
apiURL := fmt.Sprintf("https://public-api.wordpress.com/rest/v1.1/sites/%s/posts/%d/replies/?order=ASC", host, postID)
body, err := wpGet(apiURL)
if err != nil {return nil, err}
var resp struct {Comments []wpComComment `json:"comments"`}
if err := json.Unmarshal(body, &resp); err != nil {return nil, err}
return resp.Comments, nil }
// wpGet выполняет GET запрос
func wpGet(apiURL string) ([]byte, error) {
resp, err := http.Get(apiURL)
if err != nil {return nil, err}
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return nil, fmt.Errorf("статус %d", resp.StatusCode)}
return io.ReadAll(resp.Body) }
// wpExtractSlug извлекает slug поста из пути URL
func wpExtractSlug(path string) string {
path = strings.Trim(path, "/")
if path == "" {return ""}
parts := strings.Split(path, "/")
for i := len(parts) - 1; i >= 0; i-- {
	if parts[i] != "" {return parts[i]}}
return "" }