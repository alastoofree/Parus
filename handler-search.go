// parus/handlers-search.go

package main

import (
	"net/http"
	"parus/unison")

func handleProjectSearch(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
q := r.URL.Query().Get("q")
results, err := unison.SearchInProject(projectID, q)
if err != nil {writeError(w, 500, err.Error()); return}
if results == nil {results = []map[string]string{}}
writeJSON(w, results) }
// 
func handleFrameSearch(w http.ResponseWriter, r *http.Request, projectID, frameIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
query := r.URL.Query().Get("q")
results := unison.SearchInFrame(projectID, frameIndex, query)
if results == nil {results = []map[string]string{}}
writeJSON(w, results) }
//
func handleGlobalSearch(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
q := r.URL.Query().Get("q")
tags, err := unison.SearchGlobal(q)
if err != nil {writeError(w, 500, err.Error()); return}
if tags == nil {tags = []unison.GlobalFrame{}}
writeJSON(w, tags) }
//
func handleFullSearch(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
q := r.URL.Query().Get("q")
results, err := unison.SearchAllProjects(q)
if err != nil {writeError(w, 500, err.Error()); return}
if results == nil {results = []unison.SearchResult{}}
writeJSON(w, results) }
//
func handleMetaFrameSearch(w http.ResponseWriter, r *http.Request, tagIndex string) {
if r.Method != http.MethodGet {writeError(w, 405, "method not allowed"); return}
results, err := unison.SearchByGlobalTag(tagIndex)
if err != nil {writeError(w, 500, err.Error()); return}
if results == nil {results = []unison.SearchResult{}}
writeJSON(w, results) }
