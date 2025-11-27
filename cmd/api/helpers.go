package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Pedro-J-Kukul/salesapi/internal/data"
	"github.com/Pedro-J-Kukul/salesapi/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// creating an envelope type
type envelope map[string]any

func (a *app) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	// encodes data into json format by using indenting for better readability
	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// append json and add a new line after each appendage
	jsResponse = append(jsResponse, '\n')

	// add any headers that we want to the response
	for key, value := range headers {
		w.Header()[key] = value
	}

	// set content type to header
	w.Header().Set("Content-Type", "application/json")

	// Explicitly setting the response status code
	w.WriteHeader(status)

	// writing the json to the body, but also checking for errors
	_, err = w.Write(jsResponse)
	if err != nil {
		return err
	}

	// returns no error/empty
	return nil
}

func (a *app) readJSON(w http.ResponseWriter, r *http.Request, dest any) error {

	// limit the size of the request body to 256000 bytes
	maxBytes := 256_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// our decoder will check for unknown fields
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// let start the decoding
	err := dec.Decode(dest)

	// Check for different errors
	if err != nil {
		// syntax error
		var syntaxError *json.SyntaxError
		// incorrect type error
		var unmarshalTypeError *json.UnmarshalTypeError
		// empty body error
		var invalidUnmarshalError *json.InvalidUnmarshalError
		// max size error
		var maxBytesError *http.MaxBytesError

		// using a switch to handle different errors
		switch {
		// check for syntax error
		case errors.As(err, &syntaxError):
			return fmt.Errorf("the body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			// check for unexpected EOF error
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("the body contains badly-formed JSON")
			// check for incorrect type error
		case errors.As(err, &unmarshalTypeError):
			// if the field is not empty, it means we have a specific field that is incorrect
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("the body contains the incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("the body contains the incorrect  JSON type (at character %d)", unmarshalTypeError.Offset)
			// check for empty body error
		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")
			// check for unknown field error
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(),
				"json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
			//Size
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("the body must not be larger than %d bytes", maxBytesError.Limit)
			// some error the programmer made
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	// call decode again to check if there is only a single json value in the body
	err = dec.Decode(&struct{}{})

	// if the error is not EOF, then there is more than one value in the body
	if !errors.Is(err, io.EOF) {
		return errors.New("the body must only contain a single JSON value")
	}

	return nil
}

// Helper function to read an id parameter from the url
func (a *app) readIDParam(r *http.Request) (int64, error) {
	// get the url parameters
	params := httprouter.ParamsFromContext(r.Context())

	// convert the id parameter to an int64
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

/************************************************************************************************************/
// Helper functions for reading URL parameters
/************************************************************************************************************/

// readIDParameter extracts and validates an "id" parameter from the URL
func (app *app) readIDParameter(r *http.Request) (int64, error) {

	params := httprouter.ParamsFromContext(r.Context()) // get the URL parameters from the request context

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64) // parse the "id" parameter as a base-10 int64
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter") // return an error if parsing fails or id is less than 1
	}

	return id, nil // return the valid id
}

// getSingleQueryParameter retrieves a single query parameter from the URL, returning a default value if not found
func (app *app) getSingleQueryParameter(params url.Values, key string, defaultValue string) string {
	result := params.Get(key) // get the value of the specified query parameter
	if result == "" {
		return defaultValue // return the default value if the parameter is not found
	}
	return result // return the parameter value
}

// getMultipleQueryParameter retrieves multiple values for a query parameter from the URL, returning a default slice if not found
func (app *app) getMultipleQueryParameter(params url.Values, key string, defaultValue []string) []string {
	result := params.Get(key) // get the values of the specified query parameter
	if result == "" {
		return defaultValue // return the default slice if the parameter is not found
	}
	return strings.Split(result, ",") // split the parameter value by commas and return the resulting slice
}

// getSingleIntQueryParameter retrieves and validates a single integer query parameter from the URL, returning a default value if not found or invalid
func (app *app) getSingleIntQueryParameter(params url.Values, key string, defaultValue int64, v *validator.Validator) int64 {
	result := params.Get(key) // get the value of the specified query parameter
	if result == "" {
		return defaultValue // return the default value if the parameter is not found
	}

	i, err := strconv.ParseInt(result, 10, 64) // attempt to convert the parameter value to an integer
	if err != nil {
		v.AddError(key, "must be an integer value") // add a validation error if conversion fails
		return defaultValue                         // return the default value in case of error
	}

	return i // return the valid integer value
}

// GetSingleFloatQueryParameter retrieves and validates a single float query parameter from the URL, returning a default value if not found or invalid
func (app *app) getSingleFloatQueryParameter(params url.Values, key string, defaultValue float64, v *validator.Validator) float64 {
	result := params.Get(key) // get the value of the specified query parameter
	if result == "" {
		return defaultValue // return the default value if the parameter is not found
	}

	f, err := strconv.ParseFloat(result, 64) // attempt to convert the parameter value to a float
	if err != nil {
		v.AddError(key, "must be a float value") // add a validation error if conversion fails
		return defaultValue                      // return the default value in case of error
	}

	return f // return the valid float value
}

// getSingleDateQueryParameter retrieves and validates a single date query parameter from the URL, returning a default value if not found or invalid
func (app *app) getSingleDateQueryParameter(params url.Values, key string, defaultValue string, v *validator.Validator) string {
	result := params.Get(key) // get the value of the specified query parameter
	if result == "" {
		return defaultValue // return the default value if the parameter is not found
	}

	// Validate the date format (assuming "YYYY-MM-DD" format)
	if _, err := strconv.ParseInt(strings.ReplaceAll(result, "-", ""), 10, 64); err != nil || len(result) != 10 {
		v.AddError(key, "must be a valid date in YYYY-MM-DD format") // add a validation error if format is incorrect
		return defaultValue                                          // return the default value in case of error
	}

	return result // return the valid date string
}

// getOptionalBoolQueryParameter retrieves a boolean query parameter returning a pointer if present.
func (app *app) getOptionalBoolQueryParameter(params url.Values, key string, v *validator.Validator) *bool {
	value := params.Get(key)
	if value == "" {
		return nil
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		v.AddError(key, "must be true or false")
		return nil
	}

	return &b
}

// getOptionalInt64QueryParameter retrieves an int64 query parameter returning a pointer if present.
func (app *app) getOptionalInt64QueryParameter(params url.Values, key string, v *validator.Validator) *int64 {
	value := params.Get(key)
	if value == "" {
		return nil
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return nil
	}

	return &i
}

// readFilters constructs a Filters struct using standard query parameters and validates it.
func (app *app) readFilters(query url.Values, defaultSort string, defaultPageSize int64, safelist []string, v *validator.Validator) data.Filter {
	filters := data.Filter{
		Page:         app.getSingleIntQueryParameter(query, "page", 1, v),
		PageSize:     app.getSingleIntQueryParameter(query, "page_size", defaultPageSize, v),
		SortBy:       app.getSingleQueryParameter(query, "sort", defaultSort),
		SortSafeList: safelist,
	}

	data.ValidateFilters(v, filters)
	return filters
}

/************************************************************************************************************/
// Go routine helper functions
/************************************************************************************************************/
// background runs a function in the background as a goroutine, recovering from any panics and logging them
func (app *app) background(fn func()) {
	app.wg.Add(1) // increment the wait group counter

	go func() {
		defer app.wg.Done() // decrement the wait group counter when the goroutine completes

		// recover from any panics and log the error
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error("panic recovered in background goroutine", slog.Any("error", err)) // log the panic error
			}
		}()

		fn() // execute the provided function
	}()
}
