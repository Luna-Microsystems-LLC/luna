LUAC=env luac
LUA=env lua
SRC=./

all: luna lasm lcc
.PHONY = clean

luna: $(SRC)/luna.lua
	sudo mkdir -p /usr/local/bin/lvm
	sudo $(LUAC) -o /usr/local/bin/lvm/luna $(SRC)/luna.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/luna "$$@"' >> /usr/local/bin/luna
	sudo chmod +x /usr/local/bin/luna


lasm: $(SRC)/lasm.lua
	sudo mkdir -p /usr/local/bin/lvm
	sudo $(LUAC) -o /usr/local/bin/lvm/lasm $(SRC)/lasm.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lasm "$$@"' >> /usr/local/bin/lasm
	sudo chmod +x /usr/local/bin/lasm

lcc: $(SRC)/lcc.lua
	sudo mkdir -p /usr/local/bin/lvm
	sudo $(LUAC) -o /usr/local/bin/lvm/lcc $(SRC)/lcc.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lcc "$$@"' >> /usr/local/bin/lcc
	sudo chmod +x /usr/local/bin/lcc

clean:
	sudo rm /usr/local/bin/luna
	sudo rm /usr/local/bin/lasm
	sudo rm /usr/local/bin/lcc
	sudo rm -rf /usr/local/bin/lvm
