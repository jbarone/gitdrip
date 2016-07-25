prefix=/usr/local

# files that need mode 755
EXEC_FILES=git-drip

# files that need mode 644
SCRIPT_FILES =git-drip-init
SCRIPT_FILES+=git-drip-feature
SCRIPT_FILES+=git-drip-hotfix
SCRIPT_FILES+=git-drip-release
SCRIPT_FILES+=git-drip-version
SCRIPT_FILES+=gitdrip-common
SCRIPT_FILES+=gitdrip-shFlags

all:
	@echo "usage: make install"
	@echo "       make uninstall"

install:
	install -d -m 0755 $(prefix)/bin
	install -m 0755 $(EXEC_FILES) $(prefix)/bin
	install -m 0644 $(SCRIPT_FILES) $(prefix)/bin

uninstall:
	test -d $(prefix)/bin && \
	cd $(prefix)/bin && \
	rm -f $(EXEC_FILES) $(SCRIPT_FILES)
