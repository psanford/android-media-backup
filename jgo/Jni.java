package io.sanford.jgo;

import java.lang.String;
import android.content.pm.PackageManager;
import android.Manifest;
import android.os.Handler;
import android.app.Activity;
import android.app.Fragment;
import android.app.FragmentTransaction;
import android.util.Log;
import android.view.View;
import android.content.Context;


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
    if (ctx.checkSelfPermission(Manifest.permission.READ_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
      requestPermissions(new String[]{Manifest.permission.READ_EXTERNAL_STORAGE}, PERMISSION_REQUEST);
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
  public void onRequestPermissionsResult (int requestCode, String[] permissions, int[] grantResults) {
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

  static private native void permissionResult(boolean allowed);
}
