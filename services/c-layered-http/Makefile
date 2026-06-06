CC ?= gcc
CFLAGS ?= -O2 -Wall -Wextra -Werror -std=c11 -D_GNU_SOURCE -I/usr/include/node
SYMDIR := libssl-sym
LDFLAGS ?= -L$(SYMDIR) -lssl -lcrypto
SRCS := $(shell find src -name '*.c')
OBJS := $(SRCS:.c=.o)
BIN = bin/c-layered-http

all: $(BIN)

$(shell mkdir -p $(SYMDIR))
$(shell [ -f $(SYMDIR)/libssl.so ]    || ln -sf /usr/lib/x86_64-linux-gnu/libssl.so.3    $(SYMDIR)/libssl.so)
$(shell [ -f $(SYMDIR)/libcrypto.so ] || ln -sf /usr/lib/x86_64-linux-gnu/libcrypto.so.3 $(SYMDIR)/libcrypto.so)

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

$(BIN): $(OBJS)
	@mkdir -p bin
	$(CC) $(OBJS) -o $@ $(LDFLAGS)

test: $(BIN)
	./tests/test.sh

clean:
	rm -rf bin $(SYMDIR) $(OBJS)
