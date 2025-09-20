# Luna<br>
A simple, lightweight RISC CPU architecture.<br><br>

# Requirements<br>
- Windows, MacOS, Linux, or FreeBSD<br>
- Lua (only if you want to run the legacy Luna L1 virtual machine and tools)<br>
- Go <if you are compiling manually><br><br>

# Manual Installation (MacOS, Linux, FreeBSD)<br>
- Clone the repository using `git clone`<br>
- Navigate into the directory<br>
- Run `make; make install` to install the Luna L2 emulator and toolchain<br>
- Run `luna-l2 <disk image>` to run an application<br>
- Note: if you would like to install the legacy Luna L1 emulator and toolchain, run `make legacy` to install it as well as the assembler and C compiler. Then run `luna-l1 <disk image>` to run an application<br>

# Manual Installation (Windows)<br>
- Clone the repository using `git clone` or download it as a ZIP and then unzip it<br>
- Open the directory<br>
- Run the `build_windows.bat` file to build the Luna L2 emulator and toolchain<br>
- Install the applications into your PATH variable<br>
- Run `luna-l2 <disk image>` to run an application<br>

# Running an application<br>
- Run `luna-l2 <disk image>` to run an application.<br>
- To run a disk image with Luna L1, run `luna-l1 <disk image>` instead.<br><br>
