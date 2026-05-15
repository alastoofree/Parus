// parus/handler-frames.go

package main

import (
	"strings"
	"net/http"
	"time"
	"parus/unison")

func handleFrames(w http.ResponseWriter, r *http.Request, projectID string) {
p, err := unison.LoadProject(projectID)
if err != nil {writeError(w, 404, "project not found"); return}
switch r.Method {
case http.MethodGet:
	frames, err := unison.ListFrames(projectID, p.Meta)
	if err != nil {writeError(w, 500, err.Error()); return}
	type frameJSON struct {
		Index string `json:"index"`
		Title string `json:"title"`
		IsTag bool   `json:"isTag"`
		Icon  string `json:"icon"`}
	result := make([]frameJSON, 0, len(frames))
	for _, f := range frames {
		title := unison.FramePairRight(f, "title")
        isTag := unison.FramePairRight(f, "rank") != ""
        icon  := unison.FramePairRight(f, "icon")
		result = append(result, frameJSON{
			Index: f.Index,
			Title: title,
			IsTag: isTag,
			Icon:  icon,})}
	writeJSON(w, result)
case http.MethodPost:
	var body struct {
		Title string `json:"title"`}
	if err := decodeJSON(r, &body); 
    err != nil {writeError(w, 400, "invalid json"); return}
	frame, err := unison.CreateFrame(projectID, p.Meta, body.Title)
	if err != nil {writeError(w, 500, err.Error()); return}
//     
	go func() {
	_ = unison.BuildIndex(projectID, frame.Index)
	_ = unison.BuildGlobalIndex()}()
//     
	frameTitle := ""
	for _, p := range frame.Pairs {
		if p.Left == "title" { frameTitle = p.Right; break }}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]string{"index": frame.Index, "title": frameTitle})
default: writeError(w, 405, "method not allowed")} }
// 
func handleOneFrame(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
switch r.Method {
case http.MethodGet:
	frame, err := unison.LoadFrame(projectID, frameIndex)
	if err != nil {writeError(w, 404, "frame not found"); return}
	writeJSON(w, map[string]any{
		"index": frame.Index,
		"pairs": frame.Pairs,})
case http.MethodPut:
	var body struct {
		Pairs       []unison.SlotPair `json:"pairs"`
		AliasBefore []string          `json:"aliasBefore"`}
	if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
	// Читаем старый фрейм ДО записи — для детекции тег→фрейм
	tagBefore := ""
	if oldFrame, err2 := unison.LoadFrame(projectID, frameIndex); err2 == nil {
		for _, p := range oldFrame.Pairs {
			if p.Left == "rank" { tagBefore = p.Right; break }}}
	// Обновляем modified
	now := time.Now().Format("2006-01-02")
	for i, p := range body.Pairs {
		if p.Left == "modified" { body.Pairs[i].Right = now }}
	frame := unison.Frame{
		Index: frameIndex,
		Pairs: body.Pairs,}
	// Проверка уникальности титула и добавление недостающих слотов
    if proj, err2 := unison.LoadProject(projectID); err2 == nil {
        for i, p := range body.Pairs {
            if p.Left == "title" {
                skipTitle := ""
                if oldFrame, err3 := unison.LoadFrame(projectID, frameIndex); err3 == nil {
                    skipTitle = unison.FramePairRight(oldFrame, "title")}
                body.Pairs[i].Right = unison.UniqueTitle(
                    projectID, p.Right, skipTitle, proj.Meta); break}}
        frame = unison.EnsureSlots(frame, []string{"tab", "i-sort", "west", "nord", "graph", "list", "match-nord", "match-south", "chrono", "i-chrono", "dict", "i-dict" , "show-product", "product"})
	// Синхронизируем links и includes из контента
	frame.Pairs = unison.SyncContentLinks(frame.Pairs)
	// Записываем фрейм
	if err := unison.WriteFrame(projectID, frame); 
        err != nil {writeError(w, 500, err.Error()); return}
    if t := unison.FramePairRight(frame, "title"); t != "" {
        _ = unison.UpdateFrameTitle(projectID, frameIndex, t)}
	// Детектируем трансформацию тег→фрейм
	tagAfter := ""
	for _, p := range frame.Pairs {
		if p.Left == "rank" { tagAfter = p.Right; break }}
	if tagBefore != "" && tagAfter == "" {_ = unison.SyncTagRemoval(projectID, frameIndex)}
    if tagBefore == "" && tagAfter != "" {_ = unison.Count(projectID, 0, 1)}
    if tagBefore != "" && tagAfter == "" {_ = unison.Count(projectID, 0, -1)}
	// Обрабатываем алиас-группу
	newAliases := []string{}
    for _, p := range body.Pairs {
        if p.Left == "alias" {newAliases = strings.Fields(p.Right); break}}
    _ = unison.SyncAliasGroup(projectID, frameIndex, newAliases)
	writeJSON(w, map[string]string{"index": frameIndex})
// 
        go func() {
            _ = unison.BuildIndex(projectID, frame.Index)
            if unison.FramePairRight(frame, "rank") != "" {
                _ = unison.BuildGlobalIndexForBase(projectID)}}() }
//     
default: writeError(w, 405, "method not allowed")} }
// 
func handleContent(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
content, contentType, err := unison.ResolveContent(projectID, frameIndex)
if err != nil {writeError(w, 500, err.Error()); return}
raw := ""
if rawFrame, err2 := unison.LoadFrame(projectID, frameIndex); 
    err2 == nil {
	for _, p := range rawFrame.Pairs {
		if p.Left == "content" { raw = p.Right; break }}}
product := ""
if rawFrame, err2 := unison.LoadFrame(projectID, frameIndex); err2 == nil {
	product = unison.ResolveProduct(projectID, rawFrame)}
writeJSON(w, map[string]string{
	"content":     content,
	"contentRaw":  raw,
	"contentType": contentType,
	"product":     product,}) }
// 
func handleGraph(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
data, err := unison.GetGraph(projectID, frameIndex)
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, data) }
// 
func handleTags(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return }
p, err := unison.LoadProject(projectID)
if err != nil {writeError(w, 404, "project not found"); return }
tags, err := unison.ListTags(projectID, p.Meta)
if err != nil {writeError(w, 500, err.Error()); return }
type tagJSON struct {
	Index string `json:"index"`
	Title string `json:"title"`
	Icon string `json:"icon"`
	Color  string `json:"color"`}
result := make([]tagJSON, 0)
for _, f := range tags {
	result = append(result, tagJSON{
		Index: f.Index,
		Title: unison.FramePairRight(f, "title"),
		Icon: unison.FramePairRight(f, "icon"),
		Color:  unison.FramePairRight(f, "color"),})}
writeJSON(w, result) }
// 
func handleChildren(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
p, err := unison.LoadProject(projectID)
if err != nil {writeError(w, 404, "project not found"); return}
iSort := r.URL.Query().Get("isort")
aSort := r.URL.Query().Get("asort")
children, err := unison.GetChildren(projectID, frameIndex, p.Meta, iSort, aSort)
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, children) }
// 
func handleChrono(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
data, err := unison.GetChrono(projectID, frameIndex)
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, data) }