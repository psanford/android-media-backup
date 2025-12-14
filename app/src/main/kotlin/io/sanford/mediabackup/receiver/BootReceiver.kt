package io.sanford.mediabackup.receiver

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log
import io.sanford.mediabackup.worker.UploadWorkerScheduler

/**
 * Receiver that schedules the upload worker on device boot.
 */
class BootReceiver : BroadcastReceiver() {

    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            Log.d(TAG, "Boot completed, scheduling upload worker")
            UploadWorkerScheduler.schedule(context)
        }
    }

    companion object {
        private const val TAG = "BootReceiver"
    }
}
