package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dalais/sdku_backend/components"
	"github.com/dalais/sdku_backend/models"
	"github.com/golang/gddo/httputil/header"
)

// Register ...
func Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var answer components.RAnswer
		if r.Header.Get("Content-Type") != "" {
			value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
			if value != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var u models.User
		err := dec.Decode(&u)
		if err != nil {
			var syntaxError *json.SyntaxError
			var unmarshalTypeError *json.UnmarshalTypeError

			switch {
			// Catch any syntax errors in the JSON and send an error message
			// which interpolates the location of the problem to make it
			// easier for the client to fix.
			case errors.As(err, &syntaxError):
				msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
				http.Error(w, msg, http.StatusBadRequest)

			// In some circumstances Decode() may also return an
			// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
			// is an open issue regarding this at
			// https://github.com/golang/go/issues/25956.
			case errors.Is(err, io.ErrUnexpectedEOF):
				msg := fmt.Sprintf("Request body contains badly-formed JSON")
				http.Error(w, msg, http.StatusBadRequest)

			// Catch any type errors, like trying to assign a string in the
			// JSON request body to a int field in our Person struct. We can
			// interpolate the relevant field name and position into the error
			// message to make it easier for the client to fix.
			case errors.As(err, &unmarshalTypeError):
				msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
				http.Error(w, msg, http.StatusBadRequest)

			// Catch the error caused by extra unexpected fields in the request
			// body. We extract the field name from the error message and
			// interpolate it in our custom error message. There is an open
			// issue at https://github.com/golang/go/issues/29035 regarding
			// turning this into a sentinel error.
			case strings.HasPrefix(err.Error(), "json: unknown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
				msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
				http.Error(w, msg, http.StatusBadRequest)

			// An io.EOF error is returned by Decode() if the request body is
			// empty.
			case errors.Is(err, io.EOF):
				msg := "Request body must not be empty"
				http.Error(w, msg, http.StatusBadRequest)

			// Catch the error caused by the request body being too large. Again
			// there is an open issue regarding turning this into a sentinel
			// error at https://github.com/golang/go/issues/30715.
			case err.Error() == "http: request body too large":
				msg := "Request body must not be larger than 1MB"
				http.Error(w, msg, http.StatusRequestEntityTooLarge)

			// Otherwise default to logging the error and sending a 500 Internal
			// Server Error response.
			default:
				log.Println(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
		err = dec.Decode(&struct{}{})
		if err != io.EOF {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		fields, _ := json.Marshal(u)
		answer.Data = string(fields)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(answer)
	})
}
