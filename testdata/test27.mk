%.o: %.c
	generic $stem
foo.o: foo.c
	specific foo
foo.c:
	gen foo.c
