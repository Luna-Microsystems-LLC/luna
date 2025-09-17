# Luna<br>
A simple, lightweight RISC CPU architecture.<br><br>

# Requirements<br>
- Windows, MacOS, Linux, or FreeBSD<br>
- Lua (only if you want to run the legacy Luna L1 virtual machine and tools)<br>
- Go <if you are compiling manually><br><br>

# Manual Installation (MacOS, Linux, FreeBSD)<br>
- Clone the repository using `git clone`<br>
- Navigate into the directory<br>
- Run `make; make install` to install the Luna L2 virtual machine, the Luna Compiler collection (`lcc`), and the Luna linker (`lld`)<br>
- Run `luna-l2 <disk image>` to run an application.<br>
- Note: if you would like to install the legacy Luna L1 architecture, run `make legacy` to install it as well as the assembler and C compiler. Then run `luna-l1 <disk image>` to run an application.<br>

# Manual Installation (Windows)<br>
- Download the Luna repository as a .zip file and then unzip it.<br>
- Open Command Prompt or PowerShell and navigate into the unzipped directory.<br>
- Navigate into the `l2` directory<br>
- Run `go build -o luna-l2 luna_l2.go`<br>
- Repeat as necessary for the assembler, linker, and frontend `lcc`.<br><br>

# Running an application<br>
- Run `luna-l2 <disk image>` to run an application.<br>
- To run a disk image with Luna L1, run `luna-l1 <disk image>` instead.<br><br>
