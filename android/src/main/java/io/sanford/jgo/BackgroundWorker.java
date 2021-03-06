package io.sanford.jgo;

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
  public BackgroundWorker(Context context, WorkerParameters params) {
    super(context, params);
    Log.d("sanford android-media-backup", "BackgroundWorker()");
  }

  public Result doWork() {
    Log.d("sanford android-media-backup", "start runBackgroundJob()");
    runBackgroundJob();
    Log.d("sanford android-media-backup", "complete runBackgroundJob()");

    return Result.success();
  }

  static void launchBackgroundWorker(Context context) {
    Log.d("sanford android-media-backup", "LaunchBackgroundWorker");
    Constraints constraints = new Constraints.Builder()
      .setRequiresBatteryNotLow(true)
      .setRequiredNetworkType(NetworkType.CONNECTED)
      .build();
      // NetworkType.UNMETERED would be for wifi only.
      // we don't use that because you might want to run on the
      // mobile network
      // We might also want to consider an option for
      // .setRequiresCharging(true)


    PeriodicWorkRequest workRequest = new PeriodicWorkRequest.Builder(BackgroundWorker.class, 15, TimeUnit.MINUTES)
      .setConstraints(constraints)
      .build();

    WorkManager.getInstance().enqueue(workRequest);
    Log.d("sanford android-media-backup", "LaunchBackgroundWorker DONE");
  }

  static private native void runBackgroundJob();
}
