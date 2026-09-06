package com.sketchymaze.doodle;

import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.content.pm.ActivityInfo;
import android.database.Cursor;
import android.net.Uri;
import android.provider.OpenableColumns;
import android.system.Os;
import android.util.Log;

import org.libsdl.app.SDLActivity;
import org.libsdl.app.SDLSurface;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

/**
 * Thin activity around SDL2's own SDLActivity. All of the real app logic is
 * Go+SDL2 code compiled into libdoodle.so (see cmd/doodle-android in the Go
 * module, and android/scripts/build-native-libs.sh which produces the .so);
 * this class only tells the stock SDL Java glue which native library to load
 * and which exported C function is the entry point.
 */
public class MainActivity extends SDLActivity {

    @Override
    protected String[] getLibraries() {
        return new String[] {
            // c++_shared must load before SDL2_ttf/SDL2_mixer, which were
            // built against it (see android/scripts/build-native-libs.sh).
            "c++_shared",
            "SDL2",
            "SDL2_ttf",
            "SDL2_mixer",
            "doodle"
        };
    }

    @Override
    protected String getMainFunction() {
        return "SDL_main";
    }

    /**
     * Renders into a smaller, density-scaled buffer instead of the phone's
     * raw physical resolution, so the game's fixed-pixel UI doesn't shrink
     * to a third of its intended size on a dense phone panel. See
     * ScaledSDLSurface's own doc comment for the full explanation.
     */
    @Override
    protected SDLSurface createSDLSurface(Context context) {
        return new ScaledSDLSurface(context);
    }

    /**
     * SDL_CreateWindow() calls into this (via native setOrientation() JNI,
     * exposed as the overridable setOrientationBis() below) using the game's
     * *original, un-scaled* requested window size -- balance.Width/Height,
     * 1024x768 -- since that's what's passed to sdl.New() before
     * ScaledSDLSurface ever resizes anything. With a resizable window and no
     * SDL_HINT_ORIENTATIONS hint, the stock implementation resolves that to
     * SCREEN_ORIENTATION_FULL_USER and calls setRequestedOrientation() with
     * it, silently overriding AndroidManifest.xml's
     * android:screenOrientation="fullSensor" at runtime. FULL_USER behaves
     * differently from fullSensor: if the user has auto-rotate off, it locks
     * to whatever orientation the system currently considers "current" --
     * which produced a wrong cold-launch orientation that only
     * self-corrected after the first live device rotation forced a fresh
     * sensor read. Since the game already has its own responsive
     * portrait/landscape UI and the manifest already declares the
     * orientation behavior we want, just keep that instead of letting this
     * heuristic override it.
     */
    @Override
    public void setOrientationBis(int w, int h, boolean resizable, String hint) {
        setRequestedOrientation(ActivityInfo.SCREEN_ORIENTATION_FULL_SENSOR);
    }

    /**
     * Sets $HOME/$XDG_CONFIG_HOME/$XDG_CACHE_HOME before super.loadLibraries()
     * triggers System.loadLibrary("doodle") (and thus dlopen()s libdoodle.so).
     * The idea was for github.com/kirsle/configdir (used by pkg/userdir) to
     * pick these up when its package init() runs at dlopen() time -- but in
     * practice this didn't pan out: even with Os.setenv() below confirmed (by
     * logging) to succeed without throwing, the directories it should have
     * caused pkg/userdir to create never appeared on disk. The likely
     * explanation is that Go's runtime reads its own snapshot of the process
     * environment once at startup rather than calling libc's getenv() live,
     * so it never observed this change regardless of ordering.
     *
     * pkg/userdir's paths are now set directly instead, from Go, via
     * cmd/doodle-android/main_android.go's setupUserdir() (using
     * SDL_AndroidGetInternalStoragePath() through JNI, sidestepping
     * environment variables entirely) -- see android/README.md's "Save
     * data / settings / user-created levels" section for the full story.
     * These setenv() calls are left in place as a harmless fallback for
     * anything else that might consult $HOME, but nothing in this codebase
     * is known to depend on them working.
     */
    @Override
    public void loadLibraries() {
        try {
            String home = getFilesDir().getAbsolutePath();
            Log.v("SDL", "MainActivity.loadLibraries(): setting HOME=" + home);
            Os.setenv("HOME", home, true);
            Os.setenv("XDG_CONFIG_HOME", home + "/.config", true);
            Os.setenv("XDG_CACHE_HOME", home + "/.cache", true);
            Log.v("SDL", "MainActivity.loadLibraries(): setenv calls completed without throwing");
        } catch (Throwable e) {
            Log.e("SDL", "MainActivity.loadLibraries(): setenv failed", e);
        }

        super.loadLibraries();
    }

    // ------------------------------------------------------------------
    // Native file dialogs (pkg/native's OpenFile/SaveFile on Android).
    // See pkg/native/android/filepicker.go for the Go/JNI side that calls
    // into showFilePicker() below and blocks waiting for
    // nativeFilePickerResult() to be called back.
    // ------------------------------------------------------------------

    private static final int REQUEST_CODE_FILE_PICKER = 1001;
    private boolean mFilePickerForSave;

    /**
     * Launches Android's Storage Access Framework document picker. Called
     * from Go via JNI (pkg/native/android/filepicker.go's launchPicker()),
     * never directly from other Java code.
     */
    public void showFilePicker(boolean forSave, String suggestedName) {
        mFilePickerForSave = forSave;

        Intent intent = new Intent(forSave ? Intent.ACTION_CREATE_DOCUMENT : Intent.ACTION_OPEN_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        // SAF filters by MIME type, and this game's own extensions
        // (.level, .doodad) have no registered MIME type to filter by, so
        // show every file type rather than trying to translate pkg/native's
        // "*.ext *.ext2"-style filter string into one (see PickFile's own
        // doc comment on the Go side).
        intent.setType("*/*");
        if (forSave && suggestedName != null && !suggestedName.isEmpty()) {
            intent.putExtra(Intent.EXTRA_TITLE, suggestedName);
        }

        try {
            startActivityForResult(intent, REQUEST_CODE_FILE_PICKER);
        } catch (Exception e) {
            Log.e("SDL", "showFilePicker: couldn't launch picker intent", e);
            nativeFilePickerResult(null, false);
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);

        if (requestCode != REQUEST_CODE_FILE_PICKER) {
            return;
        }

        if (resultCode != Activity.RESULT_OK || data == null || data.getData() == null) {
            nativeFilePickerResult(null, false); // user canceled
            return;
        }

        Uri uri = data.getData();
        if (mFilePickerForSave) {
            // Not reachable today -- pkg/native/file_dialog_android.go's
            // SaveFile() returns an error before ever calling PickFile(),
            // since actually writing the caller's bytes back out to this
            // content:// URI isn't wired up yet. Report failure here too
            // rather than a local path that wouldn't correspond to
            // anywhere real, in case that ever changes without this branch
            // being revisited.
            Log.w("SDL", "showFilePicker: ACTION_CREATE_DOCUMENT result isn't handled yet");
            nativeFilePickerResult(null, false);
            return;
        }

        String localPath = copyUriToCache(uri);
        nativeFilePickerResult(localPath, localPath != null);
    }

    /**
     * Copies a content:// URI's bytes into a real file under this app's
     * cache directory, named after the picked file's own display name
     * where available. Native Go code can't read a content:// URI
     * directly the way it can a real filesystem path via os.Open(), so
     * this is what makes the picked file actually usable on the Go side.
     * Returns the local absolute path, or null on failure.
     */
    private String copyUriToCache(Uri uri) {
        String displayName = queryDisplayName(uri);
        if (displayName == null || displayName.isEmpty()) {
            displayName = "picked-" + System.currentTimeMillis();
        }

        File outFile = new File(getCacheDir(), displayName);
        try (InputStream in = getContentResolver().openInputStream(uri);
             OutputStream out = new FileOutputStream(outFile)) {
            if (in == null) {
                return null;
            }
            byte[] buffer = new byte[8192];
            int read;
            while ((read = in.read(buffer)) != -1) {
                out.write(buffer, 0, read);
            }
            return outFile.getAbsolutePath();
        } catch (IOException e) {
            Log.e("SDL", "copyUriToCache: failed to copy " + uri, e);
            return null;
        }
    }

    private String queryDisplayName(Uri uri) {
        try (Cursor cursor = getContentResolver().query(uri, null, null, null, null)) {
            if (cursor != null && cursor.moveToFirst()) {
                int idx = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
                if (idx >= 0) {
                    return cursor.getString(idx);
                }
            }
        } catch (Exception e) {
            Log.w("SDL", "queryDisplayName: couldn't query " + uri, e);
        }
        return null;
    }

    /**
     * Implemented in cmd/doodle-android/main_android.go (exported as
     * Java_com_sketchymaze_doodle_MainActivity_nativeFilePickerResult):
     * delivers showFilePicker()'s result back to the Go side, unblocking
     * pkg/native/android's PickFile().
     */
    private native void nativeFilePickerResult(String path, boolean ok);
}
