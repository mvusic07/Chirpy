package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/mvusic07/Chirpy/internal/auth"
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

func (cfg *apiConfig) handlerChirpsDeleteById(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	chirpId := r.PathValue("chirpId")
	chirpfind, err := cfg.db.GetChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp with provided id doesn't exist", err)
		return
	}
	if userId != chirpfind.UserID {
		respondWithError(w, http.StatusForbidden, "User not authorized to delete this chirp", err)
		return
	}

	err = cfg.db.DeleteChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while deleting a chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
