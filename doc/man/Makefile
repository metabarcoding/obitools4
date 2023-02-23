MANPAGES= obigrep 

BUILDDIR=../build
MANDIR=$(BUILDDIR)/_man
MANDEST=$(MANDIR)/man1
HTMLDEST=$(MANDIR)/html

MANSRC=$(MANPAGES:=.qmd)
DEPS=$(patsubst %,depends/%,$(MANPAGES:=.d))
MAN=$(patsubst %,$(MANDEST)/%,$(MANSRC:.qmd=.man))



all: $(MAN) 

clean:
	rm -f $(MAN) 
	rm -rf depends

.PHONY: all

$(MANDEST):
	@echo Creating $@ directory
	@mkdir -p $@

$(MAN) : $(MANDEST)/%.man : %.qmd $(MANDEST)
	@echo "Rendering the man page for " $(notdir $(@:.man=))
	@quarto render  $< --to man 
	@mv $(notdir $@) $@
	@echo =====================================================
	@echo

depends/%.d: %.qmd
	@mkdir -p depends
	@echo Generating depends file for $(notdir $(@:.qmd=))
	@awk -v src=$< 'BEGIN {printf("%s: ",src)} \
	                /\{\{< *include *[^>]+>\}\}/ {sub(/^ *\{\{< *include */,"",$$0); \
										sub(/ *> *\}\} */,"",$$0); \
										printf("%s ",$$0)}' $< > $@

-include $(DEPS)										