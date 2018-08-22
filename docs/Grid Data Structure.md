# Grid Data Structure Ideas

Ideas for managing level pixel grids.

## Chunks

Divide the grid into Chunks of an arbitrary size. The canvas metadata would
specify the chunk size (e.g. 1000x1000 pixels) so it's able to vary by the
drawing itself.

A `ChunkManager` structure would wrap around the chunks and give each one a
coordinate value. To convert an absolute pixel value into a chunk coordinate
you multiply it by the chunk size.

So a pixel coord of `123456` in a map with 1000x1000 chunk sizes would be
`(123456) / 1000 = 123.456 = 123` (rounded down). This would give you a chunk
index and then that chunk is responsible for knowing where the exact pixel is.

## Chunk Types

Each chunk could organize its pixels into two types of structures depending
on the density of the chunk:

1. A hash map for sparse chunks.
2. A 2D array of fixed size for dense chunks.

So if a chunk is <70% filled it would use a hash map and when it gets heavier
than that, switch to a 2D array.

Something like,

```go
type Chunk struct {
    Type int `json:"type"` // map vs. 2D array
    Map map[Point]int      // map of (X,Y) points to Palette entry
    Grid [][]int           // 2D array of Palette entries
}
```

## External Chunks

For normal single-player maps that aren't infinite in size, all the chunks
are stored in the singular level file. Larger maps can start saving their
chunks to disk in external files.

Some internal threshhold can be used to decide when to start saving chunks
as external files. If the chunk size is 1000x1000 pixels it could start saving
external files after (say) 9 different chunks are allocated for the first time,
or some time/space tradeoff like that.
