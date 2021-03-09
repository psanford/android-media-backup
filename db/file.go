package db

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/disintegration/imageorient"
	"github.com/nfnt/resize"
)

var (
	size                   = 512
	startProcessThumbsOnce sync.Once
	thumbReqChan           = make(chan File, 10)
)

func (db *DB) Thumbnail(dbf File) (image.Image, error) {
	log.Printf("get thumb for %s", dbf.Path)
	if db.cacheDir == "" {
		log.Printf("no cache dir!")
		return nil, errors.New("No cachedir found")
	}
	f, err := os.Open(filepath.Join(db.cacheDir, dbf.Name))
	if err != nil {
		startProcessThumbsOnce.Do(func() {
			go processThumbs(db.cacheDir)
		})
		log.Printf("missing thumb for %s, request thumb", dbf.Name)
		select {
		case thumbReqChan <- dbf:
		default:
		}

		myimage := image.NewRGBA(image.Rect(0, 0, size, size)) // x1,y1,  x2,y2
		mygreen := color.RGBA{0, 100, 0, 255}                  //  R, G, B, Alpha

		// backfill entire surface with green
		draw.Draw(myimage, myimage.Bounds(), &image.Uniform{mygreen}, image.ZP, draw.Src)

		return myimage, nil
	}

	log.Printf("thumb found for %s", dbf.Path)

	img, _, err := image.Decode(f)
	log.Printf("decode done for %s", dbf.Path)
	return img, err
}

func processThumbs(cacheDir string) {
	for srcFile := range thumbReqChan {
		func() {
			dstFileName := filepath.Join(cacheDir, srcFile.Name)
			_, err := os.Stat(dstFileName)
			if err == nil {
				log.Printf("thumb already exists for file=%s", srcFile.Path)
				return
			}

			log.Printf("process thumb file=%s", srcFile.Path)

			f, err := os.Open(srcFile.Path)
			if err != nil {
				log.Printf("thumb err open file=%s err=%s", srcFile.Path, err)
				return
			}
			defer f.Close()

			img, _, err := imageorient.Decode(f)
			if err != nil {
				log.Printf("thumb err process file=%s err=%s", srcFile.Path, err)
				return
			}

			img = resize.Thumbnail(uint(size), uint(size), img, resize.NearestNeighbor)

			tmpFile, err := ioutil.TempFile(cacheDir, srcFile.Name+".tmp")
			if err != nil {
				log.Printf("thumb err open tmp file=%s err=%s", srcFile.Name, err)
				return
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			err = jpeg.Encode(tmpFile, img, nil)
			if err != nil {
				log.Printf("thumb err encode file=%s err=%s", srcFile.Path, err)
				return
			}

			os.Rename(tmpFile.Name(), dstFileName)
		}()
	}
}
