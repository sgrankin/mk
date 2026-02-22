(\w+)\.o:R: ${stem1}.c
	compile ${stem1}
all:V: foo.o
foo.c:
	gen foo.c
