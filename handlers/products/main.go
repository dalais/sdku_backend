package producthandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dalais/sdku_backend/models"
)

func main() {
}

// Index ...
func Index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		products, err := models.AllProducts()
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})
}
