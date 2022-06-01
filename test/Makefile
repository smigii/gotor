FILES    = bigfile medfile multifile
TORS     = $(patsubst %, %.torrent, $(FILES))
ARCHIVES = $(patsubst %, %.tar.gz, $(FILES))
TRACKER  = http://opentracker.i2p.rocks:6969/announce
MKTOR    = transmission-create -c "Gotor testing file(s)" -t $(TRACKER)
COMPRESS = tar -czvf

all: torrents archives

torrents: $(TORS)

archives: $(ARCHIVES)

# Create torrent
%.torrent: %
	$(MKTOR) $<

# Compress
%.tar.gz: %
	$(COMPRESS) $<.tar.gz $<

.PHONY:
clean:
	rm $(TORS) $(ARCHIVES)