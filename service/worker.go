package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func GenerateAndWriteVector(ctx context.Context, vGenerator *VectorGenerator, db *sql.DB, n Neighborhood) error {
	text := fmt.Sprintf("%s %s %s", n.Province, n.District, n.Name)
	vec, err := vGenerator.Generate(ctx, strings.Join(EdgeNGrams(text, 1, 10), " "))
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO neighborhoods (id, province_id, district_id, province, district, name, embedding) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		n.ID, n.ProvinceID, n.DistrictID, n.Province, n.District, n.Name, formatVector(Normalize(vec)))
	if err != nil {
		log.Err(err).Send()
		return fmt.Errorf("Failed to insert neighborhood %d: %v", n.ID, err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Inserted neighborhood %d, rows affected: %d", n.ID, rowsAffected)

	return nil
}
