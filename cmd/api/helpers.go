package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/abner-tech/Comments-Api.git/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (a *applicationDependences) writeJSON(w http.ResponseWriter,
	status int, data envelope,
	headers http.Header) error {
	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	jsResponse = append(jsResponse, '\n')
	//aditional headers to be set
	for key, value := range headers {
		w.Header()[key] = value
		//w.Header().Set(key, value[])
	}
	//set content type header
	w.Header().Set("Content-Type", "application/json")
	//explicitly set the response status code
	w.WriteHeader(status)
	_, err = w.Write(jsResponse)
	if err != nil {
		return err
	}

	return nil
}

func (a *applicationDependences) readJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	//max size of the request body in this case is 250KB
	maxBytes := 256_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	//decoder to check for unkown fields
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	//start decoding process
	err := dec.Decode(destination)
	//check for different errors
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("the body contains badly-formd json (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("the body contains badly-formed json")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("the body contains Incorrect json type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("the body contains the incorrect json type at character: %d", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")
			//checking for unknown field errors
		case strings.HasPrefix(err.Error(), "Json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unkown field ")
			return fmt.Errorf("body containes unkown key %s", fieldName)
		//check if body is grater than limit of 250KB
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("the body must not be larger than %d bytes", maxBytesError.Limit)
		//the program messed up
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
			//default
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		//more DATA is present
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (a *applicationDependences) readIDParam(r *http.Request) (int64, error) {
	//get the url parameters
	params := httprouter.ParamsFromContext(r.Context())
	//convert id from string to int
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}

func (a *applicationDependences) getSingleQueryParameter(queryParameter url.Values, key string, defaultValue string) string {
	//url.values is a key:value hash map of the query parameters
	result := queryParameter.Get(key)
	if result == "" {
		return defaultValue
	}
	return result
}

func (a *applicationDependences) getMultipleQueryParameters(queryParameter url.Values, key string, defaultValue []string) []string {
	result := queryParameter.Get(key)
	if result == "" {
		return defaultValue
	}
	return strings.Split(result, ",")
}

// NOTE: this method can cause validation errors when attempting to convert from string to valid int value
func (a *applicationDependences) getSingleIntigetParameter(queryParameter url.Values, key string, defaultValue int, v *validator.Validator) int {
	result := queryParameter.Get(key)
	if result == "" {
		return defaultValue
	}
	//attempting to convert from string to int
	intValue, err := strconv.Atoi(result)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return intValue
}
