package upload

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/psanford/android-media-backup/db"
	"github.com/psanford/android-media-backup/jgo/wifi"
	"github.com/psanford/android-media-backup/ui/plog"
)

func Upload() error {
	p := "/sdcard/DCIM/Camera"
	files, err := ioutil.ReadDir(p)
	if err != nil {
		plog.Printf("read sdcard err: %s", err)
		return err
	}

	store, err := db.Open()
	if err != nil {
		plog.Printf("open db err: %s", err)
		return err
	}

	enabled, _ := store.Enabled()
	if !enabled {
		plog.Printf("service disabled, not uploading")
		return errors.New("service disabled")
	}

	dbFiles, err := store.GetFiles()
	if err != nil {
		plog.Printf("get files err: %s", err)
		return err
	}

	dbFilesMap := make(map[string]*db.File)
	for _, dbFile := range dbFiles {
		dbFile := dbFile
		dbFilesMap[dbFile.Name] = &dbFile
	}

	for _, f := range files {
		filename := f.Name()
		pp := filepath.Join(p, filename)
		modTime := f.ModTime()
		size := f.Size()

		plog.Printf("bgjob file=%s time=%s size=%d", filename, modTime, size)

		dbFile := dbFilesMap[filename]

		if dbFile == nil {
			plog.Printf("bgjob %s not in db, setting to pending", filename)
			dbFile, err = store.CreatePending(filename, pp, modTime, size)
			if err != nil {
				plog.Printf("bgjob %s create pending failed: %s", filename, err)
				continue
			}
			dbFilesMap[filename] = dbFile
		} else {
			plog.Printf("bgjob %s in db, state is %s", filename, dbFile.State)
		}
	}

	for _, f := range files {
		enabled, _ := store.Enabled()
		if !enabled {
			plog.Printf("service has been disabled, deferring remaining uploads")
			return errors.New("service disabled")
		}

		connState, err := wifi.ConnectionState()
		if err != nil || connState == wifi.ConnStateUnknown || connState == wifi.NoNetwork {
			plog.Printf("no network connection, deferring remaining uploads")
			return errors.New("no network")
		}

		allowMobile, _ := store.AllowMobileUpload()
		if !allowMobile && connState < wifi.Wifi {
			plog.Printf("not on wifi, deferring remaining uploads")
			return errors.New("no wifi")
		}

		filename := f.Name()
		fpath := filepath.Join(p, filename)
		modTime := f.ModTime()
		size := f.Size()

		dbFile := dbFilesMap[filename]

		if dbFile.State == db.UploadInProgress {
			plog.Printf("upload already in-progress for %s, this probably needs to be retired", dbFile.Name)
		} else if dbFile.State == db.UploadPending {
			err := store.StartUpload(dbFile.Name)
			if err != nil {
				plog.Printf("set upload to in-progress failed for=%s err=%s", dbFile.Name, err)
				continue
			}

			f, err := os.Open(fpath)
			if err != nil {
				plog.Printf("open file err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			summer := sha256.New()
			_, err = io.Copy(summer, f)
			if err != nil {
				plog.Printf("read file err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			id := hex.EncodeToString(summer.Sum(nil))

			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				plog.Printf("seek file err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			fileHeader := make([]byte, 512)
			io.ReadFull(f, fileHeader)

			contentType := http.DetectContentType(fileHeader)

			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				plog.Printf("seek file err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			dest, err := requestUploadURL(store, id, dbFile.Name, contentType, modTime, size)
			if err != nil {
				plog.Printf("request upload url err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			if dest.Status == StatusSkipUpload {
				plog.Printf("upload file skipped for=%s", dbFile.Name)
				store.EndUpload(dbFile.Name, db.UploadSkipped)
				continue
			}

			err = uploadFile(f, size, dest)
			if err != nil {
				plog.Printf("upload file err for=%s err=%s", dbFile.Name, err)
				store.EndUpload(dbFile.Name, db.UploadFailed)
				continue
			}

			plog.Printf("upload file success for=%s", dbFile.Name)
			store.EndUpload(dbFile.Name, db.UploadSuccess)
		}
	}

	return nil
}

func requestUploadURL(store *db.DB, id, name, contentType string, mtime time.Time, size int64) (*UploadDestination, error) {
	url, err := store.URL()
	if err != nil {
		return nil, err
	}
	username, err := store.Username()
	if err != nil {
		return nil, err
	}
	passwd, err := store.Password()
	if err != nil {
		return nil, err
	}

	meta := FileMetadata{
		ID:          id,
		Name:        name,
		Mtime:       mtime,
		Bytes:       size,
		ContentType: contentType,
	}

	jsontxt, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(jsontxt)

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.SetBasicAuth(username, passwd)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != http.StatusConflict {
		return nil, fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	var dest UploadDestination
	err = json.NewDecoder(resp.Body).Decode(&dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

func uploadFile(r io.Reader, size int64, dest *UploadDestination) error {
	if dest.Method == "" {
		dest.Method = "PUT"
	}
	req, err := http.NewRequest(dest.Method, dest.URL, r)
	if err != nil {
		return err
	}

	req.Header = dest.Headers
	req.ContentLength = size

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

type FileMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Mtime       time.Time `json:"mtime"`
	Bytes       int64     `json:"size"`
	ContentType string    `json:"content_type"`
	TestUpload  bool      `json:"test_upload"`
}

type Status string

var (
	StatusOK         Status = "ok"
	StatusSkipUpload Status = "skip" // file already exists
	StatusErr        Status = "error"
)

type UploadDestination struct {
	Status  Status      `json:"status"`
	Error   string      `json:"error,omitempty"`
	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
}
