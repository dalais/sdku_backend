package product

// Product ... We will first create a new type called Product
// This type will contain information about VR experiences
type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}
