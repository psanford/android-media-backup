package io.sanford.mediabackup.ui

import android.text.format.DateUtils
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp

@Composable
fun SettingsScreen(
    viewModel: MainViewModel,
    modifier: Modifier = Modifier
) {
    val config by viewModel.config.collectAsState()
    val stats by viewModel.stats.collectAsState()
    val isUploading by viewModel.isUploading.collectAsState()

    // Refresh stats when this screen becomes visible
    LaunchedEffect(Unit) {
        viewModel.refresh()
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text(
            text = "Media Backup",
            style = MaterialTheme.typography.headlineMedium
        )

        OutlinedTextField(
            value = config.url,
            onValueChange = { viewModel.setUrl(it) },
            label = { Text("Server URL") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        OutlinedTextField(
            value = config.username,
            onValueChange = { viewModel.setUsername(it) },
            label = { Text("Username") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        OutlinedTextField(
            value = config.password,
            onValueChange = { viewModel.setPassword(it) },
            label = { Text("Password") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
            visualTransformation = PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("Enable Automatic Backups")
            Switch(
                checked = config.enabled,
                onCheckedChange = { viewModel.setEnabled(it) }
            )
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("WiFi Only")
            Switch(
                checked = config.wifiOnly,
                onCheckedChange = { viewModel.setWifiOnly(it) }
            )
        }

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "Status",
            style = MaterialTheme.typography.titleMedium
        )

        StatRow("Last Sync Check", formatTime(stats.lastSyncTimeMs))
        StatRow("Last Upload", formatTime(stats.lastUploadTimeMs))
        StatRow("Pending Uploads", stats.pendingUploads.toString())
        StatRow("Uploads (30 days)", stats.recentUploads.toString())
        StatRow("Failures (30 days)", stats.recentFailedUploads.toString())

        Spacer(modifier = Modifier.height(8.dp))

        Button(
            onClick = { viewModel.triggerUpload() },
            enabled = config.enabled && !isUploading,
            modifier = Modifier.fillMaxWidth()
        ) {
            Text(if (isUploading) "Uploading..." else "Test Upload")
        }

        Button(
            onClick = { viewModel.resetFailedUploads() },
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("Reset Failed Uploads")
        }

        Button(
            onClick = { viewModel.resetDatabase() },
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("Reset Full DB State")
        }
    }
}

@Composable
private fun StatRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(label)
        Text(value)
    }
}

private fun formatTime(timeMs: Long): String {
    return if (timeMs <= 0) {
        "Never"
    } else {
        DateUtils.getRelativeTimeSpanString(
            timeMs,
            System.currentTimeMillis(),
            DateUtils.MINUTE_IN_MILLIS
        ).toString()
    }
}
