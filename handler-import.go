// parus/handler-import.go

package main

import (
	"net/http"
	"parus/unison"
    "parus/converters")

func handleImport(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
var body struct {
	From   string   `json:"from"`
	Frames []string `json:"frames"`}
if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
if body.From == "" || len(body.Frames) == 0 {writeError(w, 400, "from and frames required"); return}
count, err := unison.ImportFrames(body.From, projectID, body.Frames)  
if err != nil {writeError(w, 500, err.Error()); return}
rebuildAfterWrite(projectID)
go func() {_ = unison.BuildGlobalIndexForBase(projectID)}()
writeJSON(w, map[string]int{"imported": count}) }
// 
func handleExport(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
var body struct {
	To     string   `json:"to"`
	Frames []string `json:"frames"`}
if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
if body.To == "" || len(body.Frames) == 0 {writeError(w, 400, "to and frames required"); return}
count, err := unison.ImportFrames(projectID, body.To, body.Frames)
if err != nil {writeError(w, 500, err.Error()); return}
rebuildAfterWrite(body.To)
go func() {_ = unison.BuildGlobalIndex()}()
writeJSON(w, map[string]int{"exported": count}) }
// 
func handleConvert(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
var body struct {URL string `json:"url"`}
if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
if body.URL == "" {writeError(w, 400, "url required"); return}
p, err := unison.LoadProject(projectID)
if err != nil {writeError(w, 404, "project not found"); return}
parts, err := converters.Convert(body.URL)
if err != nil {writeError(w, 500, err.Error()); return}
count, err := converters.InjectParts(projectID, p.Meta, parts)
if err != nil {writeError(w, 500, err.Error()); return}
rebuildAfterWrite(projectID)
writeJSON(w, map[string]int{"imported": count}) }
// 
func handleUnion(w http.ResponseWriter, r *http.Request, projectID string) {
if r.Method != http.MethodPost {writeError(w, 405, "method not allowed"); return}
var body struct {Index string `json:"index"`}
if err := decodeJSON(r, &body); err != nil {writeError(w, 400, "invalid json"); return}
if body.Index == "" {writeError(w, 400, "index required"); return}
if err := unison.UnionFrames(projectID, body.Index); err != nil {writeError(w, 500, err.Error()); return}
rebuildAfterWrite(projectID)
writeJSON(w, map[string]string{"status": "ok"}) }