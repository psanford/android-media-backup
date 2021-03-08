package io.sanford.jgo;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.util.Log;
import org.gioui.Gio;

public class BootBroadcastReceiver extends BroadcastReceiver {
  static final String ACTION = "android.intent.action.BOOT_COMPLETED";

  @Override
  public void onReceive(Context context, Intent intent) {
    if (intent.getAction().equals(ACTION)) {
      Log.d("sanford android-media-backup", "BootBroadcastReceiver.onReceive()");
      // We need to load Gio in order for us to load the go code which
      // our background worker needs.
      Gio.init(context);
      BackgroundWorker.launchBackgroundWorker(context);
    }
  }
}
