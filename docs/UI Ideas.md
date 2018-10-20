# UI Toolkit Ideas

The UI toolkit was loosely inspired by Tk and could copy more of their ideas.

* **Anchor vs. Side:** currently use Anchor to mean Side when packing widgets
  into a Frame. It should be renamed to Side, and then Anchor should be how a
  widget centers itself in its space, making it easy to have a Center Middle
  widget inside a large frame.
* **Hover Background:** currently the Button sets its own color with its own
  events, but this could be moved into the BaseWidget. Tk analog is
 `activeBackground`
* **Mouse Cursor:** the BaseWidget should provide a way to configure a mouse
  cursor when hovering over the widget.

## Label

* **Text Justify:** when multiple lines of text, align them all to the
  left, center, or right.
