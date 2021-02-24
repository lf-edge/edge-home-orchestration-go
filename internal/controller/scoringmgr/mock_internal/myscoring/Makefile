#
# Makefile
#

# Target parameters
LIBRARY_FILE	:= libmyscoring.so
LIBRARY_VER		:= 1.0.1
SRC_FILES		:= \
			myscoring.c
CONF_FILE		:= myscoring.conf
OBJ_FILES		:= *.o

# Build parameter

all: clean build

build:
	$(CC) -fPIC -c $(SRC_FILES)
	$(CC) -shared -o $(LIBRARY_FILE).$(LIBRARY_VER) $(OBJ_FILES) -lm
	ln -sf $(LIBRARY_FILE).$(LIBRARY_VER) $(LIBRARY_FILE)
	-rm -f $(OBJ_FILES)

clean:
	-rm -f $(OBJ_FILES)
	-rm -f $(LIBRARY_FILE).$(LIBRARY_VER)
	-rm -f $(LIBRARY_FILE)

.PHONY: build clean