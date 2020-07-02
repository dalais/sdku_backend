package producthandler

import (
	"encoding/json"
	"fmt"
	"net/http"

	productstore "github.com/dalais/sdku_backend/store/product"
)

func main() {
}

// Index - Get all products
func Index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		products, err := productstore.AllProducts()
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	})
}
