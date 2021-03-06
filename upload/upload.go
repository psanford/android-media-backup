package upload

import (
	"io/ioutil"
	"path/filepath"

	"github.com/psanford/android-media-backup-go-experiment/db"
	"github.com/psanford/android-media-backup-go-experiment/ui/plog"
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
		} else {
			plog.Printf("bgjob %s in db, state is %s", filename, dbFile.State)
		}

	}

	return nil
}
