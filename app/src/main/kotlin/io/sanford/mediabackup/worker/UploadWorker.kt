package io.sanford.mediabackup.worker

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
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

    override suspend fun doWork(): Result {
        Log.d(TAG, "Starting upload work")

        try {
            // Update network state before upload
            updateNetworkState()

            // Trigger the upload
            Mobile.triggerUpload()

            Log.d(TAG, "Upload work completed successfully")
            return Result.success()
        } catch (e: Exception) {
            Log.e(TAG, "Upload work failed", e)
            return Result.retry()
        }
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
