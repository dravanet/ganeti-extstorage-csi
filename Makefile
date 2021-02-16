# build & install targets for ganeti-extstorage-csi

PROJECT = ganeti-extstorage-csi

DESTDIR =
PREFIX = /usr/local
BINDIR = $(PREFIX)/bin
LIBDIR = $(PREFIX)/lib/$(PROJECT)

GO = go

SOURCES = $(shell find . -name *.go)
BINARY = $(PROJECT)

all: $(BINARY)

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 $(GO) build -o $@ ./cmd/$(PROJECT)

clean:
	rm -f $(BINARY)

install: $(BINARY)
	install -m 755 -o 0 -g 0 -D -t $(DESTDIR)$(LIBDIR) $(BINARY)
	install -m 755 -o 0 -g 0 -d $(DESTDIR)$(BINDIR)
	sed -e "s,@LIBDIR@,$(LIBDIR),g" share/ganeti-extstorage-csi-install > $(DESTDIR)$(BINDIR)/ganeti-extstorage-csi-install
	chmod 555 $(DESTDIR)$(BINDIR)/ganeti-extstorage-csi-install
