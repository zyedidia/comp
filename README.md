# Comp

This tool looks for a file called `compile_commands.json` up the directory
hierarchy and uses it to attempt to build some input file. For example, `comp
file.c` will first search for `compile_commands.json`, and if found will
execute the command that is specified for building `file.c`. This can be used
to implement a simple editor-integrated linter that is aware of the commands
used to actually build the files in your project.

I may make some editor plugins for this in the future.
