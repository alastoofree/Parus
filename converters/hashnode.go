// parus/converters/hashnode.go

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
	"time")

const hashnodeAPIURL = "https://gql.hashnode.com/"

// hnQuery выполняет GraphQL запрос к Hashnode API
func hnQuery(query string, variables map[string]interface{}) (map[string]interface{}, error) {
body, err := json.Marshal(map[string]interface{}{
	"query":     query,
	"variables": variables,})
if err != nil {return nil, err}
resp, err := http.Post(hashnodeAPIURL, "application/json", bytes.NewReader(body))
if err != nil {return nil, err}
defer resp.Body.Close()
raw, err := io.ReadAll(resp.Body)
if err != nil {return nil, err}
var result map[string]interface{}
if err := json.Unmarshal(raw, &result); err != nil {return nil, err}
if errs, ok := result["errors"]; ok {
	return nil, fmt.Errorf("GraphQL ошибка: %v", errs)}
data, ok := result["data"].(map[string]interface{})
if !ok {return nil, fmt.Errorf("нет данных в ответе")}
return data, nil }
// 
func convertHashnode(pageURL string) ([]ConvertPart, error) {
parsed, err := url.Parse(pageURL)
if err != nil {return nil, fmt.Errorf("некорректный URL: %w", err)}
host        := parsed.Host
blogName    := Slugify(host)
importTag   := "hashnode-" + blogName
postSlug    := hnExtractSlug(parsed.Path)
var parts []ConvertPart
tagNames := []string{importTag, "hashnode-пост", "hashnode-комментарий", "hashnode-import"}
uniqueTags := make(map[string]bool)
for _, tn := range tagNames {
	if !uniqueTags[tn] {
		parts = append(parts, ConvertPart{Title: tn, IsTag: true})
		uniqueTags[tn] = true}}
if postSlug != "" {
	log.Printf("Hashnode: один пост, slug: %s", postSlug)
	p, err := hnFetchSinglePost(host, postSlug, importTag, uniqueTags)
	if err != nil {return nil, err}
	parts = append(parts, p...)
} else {
	log.Printf("Hashnode: весь блог, хост: %s", host)
	p, err := hnFetchAllPosts(host, importTag, uniqueTags)
	if err != nil {return nil, err}
	parts = append(parts, p...)}
log.Printf("Hashnode: импорт завершён, всего частей: %d", len(parts))
return parts, nil }
// hnFetchSinglePost загружает один пост по slug
func hnFetchSinglePost(host, slug, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
	q := `query($host: String!, $slug: String!) {
		publication(host: $host) {
			post(slug: $slug) {
				title slug
				content { markdown }
				publishedAt
				tags { name slug }
				comments(first: 50) {
					edges { node {
						author { name }
						content { text }
						dateAdded
						replies(first: 50) {
							edges { node {
								author { name }
								content { text }
								dateAdded}}}}}}}}}`
data, err := hnQuery(q, map[string]interface{}{"host": host, "slug": slug})
if err != nil {return nil, err}
pub, _ := data["publication"].(map[string]interface{})
if pub == nil {return nil, fmt.Errorf("публикация не найдена")}
post, _ := pub["post"].(map[string]interface{})
if post == nil {return nil, fmt.Errorf("пост не найден")}
return hnConvertPost(post, host, importTag, uniqueTags), nil }
// hnFetchAllPosts загружает все посты блога
func hnFetchAllPosts(host, importTag string, uniqueTags map[string]bool) ([]ConvertPart, error) {
var parts []ConvertPart
cursor := ""
for {
	q := `query($host: String!, $first: Int!, $after: String) {
		publication(host: $host) {
			posts(first: $first, after: $after) {
				pageInfo { endCursor hasNextPage }
				edges { node {
					title slug
					content { markdown }
					publishedAt
					tags { name slug }
					comments(first: 50) {
						edges { node {
							author { name }
							content { text }
							dateAdded
							replies(first: 50) {
								edges { node {
									author { name }
									content { text }
									dateAdded}}}}}}}}}}}`
	vars := map[string]interface{}{"host": host, "first": 20}
	if cursor != "" {vars["after"] = cursor}
	data, err := hnQuery(q, vars)
	if err != nil {return nil, err}
	pub, _ := data["publication"].(map[string]interface{})
	if pub == nil {break}
	postsData, _ := pub["posts"].(map[string]interface{})
	if postsData == nil {break}
	edges, _ := postsData["edges"].([]interface{})
	for _, edge := range edges {
		edgeMap, _ := edge.(map[string]interface{})
		if edgeMap == nil {continue}
		node, _ := edgeMap["node"].(map[string]interface{})
		if node == nil {continue}
		parts = append(parts, hnConvertPost(node, host, importTag, uniqueTags)...)}
	pageInfo, _ := postsData["pageInfo"].(map[string]interface{})
	hasNext, _ := pageInfo["hasNextPage"].(bool)
	if !hasNext {break}
	cursor, _ = pageInfo["endCursor"].(string)
	log.Printf("Hashnode: загружено %d постов", len(parts))
	time.Sleep(250 * time.Millisecond)}
return parts, nil }
// hnConvertPost конвертирует один пост из GraphQL ответа
func hnConvertPost(post map[string]interface{}, host, importTag string, uniqueTags map[string]bool) []ConvertPart {
var parts []ConvertPart
title, _  := post["title"].(string)
slug, _   := post["slug"].(string)
sourceURL := fmt.Sprintf("https://%s/%s", host, slug)
// Контент из markdown
content := ""
if contentMap, ok := post["content"].(map[string]interface{}); ok {
	md, _ := contentMap["markdown"].(string)
	content = md}
content += fmt.Sprintf(`<p><br><i>Источник: <a href="%s" target="_blank" rel="noopener noreferrer">%s</a></i></p>`, sourceURL, sourceURL)
postTagName := importTag + "-" + Slugify(title)
if !uniqueTags[postTagName] {
	parts = append(parts, ConvertPart{Title: postTagName, IsTag: true})
	uniqueTags[postTagName] = true}
// Теги поста
if tags, ok := post["tags"].([]interface{}); ok {
	for _, t := range tags {
		if tagMap, ok := t.(map[string]interface{}); ok {
			name, _ := tagMap["name"].(string)
			labelTag := "hn-tag-" + hnSanitizeTag(name)
			if !uniqueTags[labelTag] {
				parts = append(parts, ConvertPart{Title: labelTag, IsTag: true})
				uniqueTags[labelTag] = true}}}}
// Комментарии
var commentEdges []interface{}
if commentsMap, ok := post["comments"].(map[string]interface{}); ok {
	commentEdges, _ = commentsMap["edges"].([]interface{})}
// Строим карту заголовков комментариев
type hnCommentInfo struct {
	title  string
	author string
	date   string
	text   string}
var commentInfos []hnCommentInfo
type hnReplyInfo struct {
	title       string
	parentTitle string
	author      string
	date        string
	text        string}
var replyInfos []hnReplyInfo
for _, edge := range commentEdges {
	edgeMap, _ := edge.(map[string]interface{})
	if edgeMap == nil {continue}
	node, _ := edgeMap["node"].(map[string]interface{})
	if node == nil {continue}
	author := ""
	if authorMap, ok := node["author"].(map[string]interface{}); ok {
		author, _ = authorMap["name"].(string)}
	date := ""
	if d, ok := node["dateAdded"].(string); ok && len(d) >= 10 {date = d[:10]}
	text := ""
	if contentMap, ok := node["content"].(map[string]interface{}); ok {
		text, _ = contentMap["text"].(string)}
	commentTitle := postTagName + "-comment-" + Slugify(author) + "-" + date
	commentInfos = append(commentInfos, hnCommentInfo{title: commentTitle, author: author, date: date, text: text})
	// Ответы
	if repliesMap, ok := node["replies"].(map[string]interface{}); ok {
		if replyEdges, ok := repliesMap["edges"].([]interface{}); ok {
			for _, re := range replyEdges {
				reMap, _ := re.(map[string]interface{})
				if reMap == nil {continue}
				rNode, _ := reMap["node"].(map[string]interface{})
				if rNode == nil {continue}
				rAuthor := ""
				if authorMap, ok := rNode["author"].(map[string]interface{}); ok {
					rAuthor, _ = authorMap["name"].(string)}
				rDate := ""
				if d, ok := rNode["dateAdded"].(string); ok && len(d) >= 10 {rDate = d[:10]}
				rText := ""
				if contentMap, ok := rNode["content"].(map[string]interface{}); ok {
					rText, _ = contentMap["text"].(string)}
				replyTitle := postTagName + "-reply-" + Slugify(rAuthor) + "-" + rDate
				replyInfos = append(replyInfos, hnReplyInfo{
					title: replyTitle, parentTitle: commentTitle,
					author: rAuthor, date: rDate, text: rText})}}}}
// Ссылки на комментарии первого уровня в посте
var level1Links strings.Builder
for _, ci := range commentInfos {
	level1Links.WriteString(fmt.Sprintf("[[%s]] ", ci.title))}
if level1Links.Len() > 0 {
	content += "<p><b>Комментарии:</b> " + level1Links.String() + "</p>"}
parts = append(parts, ConvertPart{
	Title:   title,
	Content: strings.ReplaceAll(strings.TrimSpace(content), "\n", " "),
	TagName: postTagName,
	IsHTML:  false,})
// childMap для ссылок на ответы
childMap := make(map[string][]string)
for _, ri := range replyInfos {
	childMap[ri.parentTitle] = append(childMap[ri.parentTitle], ri.title)}
// Фреймы комментариев
for _, ci := range commentInfos {
	var b strings.Builder
	if ci.author != "" {b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", ci.author))}
	if ci.date != "" {b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", ci.date))}
	b.WriteString("<hr>")
	b.WriteString(ci.text)
	if children, ok := childMap[ci.title]; ok {
		b.WriteString("<p><b>Ответы:</b> ")
		for _, childTitle := range children {
			b.WriteString(fmt.Sprintf("[[%s]] ", childTitle))}
		b.WriteString("</p>")}
	parts = append(parts, ConvertPart{
		Title:   ci.title,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: postTagName,
		Rank:    "0",
		IsHTML:  true,})}
// Фреймы ответов
for _, ri := range replyInfos {
	var b strings.Builder
	if ri.author != "" {b.WriteString(fmt.Sprintf("<p><b>Автор:</b> %s</p>", ri.author))}
	if ri.date != "" {b.WriteString(fmt.Sprintf("<p><b>Дата:</b> %s</p>", ri.date))}
	b.WriteString("<hr>")
	b.WriteString(ri.text)
	parts = append(parts, ConvertPart{
		Title:   ri.title,
		Content: strings.ReplaceAll(strings.TrimSpace(b.String()), "\n", " "),
		TagName: ri.parentTitle,
		Rank:    "0",
		IsHTML:  true,})}
return parts }
// 
func hnSanitizeTag(tag string) string {
tag = strings.ReplaceAll(tag, " ", "-")
reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
return reg.ReplaceAllString(tag, "")}
// 
func hnExtractSlug(path string) string {
path = strings.Trim(path, "/")
if path == "" {return ""}
parts := strings.Split(path, "/")
for i := len(parts) - 1; i >= 0; i-- {
	if parts[i] != "" {return parts[i]}}
return ""}