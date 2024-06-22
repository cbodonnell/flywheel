package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/cbodonnell/flywheel/pkg/api/middleware"
	"github.com/cbodonnell/flywheel/pkg/log"
	"github.com/cbodonnell/flywheel/pkg/repositories"
	"github.com/cbodonnell/flywheel/pkg/repositories/models"
)

func HandleListCharacters(repository repositories.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			log.Error("failed to get user from context")
			http.Error(w, "Failed to get user from context", http.StatusInternalServerError)
			return
		}
		characters, err := repository.ListCharacters(r.Context(), user.ID)
		if err != nil {
			log.Error("failed to list characters: %v", err)
			http.Error(w, "Failed to list characters", http.StatusInternalServerError)
			return
		}

		log.Debug("listing characters with CORS!!")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(characters); err != nil {
			log.Error("failed to encode characters: %v", err)
			http.Error(w, "Failed to encode characters", http.StatusInternalServerError)
			return
		}
	}
}

func HandleCreateCharacter(repository repositories.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			log.Error("failed to get user from context")
			http.Error(w, "Failed to get user from context", http.StatusInternalServerError)
			return
		}

		name := r.FormValue("name")

		if len(name) < 1 || len(name) > 16 {
			http.Error(w, "Name must be between 1 and 16 characters", http.StatusBadRequest)
			return
		}

		nameRegex := regexp.MustCompile(`^[a-zA-Z0-9 ]+$`)
		if !nameRegex.MatchString(name) {
			http.Error(w, "Name cannot contain special characters", http.StatusBadRequest)
			return
		}

		nameExists, err := repository.NameExists(r.Context(), name)
		if err != nil {
			log.Error("failed to check if name exists: %v", err)
			http.Error(w, "Failed to check if name exists", http.StatusInternalServerError)
			return
		}

		if nameExists {
			http.Error(w, "Name already exists", http.StatusBadRequest)
			return
		}

		count, err := repository.CountCharacters(r.Context(), user.ID)
		if err != nil {
			log.Error("failed to count characters: %v", err)
			http.Error(w, "Failed to count characters", http.StatusInternalServerError)
			return
		}

		if count >= 3 {
			http.Error(w, "Character limit reached", http.StatusBadRequest)
			return
		}

		character, err := repository.CreateCharacter(r.Context(), user.ID, name)
		if err != nil {
			if repositories.IsNameExists(err) {
				http.Error(w, "Name already exists", http.StatusBadRequest)
				return
			}
			log.Error("failed to create character: %v", err)
			http.Error(w, "Failed to create character", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(character); err != nil {
			log.Error("failed to encode character: %v", err)
			http.Error(w, "Failed to encode character", http.StatusInternalServerError)
			return
		}
	}
}

func HandleDeleteCharacter(repository repositories.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*models.User)
		if !ok {
			log.Error("failed to get user from context")
			http.Error(w, "Failed to get user from context", http.StatusInternalServerError)
			return
		}
		characterID, err := strconv.Atoi(r.PathValue("characterID"))
		if err != nil {
			log.Error("failed to parse characterID: %v", err)
			http.Error(w, "Failed to parse characterID", http.StatusBadRequest)
			return
		}

		err = repository.DeleteCharacter(r.Context(), user.ID, int32(characterID))
		if err != nil {
			if repositories.IsNotFound(err) {
				http.Error(w, "Character not found", http.StatusNotFound)
				return
			}
			log.Error("failed to delete character: %v", err)
			http.Error(w, "Failed to delete character", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
	}
}
