package io.sanford.media_backup;

import java.util.concurrent.TimeUnit;
import android.content.Context;
import android.util.Log;
import androidx.work.Constraints;
import androidx.work.NetworkType;
import androidx.work.ListenableWorker.Result;
import androidx.work.Worker;
import androidx.work.WorkManager;
import androidx.work.WorkerParameters;
import androidx.work.PeriodicWorkRequest;

public class BackgroundWorker extends Worker {
   private static final String WORKER_TAG = BackgroundWorker.class.getSimpleName();

  public BackgroundWorker(Context context, WorkerParameters params) {
    super(context, params);
  }

  public Result doWork() {
    Log.d("io.sanford.media_backup", "start runBackgroundJob()");
    runBackgroundJob();
    Log.d("io.sanford.media_backup", "complete runBackgroundJob()");

    return Result.success();
  }

  static void launchBackgroundWorker(Context context) {
    Log.d("io.sanford.media_backup", "LaunchBackgroundWorker");
    Constraints constraints = new Constraints.Builder()
      .setRequiresBatteryNotLow(true)
      .setRequiredNetworkType(NetworkType.CONNECTED)
      .build();
      // NetworkType.UNMETERED would be for wifi only.
      // we don't use that because you might want to run on the
      // mobile network
      // We might also want to consider an option for
      // .setRequiresCharging(true)

    // cancel any existing workers
    WorkManager.getInstance().cancelAllWork();

    PeriodicWorkRequest workRequest = new PeriodicWorkRequest.Builder(BackgroundWorker.class, 15, TimeUnit.MINUTES)
      .addTag(BackgroundWorker.WORKER_TAG)
      .setConstraints(constraints)
      .build();

    WorkManager.getInstance().enqueue(workRequest);
  }

  static private native void runBackgroundJob();
}
