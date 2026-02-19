package service

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"unicode"
)

func formatVector(v []float64) string {
	strValues := make([]string, len(v))
	for i, val := range v {
		strValues[i] = fmt.Sprintf("%f", val)
	}
	return "[" + strings.Join(strValues, ",") + "]"
}

func findNearestNeighborhoods(db *sql.DB, targetVector []float64, limit int) ([]NeighborhoodWithScore, error) {
	// Convert slice to pgvector string format "[0.1,0.2...]"
	queryStr := formatVector(Normalize(targetVector))

	// We use <=> for Cosine Distance (common in AI/Embeddings)
	query := `
		SELECT id, name, district, province, (embedding <=> $1) as score
		FROM neighborhoods
		ORDER BY score ASC
		LIMIT $2;`

	rows, err := db.Query(query, queryStr, limit)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var ret []NeighborhoodWithScore
	for rows.Next() {
		var id int
		var name, district, province string
		var score float64
		if err := rows.Scan(&id, &name, &district, &province, &score); err != nil {
			return ret, fmt.Errorf("scan: %w", err)
		}
		ret = append(ret, NeighborhoodWithScore{
			Neighborhood: Neighborhood{
				ID:       id,
				Name:     name,
				District: district,
				Province: province,
			},
			Score: score,
		})
	}

	return ret, nil
}

func Normalize(v []float64) []float64 {
	var sum float64
	for _, val := range v {
		sum += val * val
	}

	magnitude := math.Sqrt(sum)

	// Avoid division by zero for empty or zero vectors
	if magnitude == 0 {
		return v
	}

	normalized := make([]float64, len(v))
	for i, val := range v {
		normalized[i] = val / magnitude
	}

	return normalized
}

func GenerateNGrams(input string, n int) []string {
	// Clean the input: lowercase and remove extra spaces
	input = strings.ToLower(strings.TrimSpace(input))

	// If the string is shorter than N, just return the string in a slice
	if len(input) < n {
		return []string{input}
	}

	var ngrams []string
	// Iterate through the string using runes to support UTF-8 characters
	runes := []rune(input)
	for i := 0; i <= len(runes)-n; i++ {
		ngrams = append(ngrams, string(runes[i:i+n]))
	}
	return ngrams
}

func EdgeNGrams(input string, min, max int) []string {
	// 1. Normalize: lowercase and split by non-alphanumeric characters
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	words := strings.FieldsFunc(strings.ToLower(input), f)

	var allGrams []string
	seen := make(map[string]bool)

	for _, word := range words {
		runes := []rune(word)
		length := len(runes)

		// Generate prefixes from min length up to max (or word length)
		for i := min; i <= max && i <= length; i++ {
			gram := string(runes[0:i])
			if !seen[gram] {
				allGrams = append(allGrams, gram)
				seen[gram] = true
			}
		}
	}
	return allGrams
}
