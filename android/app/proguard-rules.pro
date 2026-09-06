# Nothing Go-related to shrink/obfuscate (it's a prebuilt native .so), and the
# SDL Java glue in org.libsdl.app is called into via JNI/reflection, so keep it
# intact if minifyEnabled is ever turned on.
-keep class org.libsdl.app.** { *; }
-keep class com.sketchymaze.doodle.** { *; }
