package data

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"time"

	"github.com/abner-tech/Credentials-Api.git/internal/validator"
)

// each name begins with uppercase to make them exportable/ public
type Credential struct {
	ID            int64     `json:"id"`            //unique value per credential
	Created_at    time.Time `json:"-"`             //credential timestamp
	Email_address string    `json:"email_address"` //email address for the credential
	Name          string    `json:"name"`          //person name
	Version       int32     `json:"version"`       //icremented on each update
}

// CredentialModel that expects a connection pool
type CredentialModel struct {
	DB *sql.DB
}

// Insert Row to credentials table
// expects a pointer to the actual credential content
func (c CredentialModel) Insert(credential *Credential) error {
	//the sql query to be executed against the database table
	query := `
	INSERT INTO credentials (name, email_address)
	VALUES ($1, $2)
	RETURNING id, created_at, version`

	//the actual values to be passed into $1 and $2
	args := []any{credential.Name, credential.Email_address}

	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// execute the query against the credentials database table. We ask for the the
	// id, created_at, and version to be sent back to us which we will use
	// to update the credential struct later on
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&credential.ID,
		&credential.Created_at,
		&credential.Version)

}

func ValidateCredential(v *validator.Validator, credential *Credential) {
	//check if the email field is empty
	v.Check(credential.Email_address != "", "email_address", "must be provided")
	//check if the name field is empty
	v.Check(credential.Name != "", "name", "must be provided")
	//check if the email field is empty
	v.Check(len(credential.Email_address) <= 50, "email_address", "must not be more than 50 bytes long")
	//check is name field is empty
	v.Check(len(credential.Name) <= 25, "name", "must not be more than 25 bytes long")
	//check for email address to contain a @ and a . on its content
	_, err := mail.ParseAddress(credential.Email_address)
	if err != nil {
		v.Check(false, "email_address", "invalid email provided")
		return
	}
}

// get a credential from DB based on ID
func (c CredentialModel) Get(id int64) (*Credential, error) {
	//check if the id is valid
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	//the sql query to be excecuted against the database table
	query := `
	SELECT id, created_at, email_address, name, version
	FROM credentials
	WHERE id = $1
	`

	//declare a variable of type credential to hold the returned values
	var credential Credential

	//set 3-second context/timer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&credential.ID,
		&credential.Created_at,
		&credential.Email_address,
		&credential.Name,
		&credential.Version,
	)
	//check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &credential, nil
}

func (c CredentialModel) GetAll(content string, author string) (*[]Credential, error) {
	query := `
	SELECT id, created_at, email_address, name, version
	FROM credentials
	WHERE (to_tsvector('simple',email_address) @@
		plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple',name) @@
		plainto_tsquery('simple',$2) OR $2 = '')
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	credentials, err := c.DB.QueryContext(ctx, query, content, author)
	//check for errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	var credential []Credential
	for credentials.Next() {
		var tempCredential Credential
		if err := credentials.Scan(&tempCredential.ID, &tempCredential.Created_at, &tempCredential.Email_address, &tempCredential.Name, &tempCredential.Version); err != nil {
			return nil, err
		}
		credential = append(credential, tempCredential)
	}
	return &credential, nil
}

// update  a specific record from the credentials table
func (c CredentialModel) Update(credential *Credential) error {
	//the sql query to be excecuted against the DB table
	//Every time make an update, version number is incremented

	query := `
	UPDATE credentials
	SET email_address=$1, name=$2, version=version+1
	WHERE id = $3
	RETURNING version
	`

	args := []any{credential.Email_address, credential.Name, credential.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return c.DB.QueryRowContext(ctx, query, args...).Scan(&credential.Version)

}

// delete a specific credential form the credentials table
func (c CredentialModel) Delete(id int64) error {
	//check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}

	//sql querry to be excecuted against the database table
	query := `
	DELETE FROM credentials
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// ExecContext does not return any rows unlike QueryRowContext.
	// It only returns  information about the the query execution
	// such as how many rows were affected
	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//maybe wrong id for record was given so we sort of try checking
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
