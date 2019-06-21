package main

import (
  "net/http"
)

func test(w http.ResponseWriter, r *http.Request) {
  enableCors(&w)
  w.Write([]byte("OK"))
}

func main() {
  http.HandleFunc("/test", test)

  if err := http.ListenAndServe(":8080", nil); err != nil {
    panic(err)
  }
}

func enableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
}
