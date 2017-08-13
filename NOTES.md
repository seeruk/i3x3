# Notes

* `i3x3-fix` could become `i3x3-fixd` quite easily. We can just call `i3-msg -t get_outputs` (and we 
already have this in a function too) in a loop with a sleep for a few seconds in it, storing the
number of outputs that are active in memory. If the value has changed since the last loop, then we 
can automatically fix the workspaces.
    * This is also quite nice, as it wouldn't take up a key-binding, won't take up too much memory, 
    won't take up too much CPU time, should be very efficient, will be very accurate to i3, and will
    react quickly _enough_ for it to be effective.
    * How do we sort the displays?
        * In the order they're output in?
        * Alpha-numeric sort on display name?
        * Primary always first?
        * Is there a way at all that will ensure the same order as when you log in?
            * Probably not.
            * All that should matter is that it is consistent.
        * Given the above, maybe always have primary first, then order by name alphabetically?
            * This isn't a huge deal, primary should definitely be first, but after that, any other
            sorting mechanism only needs to be consistent. That's all.
