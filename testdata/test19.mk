all:V: result.o
%.o: %.c
	compile $stem
result.c:
	generate result.c
