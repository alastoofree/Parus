// parus/routes.go

package main

import (
	"net/http"
	"strings")

func registerRoutes() {
fs := http.FileServer(http.Dir("./web"))
http.Handle("/", noCache(fs))
http.HandleFunc("/api/projects", handleProjects)
http.HandleFunc("/api/projects/", handleProjectsWithID)
http.HandleFunc("/img/", handleImg)
http.HandleFunc("/api/terms/graph", handleTermGraph)
http.HandleFunc("/api/terms", handleTerms)
http.HandleFunc("/api/terms/", handleTermsWithPath)
http.HandleFunc("/api/global/", handleGlobal) }
// 
func handleProjectsWithID(w http.ResponseWriter, r *http.Request) {
path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
parts := strings.SplitN(path, "/", 4)
projectID := parts[0]
if projectID == "" {writeError(w, 400, "project id required"); return}
if len(parts) == 1 {handleOneProject(w, r, projectID); return}
switch parts[1] {
case "frames": if len(parts) == 2 {handleFrames(w, r, projectID)} else {handleOneFrame(w, r, projectID, parts[2])}
case "children":
	if len(parts) < 3 || parts[2] == "" {writeError(w, 400, "frame index required"); return}
	handleChildren(w, r, projectID, parts[2])
case "graph":
	if len(parts) < 3 || parts[2] == "" {writeError(w, 400, "frame index required"); return}
	handleGraph(w, r, projectID, parts[2])
case "content":
	if len(parts) < 3 || parts[2] == "" {writeError(w, 400, "frame index required"); return}
	handleContent(w, r, projectID, parts[2])
case "chrono":
	if len(parts) < 3 || parts[2] == "" {writeError(w, 400, "frame index required"); return}
	handleChrono(w, r, projectID, parts[2])
case "search":
    if len(parts) == 2 {handleProjectSearch(w, r, projectID); return}
    handleFrameSearch(w, r, projectID, parts[2])
case "tags":    handleTags(w, r, projectID)
case "rebuild": handleRebuildProject(w, r, projectID)
case "import":  handleImport(w, r, projectID)
case "export":  handleExport(w, r, projectID)
case "convert": handleConvert(w, r, projectID)
case "union":   handleUnion(w, r, projectID)
default: writeError(w, 404, "not found")} }
// 
func handleImg(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Cache-Control", "no-store")
path := strings.TrimPrefix(r.URL.Path, "/img/")
parts := strings.SplitN(path, "/", 2)
if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
    http.ServeFile(w, r, "./data/"+parts[0]+"/img/"+parts[1]); return}
http.ServeFile(w, r, "./web/img/"+path) }