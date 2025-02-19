#export GOPATH=$(shell pwd)/GO
#export GOBIN=$(GOPATH)/bin
#export PATH=$(GOBIN):$(shell echo $${PATH})

GOCMD=go
GOBUILD=$(GOCMD) build # -compiler gccgo -gccgoflags -O3
GOGENERATE=$(GOCMD) generate
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BUILD_DIR=build
OBITOOLS_PREFIX:=

PACKAGES_SRC:= $(wildcard pkg/*/*.go pkg/*/*/*.go)
PACKAGE_DIRS:=$(sort $(patsubst %/,%,$(dir $(PACKAGES_SRC))))
PACKAGES:=$(notdir $(PACKAGE_DIRS))

OBITOOLS_SRC:= $(wildcard cmd/obitools/*/*.go)
OBITOOLS_DIRS:=$(sort $(patsubst %/,%,$(dir $(OBITOOLS_SRC))))
OBITOOLS:=$(notdir $(OBITOOLS_DIRS))


define MAKE_PKG_RULE
pkg-$(notdir $(1)): $(1) pkg/obioptions/version.go
	@echo -n - Building package $(notdir $(1))...
	@$(GOBUILD) ./$(1) \
	    2> pkg-$(notdir $(1)).log \
		|| cat pkg-$(notdir $(1)).log
	@rm -f pkg-$(notdir $(1)).log
	@echo Done.
endef

define MAKE_OBITOOLS_RULE
$(OBITOOLS_PREFIX)$(notdir $(1)): $(BUILD_DIR) $(1) pkg/obioptions/version.go
	@echo -n - Building obitool $(notdir $(1))...
	@$(GOBUILD)  -o $(BUILD_DIR)/$(OBITOOLS_PREFIX)$(notdir $(1)) ./$(1) \
	             2> $(OBITOOLS_PREFIX)$(notdir $(1)).log \
				 || cat $(OBITOOLS_PREFIX)$(notdir $(1)).log
	@rm -f $(OBITOOLS_PREFIX)$(notdir $(1)).log
	@echo Done.
endef

GIT=$(shell which git 2>&1 >/dev/null && which git)
GITDIR=$(shell ls -d .git 2>/dev/null && echo .git || echo)
ifneq ($(strip $(GIT)),)
ifneq ($(strip $(GITDIR)),)
COMMIT_ID:=$(shell $(GIT) log -1 HEAD --format=%h)
LAST_TAG:=$(shell $(GIT) describe --tags $$($(GIT) rev-list --tags --max-count=1) | \
      	        tr '_' ' ')
endif
endif

OUTPUT:=$(shell mktemp)

all: obitools

packages: $(patsubst %,pkg-%,$(PACKAGES))
obitools: $(patsubst %,$(OBITOOLS_PREFIX)%,$(OBITOOLS)) 

update-deps:
	go get -u ./...

test:
	$(GOTEST) ./...

obitests: obitools
	for t in $$(find obitests -name test.sh -print) ; do \
		bash $${t} ; \
	done
	
man: 
	make -C doc man
obibook: 
	make -C doc obibook
doc: man obibook

macos-pkg: 
	@bash pkgs/macos/macos-installer-builder-master/macOS-x64/build-macos-x64.sh \
		OBITools \
		0.0.1

$(BUILD_DIR):
	mkdir -p $@


$(foreach P,$(PACKAGE_DIRS),$(eval $(call MAKE_PKG_RULE,$(P))))

$(foreach P,$(OBITOOLS_DIRS),$(eval $(call MAKE_OBITOOLS_RULE,$(P))))

pkg/obioptions/version.go: .FORCE
ifneq ($(strip $(COMMIT_ID)),)
	@cat $@ \
	| sed  -E 's/^var _Commit = "[^"]*"/var _Commit = "'$(COMMIT_ID)'"/' \
	| sed  -E 's/^var _Version = "[^"]*"/var _Version = "'"$(LAST_TAG)"'"/' \
	> $(OUTPUT)

	@diff $@ $(OUTPUT) 2>&1 > /dev/null \
    	|| echo "Update version.go : $@ to $(LAST_TAG) ($(COMMIT_ID))" \
    	&& mv $(OUTPUT) $@

	@rm -f $(OUTPUT)
endif

.PHONY: all packages obitools man obibook doc update-deps obitests .FORCE
.FORCE: