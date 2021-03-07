package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

const (
	CONTENT_TYPE = "Content-Type"
	JSON         = "application/json"
	TEXT         = "text/plain"
)

// Dispatch a JSON response using a given http response writer.
// Use the method parameter `statusCode` to set the header status code.
// Use the method parameter `v` to supply the object to be serialized to JSON.
func Json(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	w.Header().Set(CONTENT_TYPE, JSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

// Dispatch a text response using a given http response writer.
// Use the method parameter `statusCode` to set the header status code.
// Use the method parameter `v` to supply the object to be serialized to text.
func Text(w http.ResponseWriter, r *http.Request, statusCode int, v interface{}) {
	w.Header().Set(CONTENT_TYPE, TEXT)
	w.WriteHeader(statusCode)
	fmt.Fprint(w, v)
}

// Decode an object with a given interface from a JSON request.
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	contentType := r.Header.Get(CONTENT_TYPE)

	// check that the the Content-Type header has the value application/json.
	if mt, _, err := mime.ParseMediaType(contentType); err != nil || mt != JSON {
		msg := fmt.Sprintf("%s header is not `%s`", CONTENT_TYPE, JSON)
		return &RequestError{status: http.StatusUnsupportedMediaType, msg: msg}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &RequestError{status: http.StatusBadRequest, msg: msg}

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return &RequestError{status: http.StatusBadRequest, msg: msg}

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &RequestError{status: http.StatusBadRequest, msg: msg}

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &RequestError{status: http.StatusBadRequest, msg: msg}

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &RequestError{status: http.StatusBadRequest, msg: msg}

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &RequestError{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &RequestError{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
