#!/usr/bin/env python3

import codecs
import glob
import os
import markdown
import jinja2
import re

def main():
    if not os.path.isdir("./compiled"):
        print("Make output directory: ./compiled")
        os.mkdir("./compiled")
        os.mkdir("./compiled/pages")

    # Render the main index.html template.
    with codecs.open("./pages/index.html", "r", "utf-8") as fh:
        html = fh.read()
        html = render_template(html)
        with open("./compiled/index.html", "w") as outfh:
            outfh.write(html)

    # Load the Markdown wrapper HTML template.
    html_wrapper = "$CONTENT"
    with codecs.open("./pages/markdown.html", "r", "utf-8") as fh:
        html_wrapper = fh.read()

    for md in glob.glob("./pages/*.md"):
        filename = md.split(os.path.sep)[-1]
        htmlname = filename.replace(".md", ".html")
        print("Compile Markdown: {} -> {}".format(filename, htmlname))

        with codecs.open(md, 'r', 'utf-8') as fh:
            data = fh.read()
            rendered = markdown.markdown(data,
                extensions=["codehilite", "fenced_code"],
            )
            html = html_wrapper.replace("$CONTENT", rendered)
            html = render_template(html,
                title=title_from_markdown(data),
            )

            with open(os.path.join("compiled", "pages", htmlname), "w") as outfh:
                outfh.write(html)

jinja_env = jinja2.Environment()

def render_template(input, *args, **kwargs):
    templ = jinja_env.from_string(input)
    return templ.render(
        app_name="Project: Doodle",
        app_version=get_app_version(),
        *args, **kwargs
    )

def title_from_markdown(text):
    """Retrieve the title from the first Markdown header."""
    for line in text.split("\n"):
        if line.startswith("# "):
            return line[2:]

def get_app_version():
    """Get the app version from pkg/branding/branding.go in Doodle"""
    ver = re.compile(r'Version\s*=\s*"(.+?)"')
    with codecs.open("../../pkg/branding/branding.go", "r", "utf-8") as fh:
        text = fh.read()
        for line in text.split("\n"):
            m = ver.search(line)
            if m:
                return m[1]

if __name__ == "__main__":
    main()
