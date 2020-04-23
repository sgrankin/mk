PROG=mk

# Locations
DESTDIR=/usr/local
BINDIR=/bin

# Tools
GOTOOL=go
INSTALL=install

# Customizations: overwrite the above variables in a local config.mk file
<|cat config.mk 2>/dev/null || true

all:V:	$PROG

$PROG:
	$GOTOOL build

install:	$PROG
	$INSTALL $PROG $DESTDIR$BINDIR/$PROG

clean:V:
	rm -f $PROG
