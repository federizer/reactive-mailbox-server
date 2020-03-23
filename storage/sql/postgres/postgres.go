package storage

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	postgres "github.com/lib/pq"
	"log"
	"reflect"
)

// PostgresStore represents the session store.
type PostgresStore struct {
	db *sql.DB
}

func (p *PostgresStore) ListDrafts(token string) (b []byte, exists bool, err error) {
	row := p.db.QueryRow("SELECT data FROM sessions WHERE token = $1 AND current_timestamp < expiry", token)
	err = row.Scan(&b)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func List(db *sql.DB, table string, result interface{}, clause ...string) error {
	resultv := reflect.ValueOf(result)
	/*if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}*/
	slicev := resultv.Elem()
	elemt := slicev.Type().Elem()
	query := "SELECT id, subject, body_uri, mimetype FROM " + table

	if len(clause) > 0 {
		for _, v := range clause {
			query = query + " " + v
		}
	}
	log.Println(query)
	rows, err := db.Query(query)
	if err != nil {
		sqlErr, ok := err.(postgres.Error)
		if !ok {
			return err
		}
		return sqlErr
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		jsonStr := ""
		err := rows.Scan(&jsonStr)
		if err != nil {
			return err
		}
		elemp := reflect.New(elemt)
		json.Unmarshal([]byte(jsonStr), elemp.Interface())
		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}
