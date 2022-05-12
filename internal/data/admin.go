package data

import (
	"case-championship/internal/validator"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var AnonymousUser = &Admin{}

type Admin struct {
	ID       int64    `json:"id"`
	Login    string   `json:"login"`
	Password password `json:"-"`
}

func (u *Admin) IsAnonymous() bool {
	return u == AnonymousUser
}

func (a AdminModel) GetForToken(tokenScope, tokenPlaintext string) (*Admin, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
SELECT admins.id,  admins.login, admins.password_hash FROM admins
INNER JOIN tokens
ON admins.id = tokens.admins_id
WHERE tokens.hash = $1
AND tokens.scope = $2
AND tokens.expiry > $3`
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	var admin Admin
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := a.DB.QueryRowContext(ctx, query, args...).Scan(
		&admin.ID, &admin.Login, &admin.Password.hash,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &admin, nil
}

type AdminModel struct {
	DB *sql.DB
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, admin *Admin) {
	v.Check(admin.Login != "", "login", "must be provided")
	v.Check(len(admin.Login) <= 500, "login", "must not be more than 500 bytes long")

	if admin.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *admin.Password.plaintext)
	}
	if admin.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (a AdminModel) Insert(admin *Admin) error {
	query := `INSERT INTO admins (login, password_hash) VALUES ($1, $2) RETURNING id`
	args := []interface{}{admin.Login, admin.Password.hash}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// If the table already contains a record with this email address, then when we try // to perform the insert there will be a violation of the UNIQUE "users_email_key" // constraint that we set up in the previous chapter. We check for this error
	// specifically, and return custom ErrDuplicateEmail error instead.
	err := a.DB.QueryRowContext(ctx, query, args...).Scan(&admin.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "admins_login_key"`:
			return err
		default:
			return err
		}
	}
	return nil
}

func (a AdminModel) GetByLogin(login string) (*Admin, error) {
	query := `SELECT id,login, password_hash FROM admins WHERE login = $1`
	var admin Admin
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := a.DB.QueryRowContext(ctx, query, login).Scan(&admin.ID, &admin.Login, &admin.Password.hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &admin, nil
}

func (a AdminModel) Update(admin *Admin) error {
	query := ` UPDATE admins SET login = $1, password_hash = $2 WHERE id = $3 RETURNING id`
	args := []interface{}{
		admin.Login,
		admin.Password.hash,
		admin.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := a.DB.QueryRowContext(ctx, query, args...).Scan(&admin.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "admins_login_key"`:
			return err
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
