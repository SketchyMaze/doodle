# Mac OS Configuration

This directory contains a template `.app` folder for Mac OS.

## Creating the ICNS Icon

On Fedora: `dnf install libicns-utils`, Ubuntu `apt install icnsutils`

Command: `png2icns icon.icns 16.png 32.png 128.png 256.png`

Note, 48x48 is also a valid icon size but I didn't make one that size yet.

https://dentrassi.de/2014/02/25/creating-mac-os-x-icons-icns-on-linux/
