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

## Todo

* Some sort of visualisation when switching?
* Refactoring... the code stinks right now, it was a single-evening endeavour.
* Configuration of grid size (by environment variables?)
* Tests (definitely for the mathy bits!)

## License

MIT
