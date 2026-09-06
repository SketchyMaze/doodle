package com.sketchymaze.doodle;

import android.content.Context;
import android.content.pm.ActivityInfo;
import android.system.Os;
import android.util.Log;

import org.libsdl.app.SDLActivity;
import org.libsdl.app.SDLSurface;

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
}
