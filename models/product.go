package models

import (
	"log"
)

// Product ... We will first create a new type called Product
// This type will contain information about VR experiences
type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// AllProducts ...
func AllProducts() ([]*Product, error) {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*Product, 0)
	for rows.Next() {
		pr := new(Product)
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.Slug, &pr.Description); err != nil {
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
