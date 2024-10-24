package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/abner-tech/Credentials-Api.git/internal/data"
	"github.com/abner-tech/Credentials-Api.git/internal/validator"
)

func (a *applicationDependences) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	//create a struct to hold a credential
	//we use struct tags [` `] to make the names display in lowercase
	var incomingData struct {
		Email_address string `json:"email_address"`
		Name          string `json:"name"`
	}

	//perform decoding

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	credential := &data.Credential{
		Email_address: incomingData.Email_address,
		Name:          incomingData.Name,
	}

	v := validator.New()
	//do validation
	data.ValidateCredential(v, credential)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors) //implemented later
		return
	}

	//add credential to the credentials table in database
	err = a.credentialModel.Insert(credential)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//for now display the result
	// fmt.Fprintf(w, "%+v\n", incomingData)

	//set a location header, the path to the newly created credentials
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/comments/%d", credential.ID))

	//send a json response with a 201 (new reseource created) status code
	data := envelope{
		"credential": credential,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependences) fetchCommentByID(w http.ResponseWriter, r *http.Request) (*data.Credential, error) {
	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query the credentials table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)

	}

	// Call Get() to retrieve the credential with the specified id
	credential, err := a.credentialModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}

	}
	return credential, nil
}

func (a *applicationDependences) displayCommentHandler(w http.ResponseWriter, r *http.Request) {

	credential, err := a.fetchCommentByID(w, r)
	if err != nil {
		return
	}
	// display the credential
	data := envelope{
		"credential": credential,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependences) updateCommentHandler(w http.ResponseWriter, r *http.Request) {

	credential, err := a.fetchCommentByID(w, r)
	if err != nil {
		return
	}

	// Use our temporary incomingData struct to hold the data
	// Note: I have changed the types to pointer to differentiate
	// between the client leaving a field empty intentionally
	// and the field not needing to be updated
	var incomingData struct {
		Email_address *string `json:"email_address"`
		Name          *string `json:"name"`
	}

	// perform the decoding
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// We need to now check the fields to see which ones need updating
	// if incomingData.Content is nil, no update was provided
	if incomingData.Email_address != nil {
		credential.Email_address = *incomingData.Email_address
	}
	// if incomingData.Author is nil, no update was provided
	if incomingData.Name != nil {
		credential.Name = *incomingData.Name
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateCredential(v, credential)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// perform the update
	err = a.credentialModel.Update(credential)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	data := envelope{
		"credential": credential,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependences) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}
	err = a.credentialModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//diplay the credential
	data := envelope{
		"message": "signup credentials deleted successfully",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependences) listCommentHandler(w http.ResponseWriter, r *http.Request) {
	//create a struct to hold the query parameters
	//Later, fields will be added for pagination and sorting (filters)
	var queryParameterData struct {
		Content string
		Author  string
	}

	//get query parameters from url
	queryParameter := r.URL.Query()

	//load the query parameters into the created struct
	queryParameterData.Content = a.getSingleQueryParameter(queryParameter, "content", "")
	queryParameterData.Author = a.getSingleQueryParameter(queryParameter, "author", "")

	//call GetAll to retrieve all credentials of the DB
	credentials, err := a.credentialModel.GetAll(queryParameterData.Content, queryParameterData.Author)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
			return
		default:
			a.serverErrorResponse(w, r, err)
			return
		}
	}

	data := envelope{
		"credentials": credentials,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}
