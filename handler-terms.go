// parus/handler-terms.go

package main

import (
	"net/http"
	"strings"
	"parus/unison"
	terms "parus/unison/terms")

func handleTerms(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
	tlist, err := terms.ListTerms()
	if err != nil {writeError(w, 500, err.Error()); return}
	type termJSON struct {
		Path string `json:"path"`
		Name string `json:"name"`
        Rank string `json:"rank"`
        Uri  string `json:"uri"`
		Type string `json:"type"`
		Icon  string `json:"icon"`
        Color string `json:"color"`
        QBtns string `json:"qBtns"`}
	result := make([]termJSON, 0, len(tlist))
	for _, t := range tlist {
		result = append(result, termJSON{
			Path: t.Path,
			Name: terms.TermPairRight(t, "name"),
            Rank: terms.TermPairRight(t, "rank"),
            Uri:  terms.TermPairRight(t, "uri"),
			Type: terms.TermPairRight(t, "type"),
			Icon: terms.TermPairRight(t, "icon"),
            Color: terms.TermPairRight(t, "color"),
            QBtns: terms.TermPairRight(t, "q-btns"),})}
	writeJSON(w, result)
case http.MethodPost:
	var body struct {
    Name       string `json:"name"`
    SourcePath string `json:"sourcePath"`
    TargetUri  string `json:"targetUri"`}
if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
if body.Name == "" {writeError(w, 400, "name required"); return}
var t terms.TermFrame
var err error
if body.SourcePath != "" {
    t, err = terms.CloneTerm(body.SourcePath, body.Name, body.TargetUri)
} else {
    t, err = terms.CreateTerm(body.Name)}
if err != nil {writeError(w, 500, err.Error()); return}
w.WriteHeader(http.StatusCreated)
writeJSON(w, map[string]string{"path": t.Path, "name": terms.TermPairRight(t, "name")})
default:
	writeError(w, 405, "method not allowed")} }
// 
func handleOneTerm(w http.ResponseWriter, r *http.Request, termPath string) {
switch r.Method {
case http.MethodGet:
	t, err := terms.LoadTerm(termPath)
	if err != nil {writeError(w, 404, "term not found"); return}
	writeJSON(w, map[string]any{
		"path":  t.Path,
		"pairs": t.Pairs,})
case http.MethodPut:
	var body struct {
		Pairs []unison.SlotPair `json:"pairs"`}
	if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
	t, err := terms.LoadTerm(termPath)
	if err != nil {writeError(w, 404, "term not found"); return}
	t.Pairs = body.Pairs
    t.Pairs = terms.SyncTermContentLinks(t.Pairs)
	if err := terms.WriteTerm(t); err != nil {writeError(w, 500, err.Error()); return}
    go func() {_ = terms.BuildTermIndex(termPath)}()
	writeJSON(w, map[string]string{"path": termPath})
    newAliases := []string{}
    for _, p := range body.Pairs {
	    if p.Left == "alias" {newAliases = strings.Fields(p.Right); break}}
    _ = terms.SyncTermAliasGroup(termPath, newAliases)
default:
	writeError(w, 405, "method not allowed")} }
// 
func handleTermsWithPath(w http.ResponseWriter, r *http.Request) {
path := strings.TrimPrefix(r.URL.Path, "/api/terms/")
if path == "" {handleTerms(w, r); return}
if path == "children" {
	termPath := r.URL.Query().Get("path")
	iSort    := r.URL.Query().Get("isort")
	aSort    := r.URL.Query().Get("asort")
	if termPath == "" {writeError(w, 400, "path required"); return}
	children, err := terms.GetTermChildren(termPath, iSort, aSort)
	if err != nil {writeError(w, 500, err.Error()); return}
	writeJSON(w, children); return}
if path == "tags" {
	tlist, err := terms.ListTerms()
	if err != nil {writeError(w, 500, err.Error()); return}
	type tagJSON struct {
		Index string `json:"index"`
		Title string `json:"title"`
		Color string `json:"color"`
		Icon  string `json:"icon"`}
	result := make([]tagJSON, 0, len(tlist))
	for _, t := range tlist {
    rank := terms.TermPairRight(t, "rank")
    if rank == "" {continue}
    result = append(result, tagJSON{
        Index: t.Path,
        Title: terms.TermPairRight(t, "name"),
        Color: terms.TermPairRight(t, "color"),
        Icon:  terms.TermPairRight(t, "icon"),})}
	writeJSON(w, result); return}
if path == "content" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    content, raw, err := terms.ResolveTermContent(termPath)
    if err != nil {writeError(w, 500, err.Error()); return}
    writeJSON(w, map[string]string{
        "content":     content,
        "contentRaw":  raw,
        "contentType": "markdown",}); return}
if path == "satellite" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    switch r.Method {
    case http.MethodGet:
    tp := termPath
    if !strings.HasSuffix(tp, ".term") {
        if found, ok := terms.FindTermByName(tp); ok {tp = found} else {
            writeJSON(w, map[string]string{}); return}}
    data, err := terms.LoadTermSatellite(tp)
    if err != nil {writeJSON(w, map[string]string{}); return}
    writeJSON(w, data)
    case http.MethodPut:
        var body map[string]string
        if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
        if err := terms.WriteTermSatellite(termPath, body); err != nil {writeError(w, 500, err.Error()); return}
        writeJSON(w, map[string]string{"path": termPath})
    default:
        writeError(w, 405, "method not allowed")}; return}
if path == "satellite-svg" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    switch r.Method {
    case http.MethodGet:
    tp := termPath
    if !strings.HasSuffix(tp, ".term") {
        if found, ok := terms.FindTermByName(tp); ok {tp = found} else {
            writeJSON(w, map[string]string{"svg": ""}); return}}
    data, err := terms.LoadTermSatelliteSvg(tp)
    if err != nil {writeJSON(w, map[string]string{"svg": ""}); return}
    writeJSON(w, map[string]string{"svg": data})
    case http.MethodPut:
        var body struct {Svg string `json:"svg"`}
        if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
        if err := terms.WriteTermSatelliteSvg(termPath, body.Svg); err != nil {writeError(w, 500, err.Error()); return}
        writeJSON(w, map[string]string{"path": termPath})
    default:
        writeError(w, 405, "method not allowed")}; return}
if path == "chrono" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    data, err := terms.GetTermChrono(termPath)
    if err != nil {writeError(w, 500, err.Error()); return}
    writeJSON(w, data); return}
if path == "process" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    data, err := terms.GetTermProcess(termPath)
    if err != nil {writeError(w, 500, err.Error()); return}
    writeJSON(w, data); return}
if path == "product" {
    termPath := r.URL.Query().Get("path")
    if termPath == "" {writeError(w, 400, "path required"); return}
    data, err := terms.ResolveTermProduct(termPath)
    if err != nil {writeError(w, 500, err.Error()); return}
    writeJSON(w, data); return}
if path == "search" {
    query := r.URL.Query().Get("q")
    if query == "" {writeError(w, 400, "q required"); return}
    data, err := terms.SearchTerms(query)
    if err != nil {writeError(w, 500, err.Error()); return}
    writeJSON(w, data); return}
handleOneTerm(w, r, path) }
//
func handleTermGraph(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
path := r.URL.Query().Get("path")
if path == "" {writeError(w, 400, "path required"); return}
data, err := terms.GetTermGraph(path)
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, data) }