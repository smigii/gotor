FILES    = bigfile medfile multifile tinyfile
TORS     = $(patsubst %, %.torrent, $(FILES))
ARCHIVES = $(patsubst %, %.tar.gz, $(FILES))
TRACKER  = http://opentracker.i2p.rocks:6969/announce
MKTOR    = transmission-create -c "Gotor testing file(s)" -t $(TRACKER)
COMPRESS = tar -czvf

all: torrents archives

torrents: $(TORS)

archives: $(ARCHIVES)

# Set small piece size
tinyfile.torrent:
	$(MKTOR) -s 1 tinyfile

# Create torrent
%.torrent: %
	$(MKTOR) $<

# Compress
%.tar.gz: %
	$(COMPRESS) $<.tar.gz $<

.PHONY:
clean:
	rm $(TORS) $(ARCHIVES)
