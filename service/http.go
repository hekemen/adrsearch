package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type NeighborhoodWithScore struct {
	Neighborhood
	Score float64 `json:"score"`
}

func CreateAndStartHttp(ctx context.Context, vGenerator *VectorGenerator, db *sql.DB) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		queryTerm := r.URL.Query().Get("q")
		if queryTerm == "" {
			http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
			return
		}

		vec, err := vGenerator.Generate(ctx, strings.Join(EdgeNGrams(queryTerm, 1, 10), " "))
		if err != nil {
			http.Error(w, fmt.Sprintf("Embedding error: %v", err), http.StatusInternalServerError)
			return
		}

		results, err := findNearestNeighborhoods(db, vec, 10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	log.Info().Msg("🌍 Server started at http://localhost:8081")
	log.Err(http.ListenAndServe(":8081", mux)).Send()
}
