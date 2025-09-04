package main

import (
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsRetrieveById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpId")
	chirp, err := cfg.db.GetChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp with provided id doesn't exist", err)
	}
	chirpStruct := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, chirpStruct)
}
