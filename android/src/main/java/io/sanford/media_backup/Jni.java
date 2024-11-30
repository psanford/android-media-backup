package io.sanford.media_backup;

import android.Manifest;
import android.app.Activity;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.content.Context;
import android.content.pm.PackageManager;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.Handler;
import android.util.Log;
import android.view.View;
import java.lang.String;


public class Jni extends Fragment {
  final int PERMISSION_REQUEST = 1;

  public Jni() {
    Log.d("gio", "Jni()");
  }

  public void register(View view) {
    Log.d("gio", "Jni: register()");
    Context ctx = view.getContext();
    Handler handler = new Handler(ctx.getMainLooper());
    Jni inst = this;
    handler.post(new Runnable() {
        public void run() {
          Activity act = (Activity)ctx;
          FragmentTransaction ft = act.getFragmentManager().beginTransaction();
          ft.add(inst, "Jni");
          ft.commitNow();
        }
      });
  }

  @Override public void onAttach(Context ctx) {
    super.onAttach(ctx);
    Log.d("gio", "jni: onAttach()");
    if (ctx.checkSelfPermission(Manifest.permission.READ_MEDIA_IMAGES) != PackageManager.PERMISSION_GRANTED) {
      requestPermissions(new String[]{Manifest.permission.READ_MEDIA_IMAGES, Manifest.permission.READ_MEDIA_VIDEO, Manifest.permission.ACCESS_MEDIA_LOCATION}, PERMISSION_REQUEST);
    } else {
      permissionResult(true);
    }
  }

  @Override
  public void onDestroy() {
    Log.d("gio","onDestroy()");
    super.onDestroy();
  }

  @Override
  public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
    Log.d("gio", "Jni: onRequestPermissionsResult");
    if (requestCode == PERMISSION_REQUEST) {
      boolean granted = true;
      for (int x : grantResults) {
        if (x == PackageManager.PERMISSION_DENIED) {
          granted = false;
          break;
        }
      }
      if (!granted) {
        Log.d("gio", "Jni: permissions not granted");
      } else{
        Log.d("gio", "Jni: permissions granted");
      }

      permissionResult(granted);
    }
  }

  static int connectionState(Context ctx) {
    ConnectivityManager cs = (ConnectivityManager) ctx.getSystemService(Context.CONNECTIVITY_SERVICE);

		NetworkInfo info = cs.getActiveNetworkInfo();
		if (info == null || !info.isConnected()) {
			Log.d("gio", "No network connection");
			return -1;
		}

    if (ConnectivityManager.TYPE_WIFI == info.getType()) {
      return 1;
    }

    return 0;
  }

  static private native void permissionResult(boolean allowed);
}
