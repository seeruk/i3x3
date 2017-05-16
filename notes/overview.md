# Overview

The general steps that will need to be taken by i3x3 will should be the same on any system. The whole thing will be dynamic, and based off of the user's current configuration. For example, the number of screens they have active, and how big they want their grid to be.

1. Find active outputs.
    * This is done with `i3-msg -t get_outputs`.
    * Each entry has an `active` property that can be used to identify whether that screen is actually in-use with i3.
2. Fetch workspaces.
    * This is done with `i3-msg -t get_workspaces`.
    * The currently focused workspace is reported with the `focused` property.
    * Way may not need to worry about positions, as i3 can move containers between workspaces.
3. As we know how many outputs we have, the current workspace, and the size of the grid, we then have all that we need to just call regular i3 commands to send containers to workspaces as we see fit. These workspaces can be specifically numbered based on how many outputs we have (since each output will start with a single workspace, and they all lead on from the next).
    * We should be able to work out which workspace number to jump to quite easily based on the grid
    size and the currently focused workspace.

## Movement

Let's imagine we have 2 outputs, and a grid that's 3x3 (so unrealistic). Output 1 will start with workspace 1, and output 2 will start with workspace 2.

Output 1 Grid Layout

1  3  5
7  9  11
13 15 17

Output 2 Grid Layout

2  4  6
8  10 12
14 16 18

So, we have a workspace range of 1 through 18 inclusive. Since those numbers only show up on a specific workspace, that means based on the number given we should be able to figure out the workspace we're on.

There's probably going to be a bit of maths involved in figuring out which part of the grid to move things to, but it shouldn't be too difficult...

`b` = Output number
`c` = Current workspace
`o` = Number of outputs
`x` = Grid x size
`y` = Grid y size

Moving up: c - (o * x)
Moving up: 16 - (2 * 3) = 10

Moving down: c + (o * x)
Moving up: 2 + (2 * 3) = 8

Moving left: c - o
Moving left: 10 - 2 = 8

Moving right: c + o
Moving right: 8 + 2 = 10

So, moving around is pretty simple, but how do we know when we're at an edge (i.e. the point where we should just do nothing).

Left edge: ((c - b) / y) % 2
Left edge: ((2 - 2) / 3) % 2 = 0
Left edge: ((8 - 2) / 3) % 2 = 0
Left edge: ((14 - 2) / 3) % 2 = 0
Left edge: ((7 - 1) / 3) % 2 = 0
Left edge: ((61 - 1) / 6) % 2 = 0

Not left edge: ((9 - 1) / 3) % 2 = 0.666666...
Not left edge: ((70 - 1) / 6) % 2 = 1.5

Right edge: ((c - (b * x)) / y) % 2
Right edge: ((6 - (2 * 3)) / 3) % 2 = 0
Right edge: ((70 - (1 * 4)) / 6) % 2

Not right edge: ((4 - (2 * 3)) / 3) % 2 = 1.333333...
Not right edge: ((67 - (1 * 4)) / 6) % 2

Top edge: c - (o * x) <= 0
Top edge: 6 - (2 * 3) <= 0 = true
Top edge: 4 - (2 * 3) <= 0 = true

Not top edge: 8 - (2 * 3) <= 0 = false

Bottom edge: c + (o * x) > (o * ((x * y) - 1)) + b
Bottom edge: 14 + (2 * 3) > (2 * ((3 * 3) - 1)) + 2 = true
Bottom edge: 15 + (2 * 3) > (2 * ((3 * 3) - 1)) + 1 = true

Not bottom edge: 12 + (2 * 3) > (2 * ((3 * 3) - 1)) + 2 = false
Not bottom edge: 11 + (2 * 3) > (2 * ((3 * 3) - 1)) + 1 = false

CORRECT:
Right edge: (((c - (o * (x - 1))) - b) / y) % 2 == 0
