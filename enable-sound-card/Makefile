CC=gcc
server: enable-sound-card.c
	$(CC) -Wall -pthread -g -rdynamic -o enable-sound-card enable-sound-card.c -lrt

install:
	install -d -m755  $(DESTDIR)/bin/
	install -m755 server $(DESTDIR)/bin/enable-sound-card
