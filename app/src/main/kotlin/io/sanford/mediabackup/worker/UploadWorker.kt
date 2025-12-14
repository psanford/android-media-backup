package io.sanford.mediabackup.worker

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import android.os.PowerManager
import android.util.Log
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import mobile.Mobile

/**
 * Background worker that performs media uploads.
 */
class UploadWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    private var wakeLock: PowerManager.WakeLock? = null

    override suspend fun doWork(): Result {
        Log.d(TAG, "Starting upload work")

        try {
            // Acquire wake lock to prevent CPU from sleeping
            acquireWakeLock()

            // Update network state before upload
            updateNetworkState()

            // Trigger the upload
            Mobile.triggerUpload()

            Log.d(TAG, "Upload work completed successfully")
            return Result.success()
        } catch (e: Exception) {
            Log.e(TAG, "Upload work failed", e)
            return Result.retry()
        } finally {
            releaseWakeLock()
        }
    }

    private fun acquireWakeLock() {
        try {
            val powerManager = applicationContext.getSystemService(Context.POWER_SERVICE) as PowerManager
            wakeLock = powerManager.newWakeLock(
                PowerManager.PARTIAL_WAKE_LOCK,
                "MediaBackup::UploadWakeLock"
            ).apply {
                acquire(30 * 60 * 1000L) // 30 minutes max
            }
            Log.d(TAG, "Wake lock acquired")
        } catch (e: Exception) {
            Log.w(TAG, "Could not acquire wake lock: ${e.message}")
        }
    }

    private fun releaseWakeLock() {
        try {
            wakeLock?.let {
                if (it.isHeld) {
                    it.release()
                    Log.d(TAG, "Wake lock released")
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Error releasing wake lock: ${e.message}")
        }
        wakeLock = null
    }

    private fun updateNetworkState() {
        val connectivityManager = applicationContext.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = connectivityManager.activeNetwork
        val state = if (network == null) {
            Mobile.NetworkNoNetwork
        } else {
            val capabilities = connectivityManager.getNetworkCapabilities(network)
            when {
                capabilities == null -> Mobile.NetworkNoNetwork
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> Mobile.NetworkWifi
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> Mobile.NetworkNotWifi
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> Mobile.NetworkWifi
                else -> Mobile.NetworkStateUnknown
            }
        }
        Mobile.setNetworkState(state)
        Log.d(TAG, "Updated network state: $state")
    }

    companion object {
        private const val TAG = "UploadWorker"
    }
}
