package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"../common/util"
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
		[]string{"_id", "title", "rule", "createdAt"},
		[]string{"integer not null primary key", "text not null", "text not null", "timestamp DEFAULT CURRENT_TIMESTAMP"})
	self.createIndex(JobDataset, "createdAt")

	self.CreateDataset(TargetDataset,
		[]string{"_id", "hash", "mtag", "createdAt"},
		[]string{"integer not null primary key", "text not null", "text not null", "timestamp DEFAULT CURRENT_TIMESTAMP"})
	self.createIndex(TargetDataset, "hash")
	self.createIndex(TargetDataset, "createdAt")

	return nil
}

func (self *SqliteStore) HasDataset(datasetName string) bool {
	sql := fmt.Sprintf(`SELECT name FROM sqlite_master WHERE type='table' AND name='%s'`, datasetName)
	rows, errSql := self.db.Query(sql)
	if errSql != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

func (self *SqliteStore) CreateDataset(datasetName string, fieldNames []string, fieldTypes []string) error {
	fieldCount := util.IntMin(len(fieldNames), len(fieldTypes))
	if fieldCount <= 0 {
		return errors.New("Dataset must have at least one field")
	}

	fields := make([]string, fieldCount)
	for i := 0; i < fieldCount; i++ {
		fields[i] = fmt.Sprintf("%s %s", fieldNames[i], fieldTypes[i])
	}
	fieldsString := strings.Join(fields, ",")

	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s)`, datasetName, fieldsString)
	_, errSql := self.db.Exec(sql)
	if errSql != nil {
		return errSql
	}

	return nil
}

func (self *SqliteStore) EmptyDataset(datasetName string) error {
	sqlDelete := fmt.Sprintf(`DELETE FROM "%s"`, datasetName)
	_, errDelete := self.db.Exec(sqlDelete)
	if errDelete != nil {
		return errDelete
	}

	return nil
}

func (self *SqliteStore) InsertObject(datasetName string, fields []string, values []interface{}) (oid string, err error) {
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

	sql := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, datasetName, fString, vString)
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

func (self *SqliteStore) DeleteObjects(datasetName string, oids []string) (count int64, err error) {
	ids := []interface{}{}
	for _, oid := range oids {
		id, err := strconv.Atoi(oid)
		if err == nil {
			ids = append(ids, id)
		}
	}

	idCount := len(ids)
	if idCount == 0 {
		return 0, nil
	}

	var (
		sql    string
		errSql error
	)

	if idCount == 1 {
		sql = fmt.Sprintf(`DELETE FROM "%s" WHERE _id = ?`, datasetName)
	} else {
		holders := make([]string, idCount)
		for i := 0; i < idCount; i++ {
			holders[i] = "?"
		}
		inClause := strings.Join(holders, ",")

		sql = fmt.Sprintf("DELETE FROM %s WHERE _id IN (%s)", datasetName, inClause)
	}

	result, errSql := self.db.Exec(sql, ids...)
	if errSql != nil {
		return -1, errSql
	}

	return result.RowsAffected()
}

func (self *SqliteStore) UpdateObject(datasetName string, oid string, fields []string, values []interface{}) (count int64, err error) {
	idInt, errId := strconv.Atoi(oid)
	if errId != nil {
		return 0, errId
	}
	fieldCount := util.IntMin(len(fields), len(values))
	if fieldCount <= 0 {
		return 0, errors.New("No object will be updated")
	}

	fvs := make([]string, fieldCount)
	for i := 0; i < fieldCount; i++ {
		fvs[i] = fmt.Sprintf("%s=?", fields[i])
	}
	fvString := strings.Join(fvs, ",")

	params := make([]interface{}, fieldCount+1)
	copy(params, values)
	params[fieldCount] = idInt

	sql := fmt.Sprintf(`UPDATE "%s" SET %s WHERE _id=?`, datasetName, fvString)
	result, errSql := self.db.Exec(sql, params...)
	if errSql != nil {
		return -1, errSql
	}

	return result.RowsAffected()
}

func (self *SqliteStore) QueryAllJobs() (jobs []map[string]interface{}, err error) {
	sql := fmt.Sprintf(`SELECT "_id","title","rule","status" FROM "%s" ORDER BY createdAt DESC`, JobDataset)

	rows, errSql := self.db.Query(sql)
	if errSql != nil {
		return nil, errSql
	}
	defer rows.Close()

	objects := []map[string]interface{}{}
	for rows.Next() {
		var (
			_id    int64
			title  string
			rule   string
			status string
		)

		errSql = rows.Scan(&_id, &title, &rule, &status)
		if errSql != nil {
			return nil, errSql
		}

		objects = append(objects, map[string]interface{}{
			"_id":    strconv.FormatInt(_id, 10),
			"title":  title,
			"rule":   rule,
			"status": status,
		})
	}

	errSql = rows.Err()
	if errSql != nil {
		return nil, errSql
	}

	return objects, nil
}

func (self *SqliteStore) GetJob(id string) (job map[string]interface{}, err error) {
	sql := fmt.Sprintf(`SELECT "_id","title","rule","status" FROM "%s" WHERE _id=? LIMIT 1`, JobDataset)

	idInt, errId := strconv.Atoi(id)
	if errId != nil {
		return nil, errId
	}

	rows, errSql := self.db.Query(sql, idInt)
	if errSql != nil {
		return nil, errSql
	}
	defer rows.Close()

	var object map[string]interface{} = nil
	if rows.Next() {
		var (
			_id    int64
			title  string
			rule   string
			status string
		)

		errSql = rows.Scan(&_id, &title, &rule, &status)
		if errSql != nil {
			return nil, errSql
		}

		object = map[string]interface{}{
			"_id":    strconv.FormatInt(_id, 10),
			"title":  title,
			"rule":   rule,
			"status": status,
		}
	}

	errSql = rows.Err()
	if errSql != nil {
		return nil, errSql
	}

	return object, nil
}

func (self *SqliteStore) GetLatestTarget(hash string) (target map[string]interface{}, err error) {
	sql := fmt.Sprintf(`SELECT "_id","hash","mtag","createdAt" FROM "%s" WHERE hash=? ORDER BY createdAt DESC LIMIT 1`, TargetDataset)

	rows, errSql := self.db.Query(sql, hash)
	if errSql != nil {
		return nil, errSql
	}
	defer rows.Close()

	var object map[string]interface{} = nil
	if rows.Next() {
		var (
			_id       int64
			hash      string
			mtag      string
			createdAt time.Time
		)

		errSql = rows.Scan(&_id, &hash, &mtag, &createdAt)
		if errSql != nil {
			return nil, errSql
		}

		object = map[string]interface{}{
			"_id":       strconv.FormatInt(_id, 10),
			"hash":      hash,
			"mtag":      mtag,
			"createdAt": createdAt,
		}
	}

	errSql = rows.Err()
	if errSql != nil {
		return nil, errSql
	}

	return object, nil
}

func (self *SqliteStore) QueryAllDataObjects(datasetName string) (objs []map[string]interface{}, err error) {
	sql := fmt.Sprintf(`SELECT "id","data" FROM "%s"`, datasetName)

	rows, errSql := self.db.Query(sql)
	if errSql != nil {
		return nil, errSql
	}
	defer rows.Close()

	objects := []map[string]interface{}{}
	for rows.Next() {
		var (
			id   string
			data string
		)

		errSql = rows.Scan(&id, &data)
		if errSql != nil {
			return nil, errSql
		}

		objects = append(objects, map[string]interface{}{
			"id":   id,
			"data": data,
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

func (self *SqliteStore) createIndex(datasetName string, fieldName string) error {
	sql := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "idx_%s_%s" ON "%s" ("%s" DESC)`, datasetName, fieldName, datasetName, fieldName)
	_, errSql := self.db.Exec(sql)
	if errSql != nil {
		return errSql
	}

	return nil
}
