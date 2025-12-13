// Package mobile provides the API for the Android app to interact with
// the Go backend. This package is compiled with gomobile bind to create
// an AAR library.
package mobile

import (
	"encoding/json"
	"log"
	"time"

	"github.com/psanford/android-media-backup/db"
	"github.com/psanford/android-media-backup/network"
	"github.com/psanford/android-media-backup/upload"
)

// Network state constants matching network.ConnState
const (
	NetworkStateUnknown = int(network.ConnStateUnknown)
	NetworkNoNetwork    = int(network.NoNetwork)
	NetworkNotWifi      = int(network.NotWifi)
	NetworkWifi         = int(network.Wifi)
)

// Init initializes the Go backend with the required directories.
// Must be called before any other function.
func Init(dataDir, cacheDir string) {
	db.SetDirectories(dataDir, cacheDir)
	log.Printf("mobile: initialized with dataDir=%s cacheDir=%s", dataDir, cacheDir)
}

// SetNetworkState sets the current network connection state.
// Use the NetworkState* constants.
func SetNetworkState(state int) {
	network.SetConnectionState(network.ConnState(state))
}

// Config represents the application configuration.
type Config struct {
	Enabled           bool
	URL               string
	Username          string
	Password          string
	AllowMobileUpload bool
}

// Stats represents upload statistics.
type Stats struct {
	LastSyncTimeMS      int64
	LastUploadTimeMS    int64
	PendingUploads      int
	RecentUploads       int
	RecentFailedUploads int
}

// File represents a media file with its upload state.
type File struct {
	Name            string
	Path            string
	CreatedMS       int64
	UploadStartedMS int64
	UploadEndMS     int64
	Size            int64
	State           int
	StateString     string
}

// EventCallback is the interface for receiving events from the Go backend.
type EventCallback interface {
	OnLog(message string)
	OnUploadProgress(filename string, state int)
	OnStatsUpdated()
}

var eventCallback EventCallback

// SetEventCallback sets the callback for receiving events.
func SetEventCallback(cb EventCallback) {
	eventCallback = cb
}

func logEvent(msg string) {
	if eventCallback != nil {
		eventCallback.OnLog(msg)
	}
	log.Println(msg)
}

// GetConfig returns the current configuration.
func GetConfig() (*Config, error) {
	store, err := db.Open()
	if err != nil {
		return nil, err
	}

	enabled, _ := store.Enabled()
	url, _ := store.URL()
	username, _ := store.Username()
	password, _ := store.Password()
	allowMobile, _ := store.AllowMobileUpload()

	return &Config{
		Enabled:           enabled,
		URL:               url,
		Username:          username,
		Password:          password,
		AllowMobileUpload: allowMobile,
	}, nil
}

// SetConfig saves the configuration.
func SetConfig(cfg *Config) error {
	store, err := db.Open()
	if err != nil {
		return err
	}

	if err := store.SetEnabled(cfg.Enabled); err != nil {
		return err
	}
	if err := store.SetURL(cfg.URL); err != nil {
		return err
	}
	if err := store.SetUsername(cfg.Username); err != nil {
		return err
	}
	if err := store.SetPassword(cfg.Password); err != nil {
		return err
	}
	if err := store.SetAllowMobileUpload(cfg.AllowMobileUpload); err != nil {
		return err
	}

	return nil
}

// SetEnabled sets the enabled state.
func SetEnabled(enabled bool) error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.SetEnabled(enabled)
}

// SetURL sets the server URL.
func SetURL(url string) error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.SetURL(url)
}

// SetUsername sets the username.
func SetUsername(username string) error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.SetUsername(username)
}

// SetPassword sets the password.
func SetPassword(password string) error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.SetPassword(password)
}

// SetAllowMobileUpload sets whether mobile upload is allowed.
func SetAllowMobileUpload(allow bool) error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.SetAllowMobileUpload(allow)
}

// GetStats returns upload statistics.
func GetStats() (*Stats, error) {
	store, err := db.Open()
	if err != nil {
		return nil, err
	}

	lastSync, _ := store.LastCheckTime()
	lastUpload, _ := store.LastFileUpload()
	pending, _ := store.PendingUploads()
	recent, _ := store.UploadsSince(time.Now().Add(-30*24*time.Hour), db.UploadSuccess)
	recentFailed, _ := store.UploadsSince(time.Now().Add(-30*24*time.Hour), db.UploadFailed)

	return &Stats{
		LastSyncTimeMS:      lastSync.UnixMilli(),
		LastUploadTimeMS:    lastUpload.UnixMilli(),
		PendingUploads:      pending,
		RecentUploads:       recent,
		RecentFailedUploads: recentFailed,
	}, nil
}

// GetFilesJSON returns all tracked files as JSON.
// Using JSON because gomobile doesn't support slices of structs.
func GetFilesJSON() (string, error) {
	store, err := db.Open()
	if err != nil {
		return "", err
	}

	dbFiles, err := store.GetFiles()
	if err != nil {
		return "", err
	}

	files := make([]File, len(dbFiles))
	for i, f := range dbFiles {
		files[i] = File{
			Name:            f.Name,
			Path:            f.Path,
			CreatedMS:       f.Created.UnixMilli(),
			UploadStartedMS: f.UploadStarted.UnixMilli(),
			UploadEndMS:     f.UploadEnd.UnixMilli(),
			Size:            f.Size,
			State:           int(f.State),
			StateString:     f.State.String(),
		}
	}

	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// TriggerUpload starts the upload process.
func TriggerUpload() error {
	logEvent("Starting upload...")
	err := upload.Upload()
	if err != nil {
		logEvent("Upload error: " + err.Error())
		return err
	}
	logEvent("Upload completed")
	if eventCallback != nil {
		eventCallback.OnStatsUpdated()
	}
	return nil
}

// ScanFiles scans for new media files.
func ScanFiles() error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	_, _, err = upload.ScanFiles(store)
	return err
}

// ResetFailedUploads resets failed uploads to pending.
func ResetFailedUploads() error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.ResetFailedUploads()
}

// ResetDatabase clears all file tracking data.
func ResetDatabase() error {
	store, err := db.Open()
	if err != nil {
		return err
	}
	return store.ResetFiles()
}

// GetThumbnailPath returns the path to a file's thumbnail, or empty if not available.
func GetThumbnailPath(filename string) string {
	store, err := db.Open()
	if err != nil {
		return ""
	}

	// The thumbnail is stored in the cache directory with the same name
	if store.CacheDir() == "" {
		return ""
	}

	return store.CacheDir() + "/" + filename
}
