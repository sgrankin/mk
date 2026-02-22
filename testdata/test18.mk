# TEST_MAIN is set by the test harness.
# Test that multiline output is correctly expanded to multiple targets

deps = `printf 'hello\nworld'`

all:V: $deps

$deps:V:
	process $target
