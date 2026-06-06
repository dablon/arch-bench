CC ?= gcc
CFLAGS ?= -O2 -Wall -Wextra -Werror -std=c11 -D_GNU_SOURCE -I/usr/include/node
SYMDIR := libssl-sym
LDFLAGS ?= -L$(SYMDIR) -lssl -lcrypto
SRC = src/main.c
BIN = bin/c-flat-http

all: $(BIN)

# Create unversioned .so symlinks in a side dir (no -dev installed, only .so.3).
$(shell mkdir -p $(SYMDIR))
$(shell [ -f $(SYMDIR)/libssl.so ]    || ln -sf /usr/lib/x86_64-linux-gnu/libssl.so.3    $(SYMDIR)/libssl.so)
$(shell [ -f $(SYMDIR)/libcrypto.so ] || ln -sf /usr/lib/x86_64-linux-gnu/libcrypto.so.3 $(SYMDIR)/libcrypto.so)

$(BIN): $(SRC)
	@mkdir -p bin
	$(CC) $(CFLAGS) $< -o $@ $(LDFLAGS)

test: $(BIN)
	./tests/test.sh

clean:
	rm -rf bin $(SYMDIR)
