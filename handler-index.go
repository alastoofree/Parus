// parus/handler-index.go

package main

import (
	"net/http"
	"strings"
	"parus/unison")

func rebuildAfterWrite(projectID string) {
go func() {
    _ = unison.RebuildIndex(projectID)
    _ = unison.RebuildTagCount(projectID)}() }
// 
func handleRebuildProject(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
if err := unison.RebuildTitles(projectID); 
    err != nil {writeError(w, 500, err.Error()); return}
if err := unison.RebuildIndex(projectID); 
    err != nil {writeError(w, 500, err.Error()); return}
if err := unison.RebuildTagCount(projectID);
    err != nil {writeError(w, 500, err.Error()); return}
if err := unison.BuildGlobalIndex();
    err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, map[string]string{"status": "ok"}) }
//
func handleGlobalRebuild(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
if err := unison.BuildGlobalIndex(); err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, map[string]string{"status": "ok"}) }
// 
func handleGlobal(w http.ResponseWriter, r *http.Request) {
path  := strings.TrimPrefix(r.URL.Path, "/api/global/")
parts := strings.SplitN(path, "/", 4)
switch parts[0] {
case "tags":
    if len(parts) < 2 || parts[1] == "" {handleListMetaFrames(w, r); return}
    tagIndex := parts[1]
    if len(parts) >= 3 && parts[2] == "frames" {handleMetaFrameSearch(w, r, tagIndex); return}
    if len(parts) >= 3 && parts[2] == "graph" {handleMetaGraph(w, r, tagIndex); return}
    writeError(w, 404, "not found")
case "search":     handleGlobalSearch(w, r)
case "fullsearch": handleFullSearch(w, r)
case "rebuild":    handleGlobalRebuild(w, r)
default: writeError(w, 404, "not found")} }
// 
func handleListMetaFrames(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
tags, err := unison.ListMetaFrames()
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, tags) }
//
func handleMetaGraph(w http.ResponseWriter, r *http.Request, tagIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
data, err := unison.GetGraph(unison.MetaBaseName, tagIndex)
if err != nil {writeError(w, 500, err.Error()); return}
writeJSON(w, data) }