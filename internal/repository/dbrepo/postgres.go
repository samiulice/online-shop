package dbrepo

import (
	"context"
	"online_store/internal/models"
	"time"
)

// GetDate return a date package for specific id
func (m *postgresDBRepo) GetDate(id int) (models.Date, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var d models.Date

	query := `
		select 
			id, name, description, package_size, package_weight, package_price, stock_level, coalesce(image, ''), created_at, updated_at
		from 
			dates
		where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.PackageSize,
		&d.PackageWeight,
		&d.PackagePrice,
		&d.StockLevel,
		&d.Image,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	return d, err
}
