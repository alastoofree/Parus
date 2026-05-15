// parus/response.go

package main

import ( "encoding/json" 
        "net/http" )

func writeJSON(w http.ResponseWriter, v any) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(v) }
// 
func decodeJSON(r *http.Request, v any) error {; return json.NewDecoder(r.Body).Decode(v) }
// 
func writeError(w http.ResponseWriter, status int, msg string) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(status)
json.NewEncoder(w).Encode(map[string]string{"message": msg}) }
// 
func noCache(h http.Handler) http.Handler {;return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Cache-Control", "no-store")
h.ServeHTTP(w, r)}) }