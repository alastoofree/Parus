// parus/main.go

package main

import ("fmt"
        "net/http"
        "parus/unison" )

func main() {
if err := unison.InitMetaBase(); 
    err != nil {fmt.Println("Warning: meta-base init failed:", err)}
registerRoutes()
const port = "3888"    
fmt.Println("Open PARUS at http://localhost:" + port + "/portal.html")
if err := http.ListenAndServe(":"+port, nil); 
    err != nil {fmt.Println("Error:", err)} }