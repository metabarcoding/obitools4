
SOURCES = apat_parse.c \
         apat_search.c \
         libstki.c
         
SRCS=$(SOURCES)

         
OBJECTS= $(patsubst %.c,%.o,$(SOURCES))

LIBFILE= libapat.a
RANLIB=ranlib


include ../global.mk

all: $(LIBFILE)

clean:
	rm -rf $(OBJECTS) $(LIBFILE)

$(LIBFILE): $(OBJECTS)
	ar -cr $@ $?
	$(RANLIB) $@
