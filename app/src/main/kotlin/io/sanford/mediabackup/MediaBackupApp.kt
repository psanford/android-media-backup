package io.sanford.mediabackup

import android.app.Application
import android.util.Log
import io.sanford.mediabackup.worker.UploadWorkerScheduler
import mobile.Mobile

class MediaBackupApp : Application() {

    override fun onCreate() {
        super.onCreate()

        // Initialize Go backend with app directories
        val dataDir = filesDir.absolutePath
        val cacheDir = cacheDir.absolutePath

        Log.d(TAG, "Initializing Go backend: dataDir=$dataDir, cacheDir=$cacheDir")
        Mobile.init(dataDir, cacheDir)

        // Schedule background upload worker
        UploadWorkerScheduler.schedule(this)
    }

    companion object {
        private const val TAG = "MediaBackupApp"
    }
}
