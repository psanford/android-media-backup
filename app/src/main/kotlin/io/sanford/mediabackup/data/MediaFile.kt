package io.sanford.mediabackup.data

import com.google.gson.annotations.SerializedName

/**
 * Represents a media file tracked by the backup system.
 * This maps to the JSON returned by Mobile.getFilesJSON().
 */
data class MediaFile(
    @SerializedName("Name")
    val name: String,

    @SerializedName("Path")
    val path: String,

    @SerializedName("CreatedMS")
    val createdMs: Long,

    @SerializedName("UploadStartedMS")
    val uploadStartedMs: Long,

    @SerializedName("UploadEndMS")
    val uploadEndMs: Long,

    @SerializedName("Size")
    val size: Long,

    @SerializedName("State")
    val state: Int,

    @SerializedName("StateString")
    val stateString: String
) {
    companion object {
        const val STATE_PENDING = 1
        const val STATE_IN_PROGRESS = 2
        const val STATE_SUCCESS = 3
        const val STATE_SKIPPED = 4
        const val STATE_FAILED = 5
        const val STATE_FILE_DELETED = 6
    }
}
