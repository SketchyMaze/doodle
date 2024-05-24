# Evolution of File Formats

This document will cover the evolution of the game's primary file formats (Level and Doodad drawings): how the on-disk format has changed over time to better compress/optimize the drawing data, and how the game continued to maintain backwards compatibility.

The game, so far, is always able to _read_ levels and doodads created by older versions (all the way back to the very first alpha build!) and, upon saving them, will convert the file format to the latest standard in order to optimize and reduce disk space usage.

The game can generally be configured (by editing feature flag constants) to _output_ drawings in the various legacy formats as well. Between the v1 JSON and v2 Gzip-JSON formats, the game is able to translate back and forth. From v3 Zipfiles onwards, back-migrating drawing files is not a supported operation - it can always save drawings _forward_ but code is not in place to e.g. take Zipfile members and put them back into the root struct to revert to a classic JSON-style file.

Table of Contents:

- [Evolution of File Formats](#evolution-of-file-formats)
- [General Design](#general-design)
  - [A common file format between Levels and Doodads](#a-common-file-format-between-levels-and-doodads)
  - [Chunks and the Chunker](#chunks-and-the-chunker)
  - [Chunk Accessors](#chunk-accessors)
- [File Format Versions](#file-format-versions)
  - [v1: Simple JSON Format](#v1-simple-json-format)
  - [v2: GZip compressed JSON](#v2-gzip-compressed-json)
  - [v3: Zip archives with external chunks](#v3-zip-archives-with-external-chunks)
    - [Structure of the Zipfile](#structure-of-the-zipfile)
    - [How I made it backwards compatible](#how-i-made-it-backwards-compatible)
    - [Migrating the ZIP file on every Save](#migrating-the-zip-file-on-every-save)
    - [The old Chunks and Files become staging areas](#the-old-chunks-and-files-become-staging-areas)
  - [v3.1: A binary chunk file format](#v31-a-binary-chunk-file-format)
  - [v4: Run Length Encoding MapAccessor](#v4-run-length-encoding-mapaccessor)
    - [File Size Savings](#file-size-savings)
      - [First Quest](#first-quest)
      - [Tutorial Levels](#tutorial-levels)
      - [Azulian Tag](#azulian-tag)
      - [Built-in Doodads](#built-in-doodads)
      - [Game Binary Size](#game-binary-size)

# General Design

Some thought and planning went into this in the very beginning. This section covers the general design goals of Levels & Doodads, how their actual pixel data is managed (in chunks), and how I left it open-ended to experiment with different chunk accessor algorithms in the future.

## A common file format between Levels and Doodads

Under the hood, the file format for .level and .doodad files are extremely similar. They share a handful of properties in common at their root data structure:

* Version (of the file format - still at version 1!)
* Game Version (that last saved this file)
* Title, Author, common metadata
* Attached file storage such as custom wallpapers

Both levels and doodads also have a Chunker somewhere that stores the actual pixels-on-screen drawing data for them.

Both file types have evolved together, and when optimizations are made for e.g. Level files, the Doodads automatically benefit too for sharing a lot of code in common.

## Chunks and the Chunker

The drawing data itself (pixels on screen) from the beginning was decided to be split into Chunks, and each Chunk would manage the pixels for its part of the overall drawing. This can enable arbitrarily large drawing sizes for levels and doodads, with theoretically "infinite" boundaries (within computer integer bounds).

Each Level or Doodad file will have one or more Chunkers. The Chunker itself stores the common properties for the drawing (like the Chunk Size, e.g., 128 square pixels by default), and it manages translating from "world coordinates" of your drawing into "chunk coordinates", so it knows which Chunk is responsible for that part of the drawing.

For an arbitrary world coordinate (like 900,-290) the Chunker can divide it by the Chunk Size of 128 and find that chunk coordinate (7,-2) is responsible for that chunk and asks it for its pixels.

Levels currently only have one Chunker, but Doodads have many (one for each frame of animation they store).

## Chunk Accessors

There is support from the beginning for each Chunk to manage its own data in any way it wants to. For example, a Chunk that is completely filled by one color of pixel could store its information _much_ more succinctly than a chunk made up of very sparse lines, where each pixel coordinate needs to be accounted for.

The first type of Chunk accessor was the **MapAccessor**, which stored the X,Y coordinates of each pixel mapped to their Palette color index (see the example below, in [v1: Simple JSON Format](#v1-simple-json-format)).

It was planned that future accessors would be added such as a **GridAccessor** for very densely packed chunks (to store in a 2D array) and have the Chunker automatically decide which format is optimal to encode it but this was still never added.

From the game's first alpha (0.0.9, July 9 2019) through version 0.14.0 (May 4 2024), the MapAccessor was the only one ever implemented.

# File Format Versions

## v1: Simple JSON Format

At first, levels were just saved as simple JSON files (whitespace compressed only), which when pretty printed (and with comment annotations added) looked like this:

```javascript
{
    // Common properties between levels and doodads
    "version": 1,     // json schema version, still at "1" today!
    "gameVersion": "0.0.10-alpha",
    "title": "Alpha",
    "author": "Noah P",
    "locked": false,  // read locked/won't open in editor
    "files": null,    // attached files
    "passwd": "",     // level password (never used)

    // The drawing data itself, divided into chunks.
    "chunks": {
        "size": 128,
        "chunks": {
            // Chunk coordinate
            "0,0": {
                "type": 0, // 0 = MapAccessor chunk type
                "data": {
                    // Each pixel coordinate mapped
                    // to a palette index number...
                    "69,32": 0,
                    "69,33": 0,
                    "70,34": 0,
                }
            }
        }
    },
    "palette": {
        "swatches": [
            // indexed color palette for the drawing
            {
                "name": "solid",
                "color": "#000000",
                "solid": true
            },
            {
                "name": "decoration",
                "color": "#999999"
            },
            {
                "name": "fire",
                "color": "#ff0000",
                "fire": true
            },
            {
                "name": "water",
                "color": "#0000ff",
                "water": true
            }
        ]
    },
    "pageType": 2,  // 2 = Bounded LevelType
    "boundedWidth": 2550,
    "boundedHeight": 3300,
    "wallpaper": "notebook.png",
    "actors": {
        // doodads in your level, by their instanced ID
        "4d193308-a52d-4153-a10d-a010445dd47b": {
            "filename": "button.doodad",
            "point": "154,74",
            "links": [
                // linked actor IDs
                "8d501581-0904-4dfb-a326-57330b2484be"
            ]
        },
        "8d501581-0904-4dfb-a326-57330b2484be": {
            "filename": "electric-door.doodad",
            "point": "320,74",
            "links": [
                "4d193308-a52d-4153-a10d-a010445dd47b"
            ]
        }
    }
}
```

A **doodad file** was very similar but had some other relevant properties in its JSON format, such as:

* Their Size (dimensions)
* JavaScript source code
* Hitbox, Tags/Options

A doodad file has one Palette like a level, but it has multiple chunkers (one for each layer of the doodad; which are how you store frames for animation or state changes).

What level and doodads have in common is File storage for attaching files into them (such as custom wallpapers for a level, or sound effects for a doodad), with their binary data encoded to base64.

In the first iteration of the file format, _all_ of this was encoded into the single JSON file on disk!

For densely packed levels, though, the JSON file got really large quickly, even with the whitespace removed.

## v2: GZip compressed JSON

The second iteration was to basically add gzip compression to the level files, which slashed their file size considerably.

How I made it backwards compatible:

The game is able to open a ".level" file which is _either_ a straight JSON file from older versions of the game, or the new gzip compressed format.

When **opening** a file, it:

1. Checks if the first byte is the ASCII character `{`, and will parse it as legacy v1 JSON format.
2. Checks if the file's opening bytes are instead a gzip header (hex `1f8b`), and will load it from GZip (v2 file format).

The GZip reader is basically a wrapper that decodes a JSON file with compression:

```go
// pkg/level/fmt_json.go

// FromGzip deserializes a gzip compressed level JSON.
func FromGzip(data []byte) (*Level, error) {
	// This function works, do not touch.
	var (
		level   = New()
		buf     = bytes.NewBuffer(data)
		reader  *gzip.Reader
		decoder *json.Decoder
	)

	reader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	decoder = json.NewDecoder(reader)
	decoder.Decode(level)

	return level, nil
}
```

## v3: Zip archives with external chunks

Very large or dense levels were resulting in enormous file sizes even with gzip compression, and they were taking a long time to load from disk since _all_ chunk data was still in one file!

A loading screen feature was added to Sketchy Maze around this time because big levels could take _seconds_ to load.

The level format was reworked again and now the .level file is basically a Zip archive with member files within. Most importantly, this enabled the drawing's chunks to be kicked out into separate files so we could manage "loading" the level more efficiently.

### Structure of the Zipfile

A level zipfile is laid out like so:

```
/
  level.json
  assets/
    screenshots/
      large.png
  chunks/
    0/
      -1,0.json
      -1,1.json
      -2,0.json
```

The `level.json` file contains most of the basic metadata from the old file format, except the chunks are evicted and stored in separate JSON files by their chunk coordinate. Doodads will have a `doodad.json` file here instead.

Notice the directory name **0/** holding the chunks for a level file: this zero is a layer ID to accommodate Doodads which share a similar file format. Levels only have one Layer (for now), so the directory name is always zero. Doodads will have each of their layers enumerated from 0, 1, 2, ...

Attached assets such as wallpapers or embedded custom doodads would be regular ZIP file members under the assets/ folder.

The level.json file as of v0.14.0 looks a bit like this:

```json
{
    "version": 1,
    "gameVersion": "0.14.0",
    "title": "The Castle",
    "author": "Noah P",
    "locked": false,
    "files": {},  // files are evicted to the assets/ folder
    "passwd": "",
    "uuid": "18f0f734-d7ad-4b10-be6d-f40d31334816",
    "rules": {
        "difficulty": 0
    },
    "chunks": {
        "size": 128,
        "chunks": {} // chunks are evicted as well
    },
    "palette": {
        "swatches": [
            {
                "name": "grass",
                "color": "#009900",
                "pattern": "noise.png",
                "solid": true
            },
            {
                "name": "fire",
                "color": "#ff0000",
                "pattern": "marker.png",
                "fire": true
            }
        ]
    },
    "pageType": 0,
    "boundedWidth": 2550,
    "boundedHeight": 3300,
    "wallpaper": "graph.png",
    "scroll": "-3072,51",
    "actors": {
        "0fac06dc-fe1e-11eb-9dfc-9cb6d0c2aa8b": {
            "filename": "key-blue.doodad",
            "point": "4403,239"
        },
        "10554b0f-fe18-11eb-9dfc-9cb6d0c2aa8b": {
            "filename": "door-blue.doodad",
            "point": "3909,344"
        },
    },
    "saveDoodads": false,
    "saveBuiltins": false
}
```

### How I made it backwards compatible

The function that reads Levels and Doodads from disk continues to check its headers:

1. If the first byte is a `{` it's a legacy old drawing and is parsed as classic JSON (format v1)
2. If the first bytes are a GZip header, try loading it as a gz compressed JSON file (format v2)
3. If instead the header is a ZIP archive, open it and look for the `level.json` or `doodad.json` for the expected file you're opening.

It was able to continue loading drawings from the _very_ earliest alpha of the game and can still load them today.

### Migrating the ZIP file on every Save

The upgrade path to save a legacy drawing in the modern ZIP format was very straightforward:

In the root Level struct (what level.json decodes into):

* Deprecate the "files" and "chunks" fields since these are evicted out into separate ZIP file members, so they should be empty in the new file format.
* So on level save, if the Level has any data in these fields (meaning you had loaded it from a legacy file format), evict those fields to clear them out when generating the new ZIP file.

When an opened level _was_ a ZIP file to begin with, a pointer to the Zipfile handle is kept accessible to make saving those levels more efficient. When a level _wasn't_ a ZIP file, I basically create a new one, write the level.json and flush out its embedded files/chunks to their correct places.

So, when you've opened a ZIP file level and you re-save it, the process is:

1. First, copy any "cold storage" chunks from the old Zipfile to the new one.
    * These are the chunks not actively in memory (see the next section about [staging areas](#the-old-chunks-and-files-become-staging-areas))
2. Then, flush out any data in the legacy "chunks" or "files" sections to external zipfile members.
    * This is the same logic for migrating an old gzip-json level, where _all_ its data was in these places...
    * As well as flushing out recently edited chunks or recently attached files (per the next section).

### The old Chunks and Files become staging areas

In the base Level struct, the old keys where Files and Chunks used to be stored have now become the staging area for "warmed up" chunks or recently attached files.

For example: the level.json file in the ZIP stores no data in these fields, and chunks are stored as separate members. Whenever the game **loads** a chunk from ZIP, it will cache it in the old Chunks structure so it has it warmed up and ready to use.

When **playing** a level: there is a chunk loading/unloading algorithm that balances memory use during gameplay. Chunks which are currently on screen may be fetched from the ZIP file and cached in the legacy Chunks structure. The game will track which chunks are accessed on the current game tick (as well as the previous couple of ticks).

If a chunk has not been accessed in a few game ticks, it is destroyed and removed from the legacy Chunks structure (along with its SDL2 texture being cleaned up, etc.); if the player scrolls the chunk back on screen, it is recalled from the ZIP file and cached again.

When **editing** a level in the editor, any chunk that receives a local modification is also stored in the old Chunks structure, and is kept there until the next save: when all the loaded chunks are flushed out to ZIP files. Chunks with modifications are NOT flushed by the auto-loading/unloading algorithm so their changes don't get lost.

## v3.1: A binary chunk file format

At this point: there is still only one Chunk accessor (the MapAccessor) and its JSON files in the zip file still looked like (if pretty printed):

```javascript
{
    "type": 0,
    "data": {
        "69,32": 1, // coordinate to palette index
        "69,33": 2,
        "70,34": 0,
    }
}
```

The next iteration was to compress these down into a binary format to shrink them further by removing the extra JSON characters (quotes, brackets, etc.) and the ASCII human readable digits.

In the ZIP file: the legacy chunks will have their .json file extension but the new binary format stores them into .bin files; so the game is able to load old and new levels by checking the file types available for their chunks.

The binary format makes use of variable-length integers provided by Go's encoding/binary package. This is the same VarInt type from Protocol Buffers: small numbers encode to a few number of bytes, and large numbers may use additional bytes.

* The **first** Uvarint in the binary format is the chunk type (0 = MapAccessor)
* The remaining data is arbitrary and up to that chunk accessor to handle how it wants.

For the MapAccessor: the remaining binary data is a repeating stream of three varints:

1. X coordinate
2. Y coordinate
3. Palette index number

For migrating old JSON chunks into binary format: on save it will always output in the .bin format (by calling the chunk accessor's MarshalBinary method), but on reading is able to handle both .bin and legacy .json.

## v4: Run Length Encoding MapAccessor

After the release of v0.14.0 of the game, a new chunk accessor has _finally_ been added to the game: the **RLEAccessor**.

The RLEAccessor is functionally identical to the MapAccessor, in that (in working memory) it stores a hash map of world coordinates to the palette color. But where the RLEAccessor is different is with the **on disk format** of how it encodes its chunks.

The on-disk format uses binary (.bin) only, and compressed the chunk's pixel data using Run Length Encoding (RLE). The algorithm is basically:

* When **compressing** your chunk data to save on disk:
    * It creates a 2D grid array of integers in order to rasterize a complete bitmap of the chunk.
        * For a chunk size of 128, this is a 128x128 2D array.
        * The values are your palette index numbers (0 to N)
        * "Null" colors that are blank in the chunk uses the value 0xFFFF.
        * Note: the gameplay logic enforces only 256 colors per level palette, but theoretically 65,534 colors could be supported before the "null" color would collide.
    * It then serializes the 2D bitmap using RLE with a series of packed Uvarints:
        1. The palette color to set
        2. The number of pixels to repeat that palette color.
* When **decompressing** the RLE encoded data, the process is reversed:
    * It creates a 2D grid of your square chunk size again (all nulls)
    * Then it decompresses the RLE encoded stream of Uvarints, filling out the grid from the top-left to bottom-right corner.
    * Finally, it scans the grid to find non-null colors to populate its regular MapAccessor struct of points-to-colors.

For a simple example: if a chunk consisted 100% of the same color on all 128x128 pixels, the compressed RLE stream contains only 3 or 4 bytes on disk:

1. The palette index number
2. The repeat number (16,384 for a 128x128 chunk grid)

For **migrating MapAccessors to RLEAccessors:**

The game is still able to read legacy MapAccessor chunks, and when **saving** a drawing back to disk, it fans out and checks all your level chunks if they need to be optimized:

* If their chunk type is a MapAccessor, copy the underlying map data into an RLEAccessor.
* Then when saving to disk, the RLEAccessor MarshalBinary() func will create the .bin file in the updated format on disk.

### File Size Savings

On average the RLE encoding slashes file sizes by over 90% for most levels, especially densely packed levels with lots of large colored areas.

Here are examples from the game's built-in level packs.

See [RLE Encoding for Levels](./RLE%20Encoding%20for%20Levels.md) for more breakdown of these numbers.

#### First Quest

| Filename           | Orig Size | New Size | Reduction |
|--------------------|-----------|----------|-----------|
| Boat.level         | 4.3M      | 292K     | 93%       |
| Castle.level       | 5.6M      | 241K     | 95%       |
| Desert-1of2.level  | 4.4M      | 248K     | 94%       |
| Desert-2of2.level  | 3.2M      | 290K     | 91%       |
| Jungle.level       | 11M       | 581K     | 94%       |
| Shapeshifter.level | 22M       | 263K     | 98%       |
| Thief 1.level      | 538K      | 193K     | 64%       |

The combined levelpack ZIP file itself:

* Filename: builtin-100-FirstQuest.levelpack
* Original: 50M (52369408)
* New size: 1.8M (1838542) 96%

The most notable improvement is Shapeshifter.level, which features **large** chunks of solid color and it compressed by 98% with the RLE encoding!

#### Tutorial Levels

Many of the Tutorial levels are made of sparsely drawn "line art" rather than solid colored areas, so the reduction in filesize is closer to ~60% instead of 90%+

| Filename           | Orig Size | New Size | Reduction |
|--------------------|-----------|----------|-----------|
| Tutorial 1.level   | 186K      | 111K     | 40%       |
| Tutorial 2.level   | 680K      | 229K     | 66%       |
| Tutorial 3.level   | 409K      | 148K     | 64%       |
| Tutorial 4.level   | 901K      | 376K     | 58%       |
| Tutorial 5.level   | 3M        | 645K     | 78%       |
| Zoo.level          | 2.8M      | 226K     | 92%       |

The levelpack ZIP:

* Filename: builtin-Tutorial.levelpack
* Original: 7.8M (8119658)
* New size: 1.6M (1650381) 79%

#### Azulian Tag

| Filename                  | Orig Size | New Size | Reduction |
|---------------------------|-----------|----------|-----------|
| AzulianTag-Forest.level   | 17M       | 312K     | 98%       |
| AzulianTag-Night.level    | 702K      | 145K     | 79%       |
| AzulianTag-Tutorial.level | 3.4M      | 185K     | 94%       |

The levelpack ZIP:

* Filename: builtin-200-AzulianTag.levelpack
* Original: 21M (21824441)
* New size: 525K (537345) 97%

#### Built-in Doodads

The RLE compression also improved the file sizes of the game's built-in doodads. For a random spot check of some:

| Filename                 | Orig Size | New Size |
|--------------------------|-----------|----------|
| anvil.doodad             | 2.7K      | 1.3K     |
| azu-blu.doodad           | 8.1K      | 5.2K     |
| azu-red.doodad           | 8.1K      | 5.2K     |
| azu-white.doodad         | 8.1K      | 5.2K     |
| box.doodad               | 29K       | 4.1K     |
| boy.doodad               | 30K       | 8.1K     |
| crumbly-floor.doodad     | 15K       | 3.3K     |
| door-blue.doodad         | 18K       | 2.7K     |
| electric-trapdoor.doodad | 9.5K      | 2.8K     |

Total file size of all builtin doodads:

* Original: 576.8 KiB
* New: 153.7 KiB (73% reduction)

#### Game Binary Size

The game binary embeds its built-in doodads and levelpacks directly, and so this optimization has also slashed the overall size of the game binary too:

* Filename: sketchymaze
* Original: 105M
* New size: 30M, 71% smaller