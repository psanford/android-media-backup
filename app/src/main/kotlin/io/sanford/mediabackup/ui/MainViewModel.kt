package io.sanford.mediabackup.ui

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import io.sanford.mediabackup.data.BackupConfig
import io.sanford.mediabackup.data.BackupRepository
import io.sanford.mediabackup.data.BackupStats
import io.sanford.mediabackup.data.MediaFile
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class MainViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = BackupRepository(application)

    private val _config = MutableStateFlow(BackupConfig())
    val config: StateFlow<BackupConfig> = _config.asStateFlow()

    private val _stats = MutableStateFlow(BackupStats())
    val stats: StateFlow<BackupStats> = _stats.asStateFlow()

    private val _files = MutableStateFlow<List<MediaFile>>(emptyList())
    val files: StateFlow<List<MediaFile>> = _files.asStateFlow()

    private val _isUploading = MutableStateFlow(false)
    val isUploading: StateFlow<Boolean> = _isUploading.asStateFlow()

    private val _logMessages = MutableStateFlow<List<String>>(emptyList())
    val logMessages: StateFlow<List<String>> = _logMessages.asStateFlow()

    init {
        refresh()
    }

    fun refresh() {
        viewModelScope.launch {
            _config.value = repository.getConfig()
            _stats.value = repository.getStats()
            repository.scanFiles()
            _files.value = repository.getFiles()

            // Fetch any pending logs from refresh operations
            val logs = withContext(Dispatchers.IO) {
                repository.getPendingLogs()
            }
            if (logs.isNotEmpty()) {
                logs.lines().filter { it.isNotBlank() }.forEach { addLog(it) }
            }
        }
    }

    fun setEnabled(enabled: Boolean) {
        viewModelScope.launch {
            val newConfig = _config.value.copy(enabled = enabled)
            _config.value = newConfig
            repository.setConfig(newConfig)
        }
    }

    fun setUrl(url: String) {
        viewModelScope.launch {
            val newConfig = _config.value.copy(url = url)
            _config.value = newConfig
            repository.setConfig(newConfig)
        }
    }

    fun setUsername(username: String) {
        viewModelScope.launch {
            val newConfig = _config.value.copy(username = username)
            _config.value = newConfig
            repository.setConfig(newConfig)
        }
    }

    fun setPassword(password: String) {
        viewModelScope.launch {
            val newConfig = _config.value.copy(password = password)
            _config.value = newConfig
            repository.setConfig(newConfig)
        }
    }

    fun setWifiOnly(wifiOnly: Boolean) {
        viewModelScope.launch {
            val newConfig = _config.value.copy(wifiOnly = wifiOnly)
            _config.value = newConfig
            repository.setConfig(newConfig)
        }
    }

    fun triggerUpload() {
        if (_isUploading.value) return

        viewModelScope.launch {
            _isUploading.value = true
            addLog("Starting upload...")

            // Start a coroutine to poll for log messages while uploading
            val logJob = launch {
                while (_isUploading.value) {
                    val logs = withContext(Dispatchers.IO) {
                        repository.getPendingLogs()
                    }
                    if (logs.isNotEmpty()) {
                        logs.lines().filter { it.isNotBlank() }.forEach { addLog(it) }
                    }
                    delay(100)
                }
            }

            val result = withContext(Dispatchers.IO) {
                repository.triggerUpload()
            }
            result.fold(
                onSuccess = { addLog("Upload completed") },
                onFailure = { addLog("Upload failed: ${it.message}") }
            )

            _isUploading.value = false
            logJob.cancel()

            // Get any remaining logs
            val remainingLogs = withContext(Dispatchers.IO) {
                repository.getPendingLogs()
            }
            if (remainingLogs.isNotEmpty()) {
                remainingLogs.lines().filter { it.isNotBlank() }.forEach { addLog(it) }
            }

            refresh()
        }
    }

    fun resetFailedUploads() {
        viewModelScope.launch {
            addLog("Resetting failed uploads...")
            val result = repository.resetFailedUploads()
            result.fold(
                onSuccess = { addLog("Failed uploads reset") },
                onFailure = { addLog("Reset failed: ${it.message}") }
            )
            refresh()
        }
    }

    fun resetDatabase() {
        viewModelScope.launch {
            addLog("Resetting database...")
            val result = repository.resetDatabase()
            result.fold(
                onSuccess = { addLog("Database reset") },
                onFailure = { addLog("Reset failed: ${it.message}") }
            )
            refresh()
        }
    }

    private fun addLog(message: String) {
        val timestamp = java.text.SimpleDateFormat("HH:mm:ss", java.util.Locale.getDefault())
            .format(java.util.Date())
        _logMessages.value = _logMessages.value + "[$timestamp] $message"
    }
}
