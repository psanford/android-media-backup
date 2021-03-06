package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"gioui.org/app"
	_ "github.com/mattn/go-sqlite3"
	"github.com/retailnext/unixtime"
)

type DB struct {
	DB *sql.DB
}

func Open() (*DB, error) {
	dir, err := app.DataDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "mediabackup.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = initDB(db)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

type UploadState int

func (s UploadState) String() string {
	switch s {
	case UploadPending:
		return "UploadPending"
	case UploadInProgress:
		return "UploadInProgress"
	case UploadSuccess:
		return "UploadSuccess"
	case UploadFailed:
		return "UploadFailed"
	default:
		return fmt.Sprintf("UnkownState<%d>", s)
	}
}

const (
	UploadPending     UploadState = 1
	UploadInProgress  UploadState = 2
	UploadSuccess     UploadState = 3
	UploadFailed      UploadState = 4
	UploadFileDeleted UploadState = 5
)

func initDB(db *sql.DB) error {
	var createConfig = `CREATE TABLE IF NOT EXISTS config (
key text PRIMARY KEY,
val
)`

	_, err := db.Exec(createConfig)
	if err != nil {
		return err
	}

	var createFile = `CREATE TABLE IF NOT EXISTS file (
name text PRIMARY KEY,
created_epoch_ms int,
upload_started_epoch_ms int,
upload_end_epoch_ms int,
size int,
path text,
state int
)`

	_, err = db.Exec(createFile)
	if err != nil {
		return err
	}

	return nil
}

type File struct {
	Name          string
	Path          string
	Created       time.Time
	UploadStarted time.Time
	UploadEnd     time.Time
	Size          int64
	State         UploadState
}

func (db *DB) GetFiles() ([]File, error) {
	rows, err := db.DB.Query("select name, created_epoch_ms, upload_started_epoch_ms, upload_end_epoch_ms, size, path, state from file order by created_epoch_ms desc")
	if err != nil {
		return nil, err
	}

	var files []File

	for rows.Next() {
		var file File
		var (
			createdMS     *int64
			uploadStartMS *int64
			uploadEndMS   *int64
		)
		err = rows.Scan(&file.Name, &createdMS, &uploadStartMS, &uploadEndMS, &file.Size, &file.Path, &file.State)
		if err != nil {
			return nil, err
		}

		if createdMS != nil {
			file.Created = unixtime.ToTime(*createdMS, time.Millisecond)
		}
		if uploadStartMS != nil {
			file.UploadStarted = unixtime.ToTime(*uploadStartMS, time.Millisecond)
		}
		if uploadEndMS != nil {
			file.UploadEnd = unixtime.ToTime(*uploadEndMS, time.Millisecond)
		}

		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (db *DB) CreatePending(name, path string, modTime time.Time, size int64) (*File, error) {
	ts := unixtime.ToUnix(modTime, time.Millisecond)
	_, err := db.DB.Exec("insert into file (name, created_epoch_ms, size, path, state) values (?,?,?,?,?)", name, ts, size, path, UploadPending)
	if err != nil {
		return nil, err
	}

	file := File{
		Name:    name,
		Path:    path,
		Created: modTime,
		Size:    size,
		State:   UploadPending,
	}

	return &file, nil
}

func (db *DB) StartUpload(name string) error {
	ts := unixtime.ToUnix(time.Now(), time.Millisecond)
	_, err := db.DB.Exec("update file set state = ?, upload_started_epoch_ms = ? where name = ?", UploadInProgress, ts, name)
	return err
}

func (db *DB) EndUpload(name string, success bool) error {
	ts := unixtime.ToUnix(time.Now(), time.Millisecond)
	state := UploadFailed
	if success {
		state = UploadSuccess
	}
	_, err := db.DB.Exec("update file set state = ?, upload_end_epoch_ms = ? where name = ?", state, ts, name)
	return err
}

func (db *DB) ResetFiles() error {
	_, err := db.DB.Exec("delete from file")
	return err
}

var (
	confKeyEnabled  = "enabled"
	confKeyURL      = "url"
	confKeyUsername = "username"
	confKeyPassword = "password"
)

func (db *DB) Enabled() (bool, error) {
	var enabled bool
	err := db.confGet(confKeyEnabled, &enabled)
	return enabled, err
}

func (db *DB) SetEnabled(val bool) error {
	return db.confSet(confKeyEnabled, val)
}

func (db *DB) URL() (string, error) {
	var url string
	err := db.confGet(confKeyURL, &url)
	return url, err
}

func (db *DB) SetURL(url string) error {
	return db.confSet(confKeyURL, url)
}

func (db *DB) Username() (string, error) {
	var username string
	err := db.confGet(confKeyUsername, &username)
	return username, err
}

func (db *DB) SetUsername(username string) error {
	return db.confSet(confKeyUsername, username)
}

func (db *DB) Password() (string, error) {
	var password string
	err := db.confGet(confKeyPassword, &password)
	return password, err
}

func (db *DB) SetPassword(password string) error {
	return db.confSet(confKeyPassword, password)
}

func (db *DB) confGet(key string, val interface{}) error {
	row := db.DB.QueryRow("select val from config where key = ?", key)
	return row.Scan(val)
}

func (db *DB) confSet(key string, val interface{}) error {
	_, err := db.DB.Exec("insert or replace into config (key, val) values (?, ?)", key, val)
	return err
}
