package io.sanford.mediabackup.worker

import android.content.Context
import android.util.Log
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import java.util.concurrent.TimeUnit

/**
 * Scheduler for the periodic upload worker.
 */
object UploadWorkerScheduler {

    private const val TAG = "UploadWorkerScheduler"
    private const val WORK_NAME = "media_backup_upload"

    /**
     * Schedule the periodic upload worker.
     */
    fun schedule(context: Context) {
        Log.d(TAG, "Scheduling upload worker")

        val constraints = Constraints.Builder()
            .setRequiredNetworkType(NetworkType.CONNECTED)
            .setRequiresBatteryNotLow(true)
            .build()

        val uploadRequest = PeriodicWorkRequestBuilder<UploadWorker>(
            15, TimeUnit.MINUTES
        )
            .setConstraints(constraints)
            .addTag(WORK_NAME)
            .build()

        WorkManager.getInstance(context).enqueueUniquePeriodicWork(
            WORK_NAME,
            ExistingPeriodicWorkPolicy.KEEP,
            uploadRequest
        )

        Log.d(TAG, "Upload worker scheduled")
    }

    /**
     * Cancel the periodic upload worker.
     */
    fun cancel(context: Context) {
        Log.d(TAG, "Cancelling upload worker")
        WorkManager.getInstance(context).cancelUniqueWork(WORK_NAME)
    }
}
