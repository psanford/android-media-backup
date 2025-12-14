package io.sanford.mediabackup.data

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import android.util.Log
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import mobile.Mobile

/**
 * Repository for interacting with the Go backend via gomobile bindings.
 */
class BackupRepository(private val context: Context) {

    private val gson = Gson()

    /**
     * Get the current backup configuration.
     */
    suspend fun getConfig(): BackupConfig = withContext(Dispatchers.IO) {
        try {
            val config = Mobile.getConfig()
            BackupConfig(
                enabled = config.enabled,
                url = config.url,
                username = config.username,
                password = config.password,
                wifiOnly = !config.allowMobileUpload
            )
        } catch (e: Exception) {
            Log.e(TAG, "Error getting config", e)
            BackupConfig()
        }
    }

    /**
     * Update the backup configuration.
     */
    suspend fun setConfig(config: BackupConfig) = withContext(Dispatchers.IO) {
        try {
            Mobile.setEnabled(config.enabled)
            Mobile.setURL(config.url)
            Mobile.setUsername(config.username)
            Mobile.setPassword(config.password)
            Mobile.setAllowMobileUpload(!config.wifiOnly)
        } catch (e: Exception) {
            Log.e(TAG, "Error setting config", e)
        }
    }

    /**
     * Get upload statistics.
     */
    suspend fun getStats(): BackupStats = withContext(Dispatchers.IO) {
        try {
            val stats = Mobile.getStats()
            Log.d(TAG, "getStats: pending=${stats.pendingUploads} recent=${stats.recentUploads}")
            BackupStats(
                lastSyncTimeMs = stats.lastSyncTimeMS,
                lastUploadTimeMs = stats.lastUploadTimeMS,
                pendingUploads = stats.pendingUploads.toInt(),
                recentUploads = stats.recentUploads.toInt(),
                recentFailedUploads = stats.recentFailedUploads.toInt()
            )
        } catch (e: Exception) {
            Log.e(TAG, "Error getting stats", e)
            BackupStats()
        }
    }

    /**
     * Get all tracked media files.
     */
    suspend fun getFiles(): List<MediaFile> = withContext(Dispatchers.IO) {
        try {
            val json = Mobile.getFilesJSON()
            val type = object : TypeToken<List<MediaFile>>() {}.type
            gson.fromJson(json, type) ?: emptyList()
        } catch (e: Exception) {
            Log.e(TAG, "Error getting files", e)
            emptyList()
        }
    }

    /**
     * Trigger an upload, updating network state first.
     */
    suspend fun triggerUpload(): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            updateNetworkState()
            Mobile.triggerUpload()
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e(TAG, "Error triggering upload", e)
            Result.failure(e)
        }
    }

    /**
     * Scan for new files.
     */
    suspend fun scanFiles(): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            Mobile.scanFiles()
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e(TAG, "Error scanning files", e)
            Result.failure(e)
        }
    }

    /**
     * Reset failed uploads to pending.
     */
    suspend fun resetFailedUploads(): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            Mobile.resetFailedUploads()
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e(TAG, "Error resetting failed uploads", e)
            Result.failure(e)
        }
    }

    /**
     * Reset the entire database.
     */
    suspend fun resetDatabase(): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            Mobile.resetDatabase()
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e(TAG, "Error resetting database", e)
            Result.failure(e)
        }
    }

    /**
     * Update the Go backend with current network state.
     */
    fun updateNetworkState() {
        val state = getNetworkState()
        Mobile.setNetworkState(state.toLong())
        Log.d(TAG, "Updated network state: $state")
    }

    private fun getNetworkState(): Int {
        val connectivityManager = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = connectivityManager.activeNetwork ?: return Mobile.NetworkNoNetwork.toInt()
        val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return Mobile.NetworkNoNetwork.toInt()

        return when {
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> Mobile.NetworkWifi.toInt()
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> Mobile.NetworkNotWifi.toInt()
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> Mobile.NetworkWifi.toInt()
            else -> Mobile.NetworkStateUnknown.toInt()
        }
    }

    /**
     * Get pending log messages from the Go backend.
     */
    fun getPendingLogs(): String {
        return try {
            Mobile.getPendingLogs()
        } catch (e: Exception) {
            Log.e(TAG, "Error getting pending logs", e)
            ""
        }
    }

    companion object {
        private const val TAG = "BackupRepository"
    }
}

/**
 * Backup configuration data class.
 */
data class BackupConfig(
    val enabled: Boolean = false,
    val url: String = "",
    val username: String = "",
    val password: String = "",
    val wifiOnly: Boolean = true
)

/**
 * Backup statistics data class.
 */
data class BackupStats(
    val lastSyncTimeMs: Long = 0,
    val lastUploadTimeMs: Long = 0,
    val pendingUploads: Int = 0,
    val recentUploads: Int = 0,
    val recentFailedUploads: Int = 0
)
