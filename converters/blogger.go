// parus/converters/blogger.go

package converters

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time")

const bloggerAPIKey = "AIzaSyCFFS96JjXHqOiwZntdCPFnW5kESrC6OKg"
const bloggerAPIBase = "https://www.googleapis.com/blogger/v3"

type bloggerBlog struct {
	ID   string `json:"id"`
	Name string `json:"name"`}
type bloggerPost struct {
	ID      string       `json:"id"`
	Title   string       `json:"title"`
	Content string       `json:"content"`
	URL     string       `json:"url"`
	Labels  []string     `json:"labels"`
	Author  *bloggerUser `json:"author"`}
type bloggerUser struct {
	DisplayName string `json:"displayName"`}
type bloggerComment struct {
	ID          string            `json:"id"`
	Content     string            `json:"content"`
	Published   string            `json:"published"`
	Author      *bloggerUser      `json:"author"`
	InReplyTo   *bloggerReplyTo   `json:"inReplyTo"`}
type bloggerReplyTo struct {
	ID string `json:"id"`}
type bloggerPostList struct {
	Items         []*bloggerPost `json:"items"`
	NextPageToken string         `json:"nextPageToken"`}
type bloggerCommentList struct {
	Items         []*bloggerComment `json:"items"`
	NextPageToken string            `json:"nextPageToken"`}

func convertBlogger(pageURL string) ([]ConvertPart, error) {
client := &http.Client{Timeout: 30 * time.Second}
// Получаем ID блога
baseURL := bloggerExtractBaseURL(pageURL)
blog, err := bloggerGetBlog(client, baseURL)
if err != nil {return nil, fmt.Errorf("ошибка получения блога: %w", err)}
blogName  := bloggerSlugify(blog.Name)
importTag := "blogger-" + blogName
log.Printf("Blogger: блог '%s' (ID: %s)", blog.Name, blog.ID)
var parts []ConvertPart
// Теги
tagNames := []string{importTag, "blogger-пост", "blogger-комментарий", "blogger-import"}
uniqueTags := make(map[string]bool)
for _, tn := range tagNames {
	if !uniqueTags[tn] {
		parts = append(parts, ConvertPart{Title: tn, IsTag: true})
		uniqueTags[tn] = true}}
// Определяем режим — один пост или весь блог
postPath := bloggerExtractPostPath(pageURL)
if postPath != "" && postPath != "/" {
	// Один пост
	post, err := bloggerGetPostByPath(client, blog.ID, postPath)
	if err != nil {return nil, fmt.Errorf("ошибка получения поста: %w", err)}
	postParts := bloggerConvertPost(client, blog.ID, post, importTag, uniqueTags)
	parts = append(parts, postParts...)
} else {
	// Весь блог
	posts, err := bloggerGetAllPosts(client, blog.ID)
	if err != nil {return nil, fmt.Errorf("ошибка получения постов: %w", err)}
	log.Printf("Blogger: найдено %d постов", len(posts))
	for _, post := range posts {
		postParts := bloggerConvertPost(client, blog.ID, post, importTag, uniqueTags)
		parts = append(parts, postParts...)
		time.Sleep(300 * time.Millisecond)}}
log.Printf("Blogger: импорт завершён, всего частей: %d", len(parts))
return parts, nil }
// bloggerConvertPost конвертирует один пост и его комментарии
func bloggerConvertPost(client *http.Client, blogID string, post *bloggerPost, importTag string, uniqueTags map[string]bool) []ConvertPart {
var parts []ConvertPart
postSlug    := importTag + "-" + bloggerSlugify(post.Title)
postTagName := postSlug
if !uniqueTags[postTagName] {
	parts = append(parts, ConvertPart{Title: postTagName, IsTag: true})
	uniqueTags[postTagName] = true}
// Метки поста
for _, label := range post.Labels {
	labelTag := "blogger-label-" + bloggerSlugify(label)
	if !uniqueTags[labelTag] {
		parts = append(parts, ConvertPart{Title: labelTag, IsTag: true})
		uniqueTags[labelTag] = true}}
// Контент поста
content := post.Content
content = strings.ReplaceAll(content, `src="http://`, `src="https://`)
content = strings.ReplaceAll(content, `src='http://`, `src='https://`)
if post.Author != nil && post.Author.DisplayName != "" {
	content += fmt.Sprintf("<p><b>Автор:</b> %s</p>", post.Author.DisplayName)}
if post.URL != "" {
	content += fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, post.URL, post.URL)}
// Комментарии
comments, err := bloggerGetAllComments(client, blogID, post.ID)
if err != nil {
	log.Printf("Blogger: ошибка получения комментариев для '%s': %v", post.Title, err)}
	log.Printf("Blogger: пост '%s' — %d комментариев", post.Title, len(comments))
// Строим карту id→title и id→parent для иерархии
idToTitle  := make(map[string]string)
idToParent := make(map[string]string)
for _, c := range comments {
	commentTitle := postTagName + "-comment-" + c.ID
	idToTitle[c.ID] = commentTitle
	if c.InReplyTo != nil && c.InReplyTo.ID != "" {
		idToParent[c.ID] = c.InReplyTo.ID}}
// Ссылки на комментарии первого уровня в посте
var level1Links strings.Builder
for _, c := range comments {
	if c.InReplyTo == nil || c.InReplyTo.ID == "" {
		level1Links.WriteString(fmt.Sprintf("[[%s]] ", idToTitle[c.ID]))}}
if level1Links.Len() > 0 {
	content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"}
parts = append(parts, ConvertPart{
	Title:   post.Title,
	Content: strings.ReplaceAll(strings.TrimSpace(content), "\n", " "),
	TagName: postTagName,
	IsHTML:  true,})
// Фреймы комментариев
childMap := make(map[string][]string)
for _, c := range comments {
	if c.InReplyTo != nil && c.InReplyTo.ID != "" {
		childMap[c.InReplyTo.ID] = append(childMap[c.InReplyTo.ID], c.ID)}}
for _, c := range comments {
	commentTitle := idToTitle[c.ID]
	parentTag := postTagName
	if c.InReplyTo != nil && c.InReplyTo.ID != "" {
		if pt, ok := idToTitle[c.InReplyTo.ID]; ok {
			parentTag = pt}}
	var b strings.Builder
	if c.Author != nil && c.Author.DisplayName != "" {
		b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", c.Author.DisplayName))}
	if c.Published != "" {
		b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", c.Published))}
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
// bloggerGetBlog получает информацию о блоге по URL
func bloggerGetBlog(client *http.Client, baseURL string) (*bloggerBlog, error) {
apiURL := fmt.Sprintf("%s/blogs/byurl?url=%s&key=%s", bloggerAPIBase, url.QueryEscape(baseURL), bloggerAPIKey)
body, err := bloggerGet(client, apiURL)
if err != nil {return nil, err}
var blog bloggerBlog
if err := json.Unmarshal(body, &blog); err != nil {return nil, err}
return &blog, nil}
// bloggerGetPostByPath получает пост по пути
func bloggerGetPostByPath(client *http.Client, blogID, path string) (*bloggerPost, error) {
apiURL := fmt.Sprintf("%s/blogs/%s/posts/bypath?path=%s&key=%s", bloggerAPIBase, blogID, url.QueryEscape(path), bloggerAPIKey)
body, err := bloggerGet(client, apiURL)
if err != nil {return nil, err}
var post bloggerPost
if err := json.Unmarshal(body, &post); err != nil {return nil, err}
return &post, nil }
// bloggerGetAllPosts получает все посты блога с пагинацией
func bloggerGetAllPosts(client *http.Client, blogID string) ([]*bloggerPost, error) {
var allPosts []*bloggerPost
pageToken := ""
for {
	apiURL := fmt.Sprintf("%s/blogs/%s/posts?maxResults=50&key=%s", bloggerAPIBase, blogID, bloggerAPIKey)
	if pageToken != "" {apiURL += "&pageToken=" + pageToken}
	body, err := bloggerGet(client, apiURL)
	if err != nil {return nil, err}
	var list bloggerPostList
	if err := json.Unmarshal(body, &list); err != nil {return nil, err}
	allPosts = append(allPosts, list.Items...)
	if list.NextPageToken == "" {break}
	pageToken = list.NextPageToken
	time.Sleep(300 * time.Millisecond)}
return allPosts, nil }
// bloggerGetAllComments получает все комментарии поста с пагинацией
func bloggerGetAllComments(client *http.Client, blogID, postID string) ([]*bloggerComment, error) {
var allComments []*bloggerComment
pageToken := ""
for {
	apiURL := fmt.Sprintf("%s/blogs/%s/posts/%s/comments?maxResults=50&key=%s", bloggerAPIBase, blogID, postID, bloggerAPIKey)
	if pageToken != "" {apiURL += "&pageToken=" + pageToken}
	body, err := bloggerGet(client, apiURL)
	if err != nil {return nil, err}
	var list bloggerCommentList
	if err := json.Unmarshal(body, &list); err != nil {return nil, err}
	allComments = append(allComments, list.Items...)
	if list.NextPageToken == "" {break}
	pageToken = list.NextPageToken
	time.Sleep(300 * time.Millisecond)}
return allComments, nil }
// bloggerGet выполняет GET запрос и возвращает тело ответа
func bloggerGet(client *http.Client, apiURL string) ([]byte, error) {
resp, err := client.Get(apiURL)
if err != nil {return nil, err}
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
	return nil, fmt.Errorf("статус %d при запросе %s", resp.StatusCode, apiURL)}
return io.ReadAll(resp.Body) }
// bloggerExtractBaseURL извлекает базовый URL блога
func bloggerExtractBaseURL(rawURL string) string {
for _, prefix := range []string{"https://", "http://"} {
	if strings.HasPrefix(rawURL, prefix) {
		rest := rawURL[len(prefix):]
		if slash := strings.Index(rest, "/"); slash != -1 {
			return prefix + rest[:slash]}
		return rawURL}}
return rawURL }
// bloggerExtractPostPath извлекает путь поста из URL
func bloggerExtractPostPath(rawURL string) string {
for _, prefix := range []string{"https://", "http://"} {
	if strings.HasPrefix(rawURL, prefix) {
		rest := rawURL[len(prefix):]
		if slash := strings.Index(rest, "/"); slash != -1 {
			return rest[slash:]}}}
return "/"}
// bloggerSlugify преобразует строку в slug
func bloggerSlugify(s string) string {
s = strings.ToLower(s)
s = strings.ReplaceAll(s, " ", "-")
s = strings.ReplaceAll(s, `"`, "")
s = strings.ReplaceAll(s, `'`, "")
return s }