package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"../util"
)

type SqliteStore struct {
	FilePath string

	db *sql.DB
}

func (self *SqliteStore) Init() error {
	if len(self.FilePath) == 0 {
		return errors.New("FilePath option of SQLite store is empty")
	}

	db, errSql := sql.Open("sqlite3", self.FilePath)
	if errSql != nil {
		return errSql
	}

	self.db = db

	self.CreateDataset(JobDataset,
		[]string{"_id", "title", "rule"},
		[]string{"integer not null primary key", "text not null", "text not null"})

	return nil
}

func (self *SqliteStore) CreateDataset(datasetName string, fieldNames []string, fieldTypes []string) (err error) {
	fieldCount := util.IntMin(len(fieldNames), len(fieldTypes))
	if fieldCount <= 0 {
		return errors.New("Dataset must have at least one field")
	}

	fields := make([]string, fieldCount)
	for i := 0; i < fieldCount; i++ {
		fields[i] = fmt.Sprintf("%s %s", fieldNames[i], fieldTypes[i])
	}
	fieldsString := strings.Join(fields, ",")

	sql := fmt.Sprintf(`CREATE TABLE "%s" (%s)`, datasetName, fieldsString)
	_, errSql := self.db.Exec(sql)
	if errSql != nil {
		return errSql
	}

	return nil
}

func (self *SqliteStore) InsertObject(dataset string, fields []string, values []interface{}) (oid string, err error) {
	fieldCount := util.IntMin(len(fields), len(values))
	if fieldCount <= 0 {
		return "", errors.New("No object will be created")
	}

	fs := make([]string, fieldCount)
	copy(fs, fields)
	fString := strings.Join(fs, ",")

	vs := make([]string, fieldCount)
	for i := 0; i < fieldCount; i++ {
		vs[i] = "?"
	}
	vString := strings.Join(vs, ",")

	sql := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, dataset, fString, vString)
	result, errSql := self.db.Exec(sql, values...)
	if errSql != nil {
		return "", errSql
	}

	insertId, errResult := result.LastInsertId()
	if errResult != nil {
		return "", errResult
	}

	return strconv.FormatInt(insertId, 10), nil
}

func (self *SqliteStore) DeleteObjects(dataset string, oids []string) (count int64, err error) {
	oidCount := len(oids)
	if oidCount == 0 {
		return 0, nil
	}

	var (
		sql    string
		errSql error
	)

	if oidCount == 1 {
		sql = fmt.Sprintf(`DELETE FROM "%s" WHERE _id = ?`, dataset)
	} else {
		holders := make([]string, oidCount)
		for i := 0; i < oidCount; i++ {
			holders[i] = "?"
		}
		inClause := strings.Join(holders, ",")

		sql = fmt.Sprintf("DELETE FROM %s WHERE _id IN (%s)", dataset, inClause)
	}

	result, errSql := self.db.Exec(sql, oids)
	if errSql != nil {
		return -1, errSql
	}

	return result.RowsAffected()
}

func (self *SqliteStore) QueryAllJobs() (jobs []map[string]interface{}, err error) {
	sql := fmt.Sprintf(`SELECT "_id","title","rule" FROM "%s"`, JobDataset)

	rows, errSql := self.db.Query(sql)
	if errSql != nil {
		return nil, errSql
	}
	defer rows.Close()

	objects := []map[string]interface{}{}
	for rows.Next() {
		var (
			_id   int
			title string
			rule  string
		)

		errSql = rows.Scan(&_id, &title, &rule)
		if errSql != nil {
			return nil, errSql
		}

		objects = append(objects, map[string]interface{}{
			"_id":   _id,
			"title": title,
			"rule":  rule,
		})
	}

	errSql = rows.Err()
	if errSql != nil {
		return nil, errSql
	}

	return objects, nil
}

func (self *SqliteStore) Destroy() error {
	if self.db != nil {
		self.db.Close()
	}

	return nil
}
