FILENAME=linuxkit

build:
	moby build -disable-content-trust -format=raw $(FILENAME).yml
	du -sh $(FILENAME).raw
