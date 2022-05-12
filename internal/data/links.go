package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Links struct {
	Id          int64  `json:"id"`
	ActiveLink  string `json:"active_link"`
	HistoryLink string `json:"history_link"`
}

type LinksModel struct {
	DB *sql.DB
}

func (l LinksModel) GetByLink(historyLinks string) (*Links, error) {
	if len(historyLinks) < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, history_link,active_link
		FROM links
        where history_link = $1`

	var links Links

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, historyLinks).Scan(
		&links.Id,
		&links.HistoryLink,
		&links.ActiveLink,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &links, nil
}

func (l LinksModel) GetById(id int) (*Links, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, history_link,active_link
		FROM links
        where id = $1`

	var links Links

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, id).Scan(
		&links.Id,
		&links.HistoryLink,
		&links.ActiveLink,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &links, nil
}

func (l LinksModel) Insert(link *Links) error {

	query := `INSERT INTO links (active_link,history_link) 
        VALUES ($1,$2)
        RETURNING id`

	args := []interface{}{link.Id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return l.DB.QueryRowContext(ctx, query, args...).Scan(&link.Id)
}

func (l LinksModel) GetAll(name string, filters Filters) ([]*Links, error) {
	query := fmt.Sprintf(`
		SELECT  id, active_link, history_link
		FROM links
        WHERE (to_tsvector('simple', history_link) @@ plainto_tsquery('simple', $1) OR $1 = '')
        ORDER BY id ASC`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{name}

	rows, err := l.DB.QueryContext(ctx, query, args...)

	defer rows.Close()

	var links []*Links
	for rows.Next() {

		var link Links

		err := rows.Scan(
			&link.Id,
			&link.ActiveLink,
			&link.HistoryLink,
		)
		if err != nil {
			return nil, err
		}

		links = append(links, &link)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (l LinksModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM links
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := l.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (l LinksModel) Update(link *Links) error {
	query := `UPDATE links 
        SET  active_link = $1, history_link = $2
        WHERE id = $3
        RETURNING id`

	args := []interface{}{
		link.ActiveLink,
		link.HistoryLink,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := l.DB.QueryRowContext(ctx, query, args...).Scan(&link.Id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
