LUAC=env luac
LUA=env lua
SRC=./
OS := $(shell uname -s)

all: luna-l2 lasm lcc
.PHONY = clean luna-l1

luna-l1:
	sudo mkdir -p /usr/local/bin/lvm
	sudo $(LUAC) -o /usr/local/bin/lvm/luna-l1 $(SRC)/l1/luna_l1.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/luna-l1 "$$@"' >> /usr/local/bin/luna-l1
	sudo $(LUAC) -o /usr/local/bin/lvm/lcc-l1 $(SRC)/l1/lcc.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lcc-l1 "$$@"' >> /usr/local/bin/lcc-l1
	sudo $(LUAC) -o /usr/local/bin/lvm/lasm-l1 $(SRC)/l1/lasm.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lasm-l1 "$$@"' >> /usr/local/bin/lasm-l1	
	sudo chmod +x /usr/local/bin/luna-l1
	sudo chmod +x /usr/local/bin/lcc-l1
	sudo chmod +x /usr/local/bin/lasm-l1

luna-l2:
	cd l2 && sudo go build -o /usr/local/bin/luna-l2 ./luna_l2.go

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
	sudo rm -f /usr/local/bin/luna-l1
	sudo rm -f /usr/local/bin/luna-l2
	sudo rm -f /usr/local/bin/lasm
	sudo rm -f /usr/local/bin/lcc
	sudo rm -rf /usr/local/bin/lvm
