#export GOPATH=$(shell pwd)/GO
#export GOBIN=$(GOPATH)/bin
#export PATH=$(GOBIN):$(shell echo $${PATH})

.DEFAULT_GOAL := all

GREEN  := \033[0;32m
YELLOW := \033[0;33m
BLUE   := \033[0;34m
NC     := \033[0m

GOFLAGS=
LDFLAGS=
GOCMD=go
GOBUILD=$(GOCMD) build $(GOFLAGS) $(if $(LDFLAGS),-ldflags='$(LDFLAGS)')
GOGENERATE=$(GOCMD) generate
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BUILD_DIR=build
OBITOOLS_PREFIX:=

PACKAGES_SRC:= $(wildcard pkg/*/*.go pkg/*/*/*.go)
PACKAGE_DIRS:=$(sort $(patsubst %/,%,$(dir $(PACKAGES_SRC))))
PACKAGES:=$(notdir $(PACKAGE_DIRS))

GITHOOK_SRC_DIR=git-hooks
GITHOOKS_SRC:=$(wildcard $(GITHOOK_SRC_DIR)/*)

GITHOOK_DIR=.git/hooks
GITHOOKS:=$(patsubst $(GITHOOK_SRC_DIR)/%,$(GITHOOK_DIR)/%,$(GITHOOKS_SRC))

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

help:
	@printf "$(GREEN)OBITools4 Makefile$(NC)\n\n"
	@printf "$(BLUE)Main targets:$(NC)\n"
	@printf "  %-20s %s\n" "all"          "Build all obitools (default)"
	@printf "  %-20s %s\n" "obitools"     "Build all obitools binaries to build/"
	@printf "  %-20s %s\n" "test"         "Run Go unit tests"
	@printf "  %-20s %s\n" "obitests"     "Run integration tests (obitests/)"
	@printf "  %-20s %s\n" "bump-version" "Increment patch version (or set with VERSION=x.y.z)"
	@printf "  %-20s %s\n" "update-deps"  "Update all Go dependencies"
	@printf "\n$(BLUE)Jujutsu workflow:$(NC)\n"
	@printf "  %-20s %s\n" "jjnew"        "Document current commit and start a new one"
	@printf "  %-20s %s\n" "jjpush"       "Release: describe, bump, generate notes, push PR, tag (VERSION=x.y.z optional)"
	@printf "  %-20s %s\n" "jjfetch"      "Fetch latest commits from origin"
	@printf "\n$(BLUE)Required tools:$(NC)\n"
	@printf "  %-20s " "go";      command -v go      >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(go version)" || printf "$(YELLOW)✗ not found$(NC)\n"
	@printf "  %-20s " "git";     command -v git     >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(git --version)" || printf "$(YELLOW)✗ not found$(NC)\n"
	@printf "  %-20s " "jj";      command -v jj      >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(jj --version)" || printf "$(YELLOW)✗ not found$(NC)\n"
	@printf "  %-20s " "gh";      command -v gh      >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(gh --version | head -1)" || printf "$(YELLOW)✗ not found$(NC)  (brew install gh)\n"
	@printf "\n$(BLUE)Optional tools (release notes generation):$(NC)\n"
	@printf "  %-20s " "aichat";  command -v aichat  >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(aichat --version)" || printf "$(YELLOW)✗ not found$(NC)  (https://github.com/sigoden/aichat)\n"
	@printf "  %-20s " "jq";      command -v jq      >/dev/null 2>&1 && printf "$(GREEN)✓$(NC) %s\n" "$$(jq --version)" || printf "$(YELLOW)✗ not found$(NC)  (brew install jq)\n"

all: install-githook obitools

obitools: $(patsubst %,$(OBITOOLS_PREFIX)%,$(OBITOOLS))

install-githook: $(GITHOOKS)

$(GITHOOK_DIR)/%: $(GITHOOK_SRC_DIR)/%
	@echo installing $$(basename $@)...
	@mkdir -p $(GITHOOK_DIR)
	@cp $< $@
	@chmod +x $@


update-deps:
	go get -u ./...

test: .FORCE
	$(GOTEST) ./...

obitests:
	@for t in $$(find obitests -name test.sh -print) ; do \
		bash $${t} || exit 1;\
	done

githubtests: obitools obitests

$(BUILD_DIR):
	mkdir -p $@


$(foreach P,$(PACKAGE_DIRS),$(eval $(call MAKE_PKG_RULE,$(P))))

$(foreach P,$(OBITOOLS_DIRS),$(eval $(call MAKE_OBITOOLS_RULE,$(P))))

pkg/obioptions/version.go: version.txt .FORCE
	@version=$$(cat version.txt); \
	cat $@ \
	| sed  -E 's/^var _Version = "[^"]*"/var _Version = "Release '$$version'"/' \
	> $(OUTPUT)

	@diff $@ $(OUTPUT) 2>&1 > /dev/null \
    	|| (echo "Update version.go to $$(cat version.txt)" && mv $(OUTPUT) $@)

	@rm -f $(OUTPUT)

bump-version:
	@current=$$(cat version.txt); \
	if [ -n "$(VERSION)" ]; then \
		new_version="$(VERSION)"; \
		echo "Setting version to $$new_version (was $$current)"; \
	else \
		echo "Incrementing version..."; \
		echo "  Current version: $$current"; \
		major=$$(echo $$current | cut -d. -f1); \
		minor=$$(echo $$current | cut -d. -f2); \
		patch=$$(echo $$current | cut -d. -f3); \
		new_patch=$$((patch + 1)); \
		new_version="$$major.$$minor.$$new_patch"; \
		echo "  New version: $$new_version"; \
	fi; \
	echo "$$new_version" > version.txt
	@echo "✓ Version updated in version.txt"
	@$(MAKE) pkg/obioptions/version.go

jjnew:
	@echo "$(YELLOW)→ Creating a new commit...$(NC)"
	@echo "$(BLUE)→ Documenting current commit...$(NC)"
	@jj auto-describe
	@echo "$(BLUE)→ Done.$(NC)"
	@jj new
	@echo "$(GREEN)✓ New commit created$(NC)"

jjpush:
	@$(MAKE) jjpush-describe
	@$(MAKE) jjpush-bump
	@$(MAKE) jjpush-notes
	@$(MAKE) jjpush-push
	@$(MAKE) jjpush-tag
	@echo "$(GREEN)✓ Release complete$(NC)"

jjpush-describe:
	@echo "$(BLUE)→ Documenting current commit...$(NC)"
	@jj auto-describe

jjpush-bump:
	@echo "$(BLUE)→ Creating new commit for version bump...$(NC)"
	@jj new
	@$(MAKE) bump-version

jjpush-notes:
	@version=$$(cat version.txt); \
	echo "$(BLUE)→ Generating release notes for version $$version...$(NC)"; \
	release_title="Release $$version"; \
	release_body=""; \
	if command -v aichat >/dev/null 2>&1; then \
		previous_tag=$$(git describe --tags --abbrev=0 --match 'Release_*' 2>/dev/null); \
		if [ -z "$$previous_tag" ]; then \
			echo "$(YELLOW)⚠ No previous Release tag found, skipping release notes$(NC)"; \
		else \
			raw_output=$$(git log --format="%h %B" "$$previous_tag..HEAD" | \
				aichat \
				"Summarize the following commits into a GitHub release note for version $$version. Ignore commits related to version bumps, .gitignore changes, or any internal housekeeping that is irrelevant to end users. Describe each user-facing change precisely without exposing code. Eliminate redundancy. Output strictly valid JSON with no surrounding text, using this exact schema: {\"title\": \"<short release title>\", \"body\": \"<detailed markdown release notes>\"}" 2>/dev/null) || true; \
			if [ -n "$$raw_output" ]; then \
				notes=$$(printf '%s\n' "$$raw_output" | python3 tools/json2md.py 2>/dev/null); \
				if [ -n "$$notes" ]; then \
					release_title=$$(echo "$$notes" | head -1); \
					release_body=$$(echo "$$notes" | tail -n +3); \
				else \
					echo "$(YELLOW)⚠ JSON parsing failed, using default release message$(NC)"; \
				fi; \
			fi; \
		fi; \
	fi; \
	printf '%s' "$$release_title" > /tmp/obitools4-release-title.txt; \
	printf '%s' "$$release_body" > /tmp/obitools4-release-body.txt; \
	echo "$(BLUE)→ Setting release notes as commit description...$(NC)"; \
	jj desc -m "$$release_title"$$'\n\n'"$$release_body"

jjpush-push:
	@echo "$(BLUE)→ Pushing commits...$(NC)"
	@jj git push --change @
	@echo "$(BLUE)→ Creating/updating PR...$(NC)"
	@release_title=$$(cat /tmp/obitools4-release-title.txt 2>/dev/null || echo "Release $$(cat version.txt)"); \
	release_body=$$(cat /tmp/obitools4-release-body.txt 2>/dev/null || echo ""); \
	branch=$$(jj log -r @ --no-graph -T 'bookmarks.map(|b| b.name()).join("\n")' 2>/dev/null | head -1); \
	if [ -n "$$branch" ] && command -v gh >/dev/null 2>&1; then \
		gh pr create --title "$$release_title" --body "$$release_body" --base master --head "$$branch" 2>/dev/null \
		|| gh pr edit "$$branch" --title "$$release_title" --body "$$release_body" 2>/dev/null \
		|| echo "$(YELLOW)⚠ Could not create/update PR$(NC)"; \
	fi

jjpush-tag:
	@version=$$(cat version.txt); \
	tag_name="Release_$$version"; \
	release_title=$$(cat /tmp/obitools4-release-title.txt 2>/dev/null || echo "Release $$version"); \
	release_body=$$(cat /tmp/obitools4-release-body.txt 2>/dev/null || echo ""); \
	install_section=$$'\n## Installation\n\n### Pre-built binaries\n\nDownload the appropriate archive for your system from the\n[release assets](https://github.com/metabarcoding/obitools4/releases/tag/Release_'"$$version"')\nand extract it:\n\n#### Linux (AMD64)\n```bash\ntar -xzf obitools4_'"$$version"'_linux_amd64.tar.gz\n```\n\n#### Linux (ARM64)\n```bash\ntar -xzf obitools4_'"$$version"'_linux_arm64.tar.gz\n```\n\n#### macOS (Intel)\n```bash\ntar -xzf obitools4_'"$$version"'_darwin_amd64.tar.gz\n```\n\n#### macOS (Apple Silicon)\n```bash\ntar -xzf obitools4_'"$$version"'_darwin_arm64.tar.gz\n```\n\nAll OBITools4 binaries are included in each archive.\n\n### From source\n\nYou can also compile and install OBITools4 directly from source using the\ninstallation script:\n\n```bash\ncurl -L https://raw.githubusercontent.com/metabarcoding/obitools4/master/install_obitools.sh | bash -s -- --version '"$$version"'\n```\n\nBy default binaries are installed in `/usr/local/bin`. Use `--install-dir` to\nchange the destination and `--obitools-prefix` to add a prefix to command names:\n\n```bash\ncurl -L https://raw.githubusercontent.com/metabarcoding/obitools4/master/install_obitools.sh | \\\n  bash -s -- --version '"$$version"' --install-dir ~/local --obitools-prefix k\n```\n'; \
	release_message="$$release_title"$$'\n\n'"$$release_body$$install_section"; \
	echo "$(BLUE)→ Creating tag $$tag_name...$(NC)"; \
	git tag -a "$$tag_name" -m "$$release_message" 2>/dev/null || echo "$(YELLOW)⚠ Tag $$tag_name already exists$(NC)"; \
	echo "$(BLUE)→ Pushing tag $$tag_name...$(NC)"; \
	git push origin "$$tag_name" 2>/dev/null || echo "$(YELLOW)⚠ Tag push failed or already pushed$(NC)"; \
	rm -f /tmp/obitools4-release-title.txt /tmp/obitools4-release-body.txt

jjfetch:
	@echo "$(YELLOW)→ Pulling latest commits...$(NC)"
	@jj git fetch
	@jj new master@origin
	@echo "$(GREEN)✓ Latest commits pulled$(NC)"

.PHONY: all obitools update-deps obitests githubtests help jjnew jjpush jjpush-describe jjpush-bump jjpush-notes jjpush-push jjpush-tag jjfetch bump-version .FORCE
.FORCE:
