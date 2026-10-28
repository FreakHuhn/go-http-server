package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request){
	hits := cfg.fileserverHits.Load()
	fmt.Fprintf(w, "Fileserver Hits: %d\n", hits)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
	cfg.fileserverHits.Store(0)
}

func main(){
	apicfg := &apiConfig{}
	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apicfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apicfg.handlerReset)
	server := &http.Server{
		Addr : ":8080",
		Handler: mux,}
		server.ListenAndServe()
}
