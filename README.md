
# Mk

Mk is a reboot of the Plan 9 mk command, which itself is a replacement for make.
This tool is for anyone who loves make, but hates all its stupid bullshit.

# Why Plan 9 mk is better than make

Plan 9 mk blows make out of the water. Yet tragically, few use or have even heard
of it. Put simply, mk takes make, keeps its simple direct syntax, but fixes
basically everything that's annoyed you over the years. To name a few things:

  1. Recipes are delimited by any indentation, not tab characters in particular.
  1. Phony targets are handled separately from file targets. Your mkfile won't
     be broken by having a file named 'clean'.
  1. Attributes instead of weird special targets like `.SECONDARY:`.
  1. Special variables like `$target`, `$prereq`, and `$stem` in place of
     make's pointlessly cryptic `$@`, `$^`, and `$*`.
  1. In addition to suffix rules (e.g. `%.o: %.c`), mk has more powerful regular
     expression rules.
  1. Sane handling of rules with multiple targets.
  1. An optional attribute to delete targets when a recipe fails, so you aren't
     left with corrupt output.
  1. Plan 9 mkfiles can not only include other mkfiles, but pipe in the output of
     recipes. Your mkfile can configure itself by doing something like
     `<|sh config.sh`.
  1. A generalized mechanism to determine if a target is out of date, for when
     timestamps won't cut it.
  1. Variables are expanded in recipes only if they are defined. They way you
     usually don't have to escape `$`.

And much more! For more, read the original mk paper: ["Mk: a successor to
make"](#).

# Improvements over Plan 9 mk

This mk stays mostly faithful to Plan 9, but makes a few (in my opinion)
improvements.

  1. A clean, modern implementation in go, that doesn't depend on the whole plan
     9 for userspace stack.
  1. Use go regular expressions, which are perl-like. The original mk used plan9
     regex, which few people know or care to learn.
  1. Allow blank lines in recipes. A recipe is any indented block of text, and
     continues until a non-indented character or the end of the file.
  1. Add an 'S' attribute to execute recipes with programs other than sh. This
     way, you don't have to separate your six line python script into its own
     file. Just stick it in the mkfile.
  1. Use sh syntax for command insertion (i.e. backticks) rather than rc shell
     syntax.


# Current State

Totally non-functional. Check back later!


