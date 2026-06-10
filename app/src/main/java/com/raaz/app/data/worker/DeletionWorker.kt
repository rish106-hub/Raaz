package com.raaz.app.data.worker

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import com.raaz.app.data.room.RaazDatabase

class DeletionWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {

    override suspend fun doWork(): Result {
        val cutoffMs = System.currentTimeMillis() - 48L * 60 * 60 * 1000
        RaazDatabase.getInstance(applicationContext)
            .chatSessionDao()
            .deleteSessionsOlderThan(cutoffMs)
        return Result.success()
    }
}
