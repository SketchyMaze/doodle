package com.sketchymaze.doodle;

import android.content.Context;
import android.graphics.Rect;
import android.util.DisplayMetrics;
import android.util.Log;
import android.view.SurfaceHolder;
import android.view.ViewGroup;

import org.libsdl.app.SDLSurface;

/**
 * SketchyMaze's UI (fonts, buttons, the level editor toolbox, ...) is laid
 * out in fixed pixel sizes tuned for a desktop window at ordinary
 * (~96 DPI) monitor density. SDL on Android always reports the *physical*
 * pixel resolution as the window size -- SDL_CreateWindow's requested
 * width/height are silently ignored on this platform (see
 * src/video/android/SDL_androidwindow.c: window->w/h get overwritten from
 * whatever the Java side reports) -- and a modern phone panel is commonly
 * 2.5-3x denser than that, so the same fixed-pixel UI renders comically
 * small. It's the same problem as running an unscaled desktop Linux UI on
 * a 4K/Retina panel.
 *
 * The fix mirrors how desktop HiDPI scaling actually works: render into a
 * smaller buffer sized in device-independent pixels (real pixels /
 * density) and let the compositor upscale it to fill the real screen.
 * SurfaceHolder.setFixedSize() does exactly this for a SurfaceView -- the
 * on-screen area doesn't change, only the resolution of what's drawn into
 * it, so the OS scales the smaller buffer back up to fill the display.
 *
 * This intentionally lives in our own package rather than patching
 * org.libsdl.app.SDLSurface directly: that file is re-copied from
 * upstream SDL2 by ../../scripts/fetch-sdl-sources.sh, which would
 * silently clobber any in-place edit there. Subclassing only relies on
 * SDLSurface's public constructor and its `protected` mWidth/mHeight
 * fields, which are part of its stable-ish extension surface (see
 * createSDLSurface() in SDLActivity, which exists for this purpose).
 *
 * One more wrinkle this constructor works around: SDLActivity.onCreate()
 * adds this view to its RelativeLayout with a bare `addView(mSurface)` --
 * no explicit LayoutParams -- which makes RelativeLayout default it to
 * WRAP_CONTENT instead of filling the screen. That breaks the assumption
 * in surfaceChanged() below that getWidth()/getHeight() are a stable,
 * buffer-size-independent reference to the real screen size: with
 * WRAP_CONTENT, they instead track something close to whatever the
 * *previous* setFixedSize() buffer was, so each call would shrink the
 * buffer a bit more than the last -- visible as the on-screen game getting
 * smaller with every rotation. ViewGroup.addView() only falls back to
 * WRAP_CONTENT defaults when the child has no LayoutParams of its own yet,
 * so setting explicit MATCH_PARENT params here, before SDLActivity ever
 * adds this view to its layout, is enough to fix it without touching
 * SDLActivity.java itself.
 *
 * A last wrinkle: on a live device rotation, by the time surfaceChanged()
 * fires the window/insets/sensor-resolved orientation ("fullSensor" in
 * AndroidManifest.xml) have already fully settled, so getWidth()/
 * getHeight() are reliable there. On a *cold launch*, though, the very
 * first surfaceChanged() call can race ahead of that settling and see a
 * transient, wrong-shaped size -- and nothing else naturally re-checks
 * afterwards. onSizeChanged() is this View's own authoritative "my real
 * measured size just changed" signal (independent of surfaceChanged's
 * timing), so re-verifying there too catches a late correction the first
 * surfaceChanged() call missed.
 */
public class ScaledSDLSurface extends SDLSurface {

    public ScaledSDLSurface(Context context) {
        super(context);
        setLayoutParams(new ViewGroup.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
    }

    /**
     * Computes the density-scaled target buffer size from this view's
     * current real size and, if the surface isn't already fixed to that
     * size, requests it. Returns true if a new size was just requested (the
     * caller should expect another surfaceChanged() callback shortly with
     * it, rather than proceeding now).
     *
     * getWidth()/getHeight() -- this View's own measured on-screen pixel
     * dimensions -- are the reference here rather than two more obvious-
     * looking but wrong alternatives:
     *
     *  - mDisplay.getRealMetrics() reflects the *device's physical rotation
     *    sensor*, not this Activity's actual rendered orientation. Those
     *    disagree whenever the app's orientation is locked (e.g.
     *    android:screenOrientation) to something other than however the
     *    phone is currently being held, which flips width and height
     *    relative to the real SurfaceView shape.
     *  - surfaceChanged()'s own `width`/`height` parameters reflect the
     *    *current buffer* size, which is already our shrunk target after an
     *    earlier setFixedSize() call. Dividing that by density *again* on
     *    a later callback would shrink the buffer a bit more each time.
     *
     * getWidth()/getHeight() avoid both problems: they don't change when
     * setFixedSize() changes the buffer resolution, and (now that the
     * constructor above forces MATCH_PARENT layout) they always match this
     * View's real, currently-rendered on-screen shape.
     */
    private boolean requestScaledBufferSize(SurfaceHolder holder) {
        int realWidth = getWidth();
        int realHeight = getHeight();
        if (realWidth <= 0 || realHeight <= 0) {
            return false; // Not laid out yet.
        }

        // metrics.density is the same "dp" scale factor Android uses
        // everywhere else (1.0 = 160dpi baseline; ~2.6 on a Pixel 7), and
        // unlike widthPixels/heightPixels it doesn't depend on rotation.
        DisplayMetrics metrics = getResources().getDisplayMetrics();
        float density = metrics.density > 0 ? metrics.density : 1.0f;
        int targetWidth = Math.round(realWidth / density);
        int targetHeight = Math.round(realHeight / density);

        Rect frame = holder.getSurfaceFrame();
        if (frame.width() == targetWidth && frame.height() == targetHeight) {
            return false; // Already correct.
        }

        Log.v("SDL", "ScaledSDLSurface: requesting " + targetWidth + "x" + targetHeight +
                " render buffer (density " + density + ") for a " + realWidth + "x" + realHeight + " view");
        holder.setFixedSize(targetWidth, targetHeight);
        return true;
    }

    @Override
    protected void onSizeChanged(int w, int h, int oldw, int oldh) {
        super.onSizeChanged(w, h, oldw, oldh);
        requestScaledBufferSize(getHolder());
    }

    @Override
    public void surfaceChanged(SurfaceHolder holder, int format, int width, int height) {
        if (requestScaledBufferSize(holder)) {
            // setFixedSize() above will trigger another surfaceChanged() call
            // with the corrected dimensions; bail out of this pass and let
            // that one do the work.
            return;
        }

        super.surfaceChanged(holder, format, width, height);

        // SDLSurface.onTouch() normalizes touch coordinates by dividing the
        // real on-screen touch position by mWidth/mHeight, which
        // super.surfaceChanged() just set to our *shrunk* render buffer size
        // above. MotionEvent coordinates, though, are always reported in
        // this View's actual on-screen pixels -- the full physical display,
        // since setFixedSize() only changes the buffer, not our layout size
        // -- so mWidth/mHeight need to track the real on-screen size here,
        // or every touch/click ends up scaled to the wrong spot.
        mWidth = getWidth();
        mHeight = getHeight();
    }
}
