LUAC=env luac
LUA=env lua
SRC=./

all: bin/luna-l2 bin/las bin/lld bin/lcc bin/lcc1 
legacy: luna-l1
.PHONY: clean install

luna-l1:
	sudo mkdir -p /usr/local/bin/lvm
	sudo $(LUAC) -o /usr/local/bin/lvm/luna-l1 $(SRC)/l1/luna_l1.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/luna-l1 "$$@"' > /usr/local/bin/luna-l1
	sudo $(LUAC) -o /usr/local/bin/lvm/lcc-l1 $(SRC)/l1/lcc.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lcc-l1 "$$@"' > /usr/local/bin/lcc-l1
	sudo $(LUAC) -o /usr/local/bin/lvm/lasm-l1 $(SRC)/l1/lasm.lua
	sudo printf '#!/bin/sh\n $(LUA) /usr/local/bin/lvm/lasm-l1 "$$@"' > /usr/local/bin/lasm-l1	
	sudo chmod +x /usr/local/bin/luna-l1
	sudo chmod +x /usr/local/bin/lcc-l1
	sudo chmod +x /usr/local/bin/lasm-l1

bin/luna-l2: $(SRC)/l2/luna_l2.go
	cd l2 && go build -o ../bin/luna-l2 ./luna_l2.go

bin/las: $(SRC)/las/las.go
	cd las && go build -o ../bin/las ./las.go

bin/lcc1: $(SRC)/lcc1/lcc1.go
	cd lcc1 && go build -o ../bin/lcc1 ./lcc1.go

bin/lcc: $(SRC)/lcc/lcc.go
	cd lcc && go build -o ../bin/lcc ./lcc.go

bin/lld: $(SRC)/lld/lld.go
	cd lld && go build -o ../bin/lld ./lld.go

install:
	sudo cp bin/* /usr/local/bin/

clean:
	sudo rm -rf /usr/local/bin/lvm
	sudo rm -f /usr/local/bin/luna-l1
	sudo rm -f /usr/local/bin/lasm-l1
	sudo rm -f /usr/local/bin/lcc-l1
	sudo rm -f /usr/local/bin/luna-l2
	sudo rm -f /usr/local/bin/las
	sudo rm -f /usr/local/bin/lcc1
	sudo rm -f /usr/local/bin/lcc
	sudo rm -f /usr/local/bin/lld
