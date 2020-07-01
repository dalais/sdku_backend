package productstore

import (
	"database/sql"
	"log"

	"github.com/dalais/sdku_backend/store"
)

// Product ... We will first create a new type called Product
// This type will contain information about VR experiences
type Product struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	CrtdAt      sql.NullString `json:"crtd_at,omitempty"`
	ChngAt      string         `json:"chng_at,omitempty"`
}

// AllProducts ...
func AllProducts() ([]*Product, error) {
	rows, err := store.Db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*Product, 0)
	for rows.Next() {
		pr := new(Product)
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.Slug, &pr.Description, &pr.CrtdAt, &pr.ChngAt); err != nil {
			log.Printf("Error %d", err)
			return nil, err
		}
		products = append(products, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
