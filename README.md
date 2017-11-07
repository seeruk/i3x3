# i3x3

Go-based i3 grid workspace manager, using i3-msg.

## Installation

Install the dependencies for [mattn/go-gtk][1] (i.e. GTK development libraries and tools). Then you
can simply run the following:

```
$ go get -u -v github.com/SeerUK/i3x3/...
```

This will install 3 binaries, `i3x3ctl`, `i3x3fixd`, and `i3x3overlayd`.

## Configuration

Configuration for i3x3 is meant to be pretty simple, here's an example of some configuration you can
drop right into your i3 config file:

```
# switch to adjacent workspace
bindsym $mod+Control+Left exec i3x3ctl -direction left
bindsym $mod+Control+Right exec i3x3ctl -direction right
bindsym $mod+Control+Up exec i3x3ctl -direction up
bindsym $mod+Control+Down exec i3x3ctl -direction down

# move focused container to adjacent workspace
bindsym $mod+Control+Mod1+Left exec i3x3ctl -direction left -move
bindsym $mod+Control+Mod1+Right exec i3x3ctl -direction right -move
bindsym $mod+Control+Mod1+Up exec i3x3ctl -direction up -move
bindsym $mod+Control+Mod1+Down exec i3x3ctl -direction down -move
```

This will allow you to use a 3x3 grid that is separate on each output currently active in i3, using 
the arrow keys to switch between, or move containers across workspaces. You will also have a hotkey
for "fixing" the workspace layout, this is useful for if you add or remove displays, because of the
way that i3x3 requires workspaces to be arranged (see below).

### Daemons

For some other functionality like the overlay, and automatic workspace redistribution, you'll need 
to launch both `i3x3fixd` and `i3x3overlayd`. You can just throw those in your `.xinitrc`, etc.

```
# i3x3fixd; for automatically redistributing workspaces
i3x3fixd > /tmp/i3x3fixd.log 2>&1 &

# i3x3overlayd; for i3x3's overlay
i3x3overlayd > /tmp/i3x3overlayd.log 2>&1 &
```  

### Grid Size

The grid size can be configured by using the environment variables `I3X3_X_SIZE` and `I3X3_Y_SIZE`.
They must be set to numeric values, and should be integers. The defaults should be obvious.

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
2, and 3). i3x3 works out which output a workspace belongs to based on the number of workspaces 
and the number of outputs. The gap between workspace numbers on each output is equal to the number 
of outputs you have. Workspaces starts in the top-left corner with the output number. If you have 3
monitors, this is how it would be arranged:


```
 Output 1  |  Output 2  |  Output 3
---------- | ---------- | ----------
 1  4  7   |  2  5  8   |  3  6  9
 10 13 16  |  11 14 17  |  12 15 18
 19 22 25  |  20 23 26  |  21 24 27
```

### But why?

You might be wondering why the workspaces aren't just arranged so that they go up 1 at a time, left
to right in rows:

```
 Output 1  |  Output 2 
---------- | ----------
 1  2  3   |  10 11 12 
 4  5  6   |  13 14 15
 7  8  9   |  16 17 18
```

Firstly, the pattern that i3x3 uses allows i3x3 to more reliably identify the display you're 
currently on, while knowing as little as possible about your physical setup. This enables some other
useful functionality.

Also, i3 starts with a workspace on each output, and the number matches the output the workspace
is on. This means when you log in, i3x3 is ready to work without having to change anything about the
workspaces that i3 gives us to start with. With the example given above you would have to "fix" the
workspaces as soon as i3 starts for i3x3 to work.

Additionally, if i3x3 did use the layout shown above, instead of the layout it actually uses, it 
would be impossible to dynamically adjust the size of the grid like it can now. For example, with
the way i3x3 works right now, if you have 2 outputs, 1 display will have all odd numbered workspaces
on, and the other will have all even numbered workspaces on. Given this, even if you have a 3x3 
grid (meaning the maximum workspace you could reach by using i3x3 alone would be 18), i3x3 can still
figure out which display any workspace belongs to (e.g. workspace 32 would belong on output 2). This
also enables i3x3-fix to work; where would the workspaces for a third display go if you reduced the 
number of outputs to 2? How would i3x3 automatically increase the grid size to accomodate more 
workspaces? Things would get very weird, very fast.

## Todo

* Tests (definitely for the mathy bits!)

## License

MIT

[1]: https://github.com/mattn/go-gtk
