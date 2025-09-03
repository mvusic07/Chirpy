package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (apiCfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (apiCfg *apiConfig) handlerCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := apiCfg.fileServerHits.Load()
	response := fmt.Sprintf(`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>`, hits)
	w.Write([]byte(response))
}

func (apiCfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	apiCfg.fileServerHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func cleaned(text string) string {
	profanewords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	words := strings.Split(text, " ")

	for i, word := range words {
		if _, ok := profanewords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")

}

func validateHandler(w http.ResponseWriter, r *http.Request) {

	type parametri struct {
		Tijelo string `json:"body"`
	}
	type greska struct {
		Error string `json:"error"`
	}
	greskainst1 := greska{
		Error: "Something went wrong",
	}
	greskainst2 := greska{
		Error: "Chirp is too long",
	}

	datt, err := json.Marshal(greskainst1)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	dataa, errr := json.Marshal(greskainst2)
	if errr != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parametri{}
	er := decoder.Decode(&params)
	if er != nil {
		w.WriteHeader(500)
		w.Write(datt)
		return
	}
	if len(params.Tijelo) > 140 {
		w.WriteHeader(400)
		w.Write(dataa)
		return
	}

	cleanProfanity := cleaned(params.Tijelo)

	type odgovor struct {
		Cleaned string `json:"cleaned_body"`
	}

	odgovorb := odgovor{
		Cleaned: cleanProfanity,
	}
	dat, err := json.Marshal(odgovorb)
	if err != nil {
		w.WriteHeader(500)
		w.Write(datt)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}

func main() {
	apiCfg := &apiConfig{}
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":08080",
		Handler: mux,
	}
	handler := http.FileServer(http.Dir("."))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(handler)))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type:", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCount)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateHandler)

	server.ListenAndServe()

}
