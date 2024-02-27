package rdbms

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type RDBMS interface {
	MigrateUp(source string) error

	MigrateDown(source string) error

	Create(query string, args []any) (uint64, error)

	Read(query string, args []any, dest []any) error

	Update(query string, args []any) error

	Delete(query string, args []any) error
}

type rdbms struct {
	db *sql.DB
}

var (
	ErrPrepareStatement = "error when tying to prepare statement"

	ErrCreate    = "error when tying to create entry"
	ErrDuplicate = "entry exists"

	ErrRead         = "error when tying to read entry"
	ErrReadNotFound = "there is no entry with provided arguments"

	ErrUpdate = "error when tying to update entry"

	ErrDelete = "error when tying to delete entry"
)

func (db *rdbms) Create(query string, args []any) (uint64, error) {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s\n%v", ErrPrepareStatement, err)
	}
	defer stmt.Close()

	var lastInsertId int
	if err = stmt.QueryRow(args...).Scan(&lastInsertId); err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, fmt.Errorf("%s\n%v", ErrDuplicate, err)
		}

		return 0, fmt.Errorf("%s\n%v", ErrCreate, err)
	}

	return uint64(lastInsertId), nil
}

func (db *rdbms) Read(query string, args []any, dest []any) error {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s\n%v", ErrPrepareStatement, err)
	}
	defer stmt.Close()

	result := stmt.QueryRow(args...)
	err = result.Scan(dest...)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(ErrReadNotFound)
		}

		return fmt.Errorf("%s\n%v", ErrRead, err)
	}

	return nil
}

func (db *rdbms) Update(query string, args []any) error {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s\n%v", ErrPrepareStatement, err)
	}
	defer stmt.Close()

	if _, err = stmt.Exec(args); err != nil {
		return fmt.Errorf("%s\n%v", ErrUpdate, err)
	}

	return nil
}

func (db *rdbms) Delete(query string, args []any) error {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s\n%v", ErrPrepareStatement, err)
	}
	defer stmt.Close()

	if _, err = stmt.Exec(args...); err != nil {
		return fmt.Errorf("%s\n%v", ErrDelete, err)
	}

	return nil
}
