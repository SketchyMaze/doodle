# RLE Encoding for Levels

This documents some of the file size savings of the game's built-in levelpacks and doodads once the file format was migrated to use Run Length Encoding (RLE) for drawing chunks.

See [Evolution of File Formats](./Evolution%20of%20File%20Formats.md) for a history of the game's file formats.

# Levels

The file sizes of the levels themselves:

## First Quest

| Filename           | Orig Size | New Size | Reduction |
|--------------------|-----------|----------|-----------|
| Boat.level         | 4.3M      | 292K     | 93%       |
| Castle.level       | 5.6M      | 241K     | 95%       |
| Desert-1of2.level  | 4.4M      | 248K     | 94%       |
| Desert-2of2.level  | 3.2M      | 290K     | 91%       |
| Jungle.level       | 11M       | 581K     | 94%       |
| Shapeshifter.level | 22M       | 263K     | 98%       |
| Thief 1.level      | 538K      | 193K     | 64%       |

In raw bytes:

| Filename           | Orig Size | New Size |
|--------------------|-----------|----------|
| Boat.level         | 4494184   | 298943   |
| Castle.level       | 5854222   | 245872   |
| Desert-1of2.level  | 4589382   | 253768   |
| Desert-2of2.level  | 3310784   | 296681   |
| Jungle.level       | 10928779  | 594601   |
| Shapeshifter.level | 22823811  | 269307   |
| Thief 1.level      | 550579    | 196731   |

The levelpack ZIP:

* Filename: builtin-100-FirstQuest.levelpack
* Original: 50M (52369408)
* New size: 1.8M (1838542) 96%

## Tutorial

| Filename           | Orig Size | New Size | Reduction |
|--------------------|-----------|----------|-----------|
| Tutorial 1.level   | 186K      | 111K     | 40%       |
| Tutorial 2.level   | 680K      | 229K     | 66%       |
| Tutorial 3.level   | 409K      | 148K     | 64%       |
| Tutorial 4.level   | 901K      | 376K     | 58%       |
| Tutorial 5.level   | 3M        | 645K     | 78%       |
| Zoo.level          | 2.8M      | 226K     | 92%       |

In raw bytes:

| Filename           | Orig Size | New Size |
|--------------------|-----------|----------|
| Tutorial 1.level   | 190171    | 113568   |
| Tutorial 2.level   | 695936    | 233880   |
| Tutorial 3.level   | 418490    | 150565   |
| Tutorial 4.level   | 921781    | 384775   |
| Tutorial 5.level   | 3059902   | 659487   |
| Zoo.level          | 2925633   | 230712   |

The levelpack ZIP:

* Filename: builtin-Tutorial.levelpack
* Original: 7.8M (8119658)
* New size: 1.6M (1650381) 79%

## Azulian Tag

| Filename                  | Orig Size | New Size | Reduction |
|---------------------------|-----------|----------|-----------|
| AzulianTag-Forest.level   | 17M       | 312K     | 98%       |
| AzulianTag-Night.level    | 702K      | 145K     | 79%       |
| AzulianTag-Tutorial.level | 3.4M      | 185K     | 94%       |

In raw bytes:

| Filename                  | Orig Size | New Size |
|---------------------------|-----------|----------|
| AzulianTag-Forest.level   | 17662031  | 318547   |
| AzulianTag-Night.level    | 718345    | 147612   |
| AzulianTag-Tutorial.level | 3508093   | 189310   |

The levelpack ZIP:

* Filename: builtin-200-AzulianTag.levelpack
* Original: 21M (21824441)
* New size: 525K (537345) 97%

# Doodads

Spot check of random doodad filesize changes:

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

# Game Binary

* Filename: sketchymaze
* Original: 105M
* New size: 30M, 71% smaller