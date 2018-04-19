package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DefaultWarnMax = 10
)

type sqlite3 struct {
	db *sql.DB
	mu sync.RWMutex
}
type sqlite3Factory struct{}

func init() {
	Register("sqlite3", &sqlite3Factory{})
}

func (f *sqlite3Factory) New(source string) (DB, error) {
	driver := &sqlite3{}

	var err error
	driver.db, err = sql.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}

	driver.mu.Lock()
	defer driver.mu.Unlock()

	_, err = driver.db.Exec(`
	CREATE TABLE IF NOT EXISTS warnings (
		chat_id BIGINT NOT NULL,
		user_id INT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		PRIMARY KEY (chat_id, user_id)
	)
	`)
	if err != nil {
		return nil, err
	}

	_, err = driver.db.Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS settings (
		chat_id BIGINT NOT NULL PRIMARY KEY,
		max_warn INT DEFAULT %d
	)
	`, DefaultWarnMax))
	if err != nil {
		return nil, err
	}

	_, err = driver.db.Exec(`
	CREATE TABLE IF NOT EXISTS word_whitelist (
		word TEXT NOT NULL PRIMARY KEY
	)
	`)
	if err != nil {
		return nil, err
	}

	_, err = driver.db.Exec(`
	CREATE TABLE IF NOT EXISTS word_blacklist (
		word TEXT NOT NULL PRIMARY KEY,
		count INT NOT NULL DEFAULT 0
	)
	`)
	if err != nil {
		return nil, err
	}

	_, err = driver.db.Exec(`
	CREATE TABLE IF NOT EXISTS rules (
		chat_id BIGINT NOT NULL PRIMARY KEY,
		rules TEXT NOT NULL
	)
	`)
	if err != nil {
		return nil, err
	}

	return driver, nil
}

func (dr *sqlite3) WarnAdd(chat_id int64, user_id int) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	updateStmt, err := tx.Prepare(`
	UPDATE warnings SET count = count + 1 WHERE chat_id = ? AND user_id = ?
	`)
	if err != nil {
		return err
	}

	updateResult, err := updateStmt.Exec(chat_id, user_id)
	if err != nil {
		return err
	}

	rowCount, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowCount == 0 {
		insertStmt, err := tx.Prepare(`
		INSERT INTO warnings (chat_id, user_id, count) VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}

		_, err = insertStmt.Exec(chat_id, user_id, 1)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (dr *sqlite3) WarnSet(chat_id int64, user_id, count int) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	updateStmt, err := tx.Prepare(`
	UPDATE warnings SET count = ? WHERE chat_id = ? AND user_id = ?
	`)
	if err != nil {
		return err
	}

	updateResult, err := updateStmt.Exec(count, chat_id, user_id)
	if err != nil {
		return err
	}

	rowCount, err := updateResult.RowsAffected()
	if err != nil {
		return err
	}

	if rowCount == 0 {
		insertStmt, err := tx.Prepare(`
		INSERT INTO warnings (chat_id, user_id, count) VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}

		_, err = insertStmt.Exec(chat_id, user_id, count)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (dr *sqlite3) WarnCount(chat_id int64, user_id int) (int, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	row := dr.db.QueryRow(`
	SELECT count FROM warnings WHERE chat_id = ? AND user_id = ?
	`, chat_id, user_id)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dr *sqlite3) WarnMax(chat_id int64) (int, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	row := dr.db.QueryRow(`
	SELECT max_warn FROM settings WHERE chat_id = ?
	`, chat_id)

	var maxWarn int
	err := row.Scan(&maxWarn)
	switch {
	case err == sql.ErrNoRows:
		return DefaultWarnMax, nil
	case err != nil:
		return 0, err
	}

	return maxWarn, nil
}

func (dr *sqlite3) WarnMaxSet(chat_id int64, count int) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO settings (chat_id, max_warn) VALUES (?, ?) 
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(chat_id, count)
	if err != nil {
		return err
	}

	return nil
}

func (dr *sqlite3) WordLogWlGet() ([]string, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	rows, err := dr.db.Query(`
	SELECT word FROM word_whitelist
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var whitelist []string

	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, err
		}

		whitelist = append(whitelist, word)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return whitelist, nil
}

func (dr *sqlite3) WordLogBlGet() ([]string, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	rows, err := dr.db.Query(`
	SELECT word FROM word_blacklist
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blacklist []string

	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return nil, err
		}

		blacklist = append(blacklist, word)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blacklist, nil
}

func (dr *sqlite3) WordLogWlAdd(word string) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	INSERT INTO word_whitelist (word) VALUES (?)
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(word)
	if err != nil {
		return err
	}

	return nil
}

func (dr *sqlite3) WordLogWlDel(word string) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	DELETE FROM word_whitelist WHERE word = ?
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(word)
	if err != nil {
		return err
	}

	return nil
}

func (dr *sqlite3) WordLogBlAdd(word string) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	INSERT INTO word_blacklist (word) VALUES (?)
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(word)
	if err != nil {
		return err
	}

	return nil
}

func (dr *sqlite3) WordLogBlDel(word string) (err error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	DELETE FROM word_blacklist WHERE word = ?
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(word)
	if err != nil {
		return err
	}

	return nil
}

func (dr *sqlite3) RulesGet(chat_id int64) (string, error) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	row := dr.db.QueryRow(`
	SELECT rules FROM rules WHERE chat_id = ?
	`, chat_id)

	var rules string
	err := row.Scan(&rules)
	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", err
	}

	return rules, nil
}

func (dr *sqlite3) RulesSet(chat_id int64, rules string) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	tx, err := dr.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare(`
	INSERT OR REPLACE INTO rules (chat_id, rules) VALUES (?, ?)
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(chat_id, rules)
	if err != nil {
		return err
	}

	return nil
}
