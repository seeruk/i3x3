# i3x3

Go-based i3 grid workspace manager, using i3-msg.


## Installation

```
$ go get -u -v github.com/SeerUK/i3x3
```

## Configuration

Configuration for i3x3 is meant to be pretty simple, here's an example of some configuration you can
drop right into your i3 config file:

```
# switch to adjacent workspace
bindsym $mod+Control+Left exec i3x3 -direction left
bindsym $mod+Control+Right exec i3x3 -direction right
bindsym $mod+Control+Up exec i3x3 -direction up
bindsym $mod+Control+Down exec i3x3 -direction down

# move focused container to adjacent workspace
bindsym $mod+Control+Mod1+Left exec i3x3 -direction left -move
bindsym $mod+Control+Mod1+Right exec i3x3 -direction right -move
bindsym $mod+Control+Mod1+Up exec i3x3 -direction up -move
bindsym $mod+Control+Mod1+Down exec i3x3 -direction down -move
```

This will allow you to use a 3x3 grid that is separate on each output currently active in i3, using 
the arrow keys to switch between, or move containers across workspaces.

## Workspace Arrangement

If you have a single output (i.e. single monitor) then you probably won't notice anything fancy 
going on at all. Your workspaces will go up in increments of 1, like you'd expect.

If you have multiple outputs however, things work slightly differently. Taking the default 3x3 grid,
this is what is how your workspaces will be distributed (bear in mind you may not have all 
workspaces active).

```
 Output 1  |  Output 2
---------- | ----------
 1  3  5   |  2  4  6
 7  9  11  |  8  10 12
 13 15 17  |  14 16 18
```

Each output will have it's own unique set of workspaces. If you only use i3x3 to navigate your 
workspaces (which is recommended, more on that in a moment) then they will never leave that output.

Each output in i3 starts with a workspace. So, if you have 3 outputs, you'll have 3 workspaces (1, 
2, and 3). From this point, i3x3 uses that value to identify the other values that are available, 
along with the number of outputs active in i3. The gap between workspace numbers on each output is
equal to the number of outputs you have. If you have 3 monitors, this is how it would be arranged:


```
 Output 1  |  Output 2  |  Output 3
---------- | ---------- | ----------
 1  4  7   |  2  5  8   |  3  6  9
 10 13 16  |  11 14 17  |  12 15 18
 19 22 25  |  20 23 26  |  21 24 27
```

This pattern allows i3x3 to more reliably identify the display you're currently on, while knowing as
little as possible about your physical setup.

## Todo

* Some sort of visualisation when switching?
* Refactoring... the code stinks right now, it was a single-evening endeavour.
* Configuration of grid size (by environment variables?)
* Tests (definitely for the mathy bits!)

## License

MIT
