# Android Media Backup - Migration Plan

## Part 1: Current Application Documentation

### Overview

Android Media Backup is a self-hosted photo/video backup application written in Go using the Gio UI framework. It automatically backs up media files from the device's camera directory to a configured server endpoint.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Android APK                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐    ┌──────────────────────────────────┐  │
│  │   Java Layer     │    │           Go Layer               │  │
│  │                  │    │                                  │  │
│  │  - Jni.java      │◄──►│  main.go ─► ui/ui.go            │  │
│  │    (permissions) │JNI │             │                    │  │
│  │                  │    │             ▼                    │  │
│  │  - Background    │◄──►│  upload/upload.go               │  │
│  │    Worker.java   │    │             │                    │  │
│  │    (scheduling)  │    │             ▼                    │  │
│  │                  │    │  db/db.go ◄─┘                    │  │
│  │  - Broadcast     │    │  (SQLite)                        │  │
│  │    Receiver.java │    │                                  │  │
│  │    (boot/network)│    │  jgo/*.go (JNI bridge)          │  │
│  └──────────────────┘    └──────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                     Gio AAR Library                             │
│                 (compiled Go + Gio runtime)                     │
└─────────────────────────────────────────────────────────────────┘
```

### Features

#### 1. Configuration Settings
- **Server URL**: HTTP(S) endpoint for upload requests
- **Username/Password**: HTTP Basic authentication credentials
- **Enable/Disable**: Master switch for automatic backups
- **WiFi-Only Mode**: Restrict uploads to WiFi connections only

#### 2. Automatic Media Backup
- Scans `/sdcard/DCIM/Camera` directory for photos and videos
- Detects new files automatically via periodic scanning
- Generates SHA256 hash for each file as unique identifier
- Detects content type from file headers (MIME type detection)
- Supports server-side deduplication via "skip" response

#### 3. Background Scheduling
- Uses AndroidX WorkManager for periodic background tasks
- 15-minute interval between checks
- Constraints:
  - Battery not low
  - Network connected
- Survives device boot via BroadcastReceiver

#### 4. Upload Protocol
Request flow:
1. POST metadata to server (JSON):
   ```json
   {
     "id": "<sha256-hash>",
     "name": "IMG_20241213_123456.jpg",
     "mtime": "2024-12-13T12:34:56Z",
     "size": 1234567,
     "content_type": "image/jpeg"
   }
   ```
2. Server responds with upload destination:
   ```json
   {
     "status": "ok",
     "url": "https://s3.example.com/presigned-url",
     "method": "PUT",
     "headers": {"x-custom-header": "value"}
   }
   ```
3. Upload file body to provided URL
4. Server can respond "skip" if file already exists

#### 5. File State Management
States: `Pending` → `InProgress` → `Success`/`Failed`/`Skipped`

| State | Description |
|-------|-------------|
| UploadPending (1) | New file, ready to upload |
| UploadInProgress (2) | Currently being uploaded |
| UploadSuccess (3) | Successfully uploaded |
| UploadSkipped (4) | Server indicated file exists |
| UploadFailed (5) | Upload failed |
| UploadFileDeleted (6) | File was deleted before upload |

#### 6. Thumbnail Generation
- Lazy-loads thumbnails on demand
- Uses EXIF orientation for correct rotation
- 512x512 pixel thumbnails
- LRU memory cache (10 items)
- Disk cache in app cache directory

#### 7. User Interface (Gio)
Three tabs with sliding animation:
1. **Settings Tab**:
   - Server URL, username, password inputs
   - Enable automatic backups toggle
   - WiFi-only toggle
   - Status displays (last sync, last upload, pending count)
   - Upload statistics (last 30 days)
   - Manual "Test Upload" button
   - "Reset Failed Uploads" button
   - "Reset Full DB State" button

2. **Files Tab**:
   - Scrollable list of tracked files
   - Thumbnail preview for each file
   - File name, creation date, state, upload timestamp

3. **Debug Tab**:
   - Version information
   - Real-time event log

#### 8. Android Permissions
```xml
READ_EXTERNAL_STORAGE    - Read photos/videos
WRITE_EXTERNAL_STORAGE   - Write thumbnail cache
ACCESS_MEDIA_LOCATION    - Access media metadata
INTERNET                 - Upload to server
ACCESS_NETWORK_STATE     - Check connection type (WiFi vs mobile)
RECEIVE_BOOT_COMPLETED   - Start on device boot
```

#### 9. Database Schema (SQLite)

**Table: config**
| Column | Type | Description |
|--------|------|-------------|
| key | TEXT PK | Configuration key |
| val | ANY | Configuration value |

Keys: `enabled`, `url`, `username`, `password`, `allow_mobile_upload`, `last_check_epoch_ms`

**Table: file**
| Column | Type | Description |
|--------|------|-------------|
| name | TEXT PK | Filename |
| created_epoch_ms | INT | File creation time |
| upload_started_epoch_ms | INT | Upload start time |
| upload_end_epoch_ms | INT | Upload completion time |
| size | INT | File size in bytes |
| path | TEXT | Full file path |
| state | INT | UploadState enum value |

### Source Code Structure

```
.
├── main.go                 # Entry point
├── go.mod / go.sum         # Go module definition
├── Makefile                # Build automation
├── shell.nix               # Nix development environment
├── version/
│   └── version.go          # Version string (set at build time)
├── ui/
│   ├── ui.go               # Main UI logic, tabs, settings
│   ├── slider.go           # Tab transition animation
│   └── plog/
│       └── plog.go         # Logging system
├── db/
│   ├── db.go               # Database operations, config storage
│   └── file.go             # Thumbnail generation and caching
├── upload/
│   └── upload.go           # Upload logic, file scanning
├── jgo/
│   ├── jgo.go              # JNI bridge, permissions, background worker
│   ├── androiddir/
│   │   └── androiddir.go   # Android directory access
│   └── wifi/
│       └── wifi.go         # Network state detection
└── android/
    ├── build.gradle        # Android build configuration
    ├── AndroidManifest.xml # Permissions and components
    └── src/main/java/io/sanford/media_backup/
        ├── Jni.java            # Permission request fragment
        ├── BackgroundWorker.java # WorkManager periodic task
        └── BroadcastReceiver.java # Boot/connectivity events
```

### Build System

**Current Process:**
1. `gogio` compiles Go code to Android AAR archive
2. Gradle builds Android project including the AAR
3. Final APK is produced

**Dependencies (shell.nix):**
- Go 1.22 (pinned via nixpkgs)
- OpenJDK 17
- Android SDK (platforms 30, 33)
- Android NDK 21.3.6528147
- Build tools 33.0.2

---

## Part 2: Migration Plan to Native Android UI

### Goals
1. Replace Gio UI with native Android UI (Kotlin + Jetpack Compose or XML layouts)
2. Keep Go backend logic for upload, database, and networking
3. Maintain all existing functionality
4. Use reproducible Nix flakes for build environment
5. Break work into small, reviewable git commits

### Technology Choices

**Recommended Stack:**
- **Language**: Kotlin (modern Android standard)
- **UI Framework**: Jetpack Compose (modern, declarative UI)
- **Architecture**: MVVM with ViewModel and StateFlow
- **Build**: Gradle with Kotlin DSL
- **Go Integration**: gomobile bind to create AAR library

**Alternative (simpler):**
- **UI Framework**: XML layouts with ViewBinding
- If Compose feels too different, XML is more traditional

### Migration Phases

---

### Phase 1: Setup and Infrastructure (Commits 1-4)

#### Commit 1: Convert shell.nix to flake.nix
Create reproducible Nix flakes configuration.

```
Files to add:
- flake.nix
- flake.lock

Files to update:
- .gitignore (add result symlink)
```

**Content of flake.nix:**
- Pin nixpkgs to specific commit
- Include Go 1.22, Android SDK/NDK, JDK 17
- Provide development shell with all tools
- Enable `nix develop` and `nix build` commands

#### Commit 2: Add gomobile and restructure Go code
Prepare Go code for gomobile binding.

```
Files to add:
- mobile/mobile.go (exported API for Android)

Files to update:
- go.mod (add gomobile dependency)
- Makefile (add gomobile bind target)
```

Create a clean Go API that can be called from Kotlin:
- `func GetConfig() *Config`
- `func SetConfig(cfg *Config)`
- `func TriggerUpload()`
- `func GetFiles() []File`
- `func GetStats() *Stats`
- `func ResetFailedUploads()`
- `func ResetDatabase()`
- Event callbacks for progress/logs

#### Commit 3: Create new Android project structure
Set up modern Android project alongside existing code.

```
Files to add:
- app/build.gradle.kts
- app/src/main/AndroidManifest.xml
- app/src/main/kotlin/io/sanford/mediabackup/
    └── MediaBackupApp.kt
- build.gradle.kts (root)
- settings.gradle.kts
- gradle.properties
- gradle/libs.versions.toml (version catalog)
```

**Key configurations:**
- compileSdk 34, targetSdk 34, minSdk 23
- Kotlin 1.9+
- Jetpack Compose BOM
- AndroidX libraries

#### Commit 4: Configure gomobile AAR integration
Build Go code as AAR and integrate with Android project.

```
Files to update:
- Makefile (new targets for gomobile)
- app/build.gradle.kts (include AAR dependency)
- flake.nix (add gomobile tool)
```

---

### Phase 2: Core Android Components (Commits 5-9)

#### Commit 5: Create repository and ViewModel layer
Set up data layer connecting to Go backend.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/
    ├── data/
    │   ├── MediaBackupRepository.kt
    │   └── GoBindings.kt (wrapper for gomobile)
    └── di/
        └── AppModule.kt (if using Hilt/Koin)
```

#### Commit 6: Implement Settings screen
Port settings tab to Compose.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/ui/
    ├── settings/
    │   ├── SettingsScreen.kt
    │   └── SettingsViewModel.kt
    └── theme/
        └── Theme.kt
```

**Features:**
- Server URL, username, password text fields
- Enable backup switch
- WiFi-only switch
- Status displays (last sync, pending count, etc.)
- Action buttons

#### Commit 7: Implement Files screen
Port files tab to Compose.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/ui/files/
    ├── FilesScreen.kt
    ├── FilesViewModel.kt
    └── FileItem.kt (composable for single file)
```

**Features:**
- LazyColumn of files
- Thumbnail loading with Coil
- File metadata display
- Upload state indicators

#### Commit 8: Implement Debug screen
Port debug tab to Compose.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/ui/debug/
    ├── DebugScreen.kt
    └── DebugViewModel.kt
```

**Features:**
- Version display
- Event log (scrollable text)
- Log streaming from Go backend

#### Commit 9: Implement navigation and main activity
Wire up all screens with navigation.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/ui/
    ├── MainActivity.kt
    ├── MainNavigation.kt
    └── MainScreen.kt (tab layout)
```

**Features:**
- Bottom navigation or tabs
- Navigation between Settings, Files, Debug
- App bar with title

---

### Phase 3: Background Services (Commits 10-12)

#### Commit 10: Implement WorkManager for background uploads
Port background scheduling to Kotlin.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/worker/
    ├── UploadWorker.kt
    └── WorkerScheduler.kt
```

**Features:**
- 15-minute periodic work
- Battery and network constraints
- Call Go upload function

#### Commit 11: Implement BroadcastReceiver for boot
Handle device boot to schedule worker.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/receiver/
    └── BootReceiver.kt

Files to update:
- AndroidManifest.xml (register receiver)
```

#### Commit 12: Implement runtime permissions
Handle storage permission requests properly.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/permission/
    └── PermissionHandler.kt
```

**Features:**
- Request READ_EXTERNAL_STORAGE (SDK < 33)
- Request READ_MEDIA_IMAGES, READ_MEDIA_VIDEO (SDK >= 33)
- Handle permission results
- Update UI based on permission state

---

### Phase 4: Refinement and Testing (Commits 13-16)

#### Commit 13: Add thumbnail loading with Coil
Efficient image loading for file list.

```
Files to update:
- app/build.gradle.kts (add Coil dependency)
- FileItem.kt (use AsyncImage)
```

#### Commit 14: Add error handling and user feedback
Improve UX with proper error states.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/ui/components/
    ├── ErrorState.kt
    ├── LoadingState.kt
    └── Snackbar.kt
```

#### Commit 15: Add upload progress notifications
Show foreground notification during uploads.

```
Files to add:
- app/src/main/kotlin/io/sanford/mediabackup/notification/
    └── UploadNotification.kt

Files to update:
- AndroidManifest.xml (notification permission, foreground service)
```

#### Commit 16: Remove Gio code and clean up
Remove unused Gio-specific code.

```
Files to remove:
- ui/ui.go
- ui/slider.go
- ui/plog/plog.go
- jgo/jgo.go (replace with gomobile bindings)
- android/ (old Android project)

Files to update:
- Makefile (remove old targets)
- go.mod (remove gio dependencies)
```

---

### Phase 5: Final Polish (Commits 17-19)

#### Commit 17: Update build documentation
Document new build process.

```
Files to add/update:
- README.md (updated build instructions)
```

#### Commit 18: Add app icon and polish
Add proper launcher icon and branding.

```
Files to add:
- app/src/main/res/mipmap-*/ic_launcher.png
- app/src/main/res/values/strings.xml
- app/src/main/res/values/colors.xml
```

#### Commit 19: Final testing and version bump
Test all functionality and prepare release.

```
Files to update:
- app/build.gradle.kts (bump version)
- CHANGELOG.md (if exists)
```

---

## Summary of Changes by File

### New Files (Native Android)
```
flake.nix
flake.lock
mobile/mobile.go
app/build.gradle.kts
app/src/main/AndroidManifest.xml
app/src/main/kotlin/io/sanford/mediabackup/
    ├── MediaBackupApp.kt
    ├── data/
    │   ├── MediaBackupRepository.kt
    │   └── GoBindings.kt
    ├── ui/
    │   ├── MainActivity.kt
    │   ├── MainScreen.kt
    │   ├── MainNavigation.kt
    │   ├── theme/Theme.kt
    │   ├── settings/
    │   │   ├── SettingsScreen.kt
    │   │   └── SettingsViewModel.kt
    │   ├── files/
    │   │   ├── FilesScreen.kt
    │   │   ├── FilesViewModel.kt
    │   │   └── FileItem.kt
    │   ├── debug/
    │   │   ├── DebugScreen.kt
    │   │   └── DebugViewModel.kt
    │   └── components/
    │       ├── ErrorState.kt
    │       └── LoadingState.kt
    ├── worker/
    │   ├── UploadWorker.kt
    │   └── WorkerScheduler.kt
    ├── receiver/
    │   └── BootReceiver.kt
    ├── permission/
    │   └── PermissionHandler.kt
    └── notification/
        └── UploadNotification.kt
build.gradle.kts (root)
settings.gradle.kts
gradle.properties
gradle/libs.versions.toml
```

### Files to Keep (Go Backend)
```
main.go (modified for gomobile)
db/db.go
db/file.go
upload/upload.go
version/version.go
jgo/wifi/wifi.go (may need adaptation)
jgo/androiddir/androiddir.go (may need adaptation)
go.mod (updated)
go.sum (updated)
```

### Files to Remove (Gio-specific)
```
ui/ui.go
ui/slider.go
ui/plog/plog.go
jgo/jgo.go
android/ (entire old directory)
shell.nix (replaced by flake.nix)
```

---

## Build Commands (After Migration)

```bash
# Enter development environment
nix develop

# Build Go AAR library
make go-aar

# Build Android APK
./gradlew assembleDebug

# Build release APK
./gradlew assembleRelease

# Run tests
./gradlew test

# Full build
make apk
```

---

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| gomobile complexity | Start with simple bindings, expand incrementally |
| Permission changes (SDK 33+) | Test on multiple API levels |
| Background restrictions | Use WorkManager properly, test on various devices |
| Go-Kotlin data marshaling | Define clear interface types, test thoroughly |
| Build system complexity | Document thoroughly, use Nix for reproducibility |

---

## Timeline Estimate

This plan does not include time estimates. The work is broken into small commits that can be done incrementally. Each commit should be independently testable and deployable where possible.
