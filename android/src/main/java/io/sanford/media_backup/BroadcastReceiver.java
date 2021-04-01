package io.sanford.media_backup;

import android.content.Context;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.util.Log;
import org.gioui.Gio;

public class BroadcastReceiver extends android.content.BroadcastReceiver {
  static final String BOOT_ACTION = "android.intent.action.BOOT_COMPLETED";

  @Override
  public void onReceive(Context context, Intent intent) {

    String a = intent.getAction();

    if (a.equals(ConnectivityManager.CONNECTIVITY_ACTION)) {
      Log.d("io.sanford.media_backup", "BroadcastReceiver.onReceive() CONNECTIVITY_ACTION");
    } else if (a.equals(BOOT_ACTION)) {
      Log.d("io.sanford.media_backup", "BroadcastReceiver.onReceive() BOOT_ACTION");
    } else {
      return;
    }

    // We need to load Gio in order for us to load the go code which
    // our background worker needs.
    Gio.init(context);
    BackgroundWorker.launchBackgroundWorker(context);
  }
}
