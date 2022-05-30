FILES   = bigfile medfile multifile
TRACKER = http://opentracker.i2p.rocks:6969/announce
TOR_CMD = transmission-create -c "Gotor testing file(s)" -t $(TRACKER)

all: bigfile.torrent medfile.torrent multifile.torrent

bigfile.torrent: bigfile
	tar cfz bigfile.tgz bigfile
	$(TOR_CMD) bigfile

medfile.torrent: medfile
	tar cfz medfile.tgz medfile
	$(TOR_CMD) medfile

multifile.torrent: multifile
	tar cfz multifile.tgz multifile
	$(TOR_CMD) multifile

