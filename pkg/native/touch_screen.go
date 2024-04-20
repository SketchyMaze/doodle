package native

// Common code to handle basic touch screen detection.

/*
IsTouchScreenMode is set to true when the user has touched the screen, and false when the mouse is moved.

This informs the game's control scheme: if the user has a touch screen and they've touched it,
the custom mouse cursor is hidden (as it becomes distracting) and if they then move a real mouse
over the window, the custom cursor appears again and we assume non-touchscreen mode once more.
*/
var IsTouchScreenMode bool
