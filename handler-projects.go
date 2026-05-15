// parus/handler-projects.go

package main

import (
	"net/http"
	"parus/unison")

func handleProjects(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
	projects, err := unison.ListProjects()
	if err != nil {writeError(w, 500, err.Error()); return}
	type projectJSON struct {
		ID    string `json:"id"`
		Title string `json:"title"`}
	result := make([]projectJSON, 0, len(projects))
	for _, p := range projects {
    title := ""
    for _, pair := range p.Meta.Frame.Pairs {
        if pair.Left == "title" {title = pair.Right; break}}
    if title == "" {
        if p.ID == unison.MetaBaseName {title = "META-BASE"} else {title = p.ID}}
    result = append(result, projectJSON{ID: p.ID, Title: title})}
		writeJSON(w, result)
case http.MethodPost:
    var body struct {
        Title        string `json:"title"`
        FileOrderLen int    `json:"fileOrderLen"`}
    body.FileOrderLen = 3
    if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
    if body.Title == "" {body.Title = "Not title"}
    p, err := unison.CreateProject(body.Title, body.FileOrderLen)
	if err != nil {writeError(w, 500, err.Error()); return}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]string{"id": p.ID, "title": body.Title})
default: writeError(w, 405, "method not allowed")} }
// 
func handleOneProject(w http.ResponseWriter, r *http.Request, projectID string) {
switch r.Method {
case http.MethodGet:
	p, err := unison.LoadProject(projectID)
	if err != nil {writeError(w, 404, "project not found"); return}
	writeJSON(w, map[string]any{
    "id":           p.ID,
    "fileOrderLen": p.Meta.FileOrderLen,
    "pairs":        p.Meta.Frame.Pairs,})
case http.MethodPut:
    var body struct {
	    Title  string `json:"title"`
	    ASort  string `json:"aSort"`
	    OLens  string `json:"oLens"`
	    Ranks  string `json:"ranks"`
	    Tabs   string `json:"tabs"`
	    West   string `json:"west"`
	    Nord   string `json:"nord"`
        Graph  string `json:"graph"`    
        List   string `json:"list"`
        Chrono string `json:"chrono"`
        DeepNord  string `json:"deepNord"`
        DeepSouth string `json:"deepSouth"`}
    if err := decodeJSON(r, &body);
    err != nil {writeError(w, 400, "invalid json"); return}
    p, err := unison.LoadProject(projectID)
    if err != nil {writeError(w, 404, "project not found"); return}
    _, err = unison.UpdateProject(projectID, p, unison.ProjectConfig{
        Title:     body.Title,
        ASort:     body.ASort,
        OLens:     body.OLens,
        Ranks:     body.Ranks,
        Tabs:      body.Tabs,
        West:      body.West,
        Nord:      body.Nord,
        Graph:     body.Graph,
        List:      body.List,
        Chrono:    body.Chrono,
        DeepNord:  body.DeepNord,
        DeepSouth: body.DeepSouth,})
    if err != nil {writeError(w, 500, err.Error()); return}
    go func() {_ = unison.BuildGlobalIndex()}()
	writeJSON(w, map[string]string{"id": projectID, "title": body.Title})
default: writeError(w, 405, "method not allowed")} }