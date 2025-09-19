## Documentation
[Jump to preamble](#preamble)<br>
[Jump to registers](#registers)
[Jump to instructions](#instructions)<br>
[Jump to interrupts](#interrupts)<br>
[Jump to assembly](#assembly)<br>
[Jump to linking](#linking)<br>
[Jump to frontend](#frontend)<br><br>

## Preamble
The Luna L2 is a simple, lightweight, RISC CPU that aims to be clean while also leveraging some luxuries from CISC, with the ultimate end goal of being easy to teach and learn.<br><br>

## Registers
L2 has 30 total registers for the storing and manipulation of data<br>
R0-R12: general purpose registers, all can be written to and read from<br>
T1-T12: temporary registers. The standard calling convention uses registers T1-T7<br>
SP: stack pointer<br>
PC: program counter/instruction pointer<br>
RE1-RE3: reserved registers (you may use RE3 for storing PC when using loops)<br><br>

## Instructions
The Luna L2 has 25 unique instructions that allow the CPU to interact with registers, memory, and the BIOS<br><br>

1. MOV: moves a value from the source to the destination; source can be register or immediate.<br>
2. HLT: stops the CPU from executing instructions.<br>
3. JMP: sets the program counter to the specified address; address can be register or immediate.<br>
4. INT: calls a BIOS interrupt. [Jump to interrupts](#interrupts)<br> 
5. JNZ: sets the program counter to the specified address if the register is not zero; address can be immediate or register.<br>
6. NOP: stalls the CPU for 1 cycle.<br>
7. CMP: sets the specified register to 1 if the other two registers are the same; otherwise sets to 0.<br>
8. JZ: sets the program counter to the specified address if the register is zero.<br>
9. INC: increments a register by 1.<br>
10. DEC: decrements a register by 1.<br>
11. PUSH: Pushes a word to the stack and increments the stack pointer by 2; word can be in a register or an immediate.<br>
12. POP: Pops a word off the stack to the specified register and decrements the stack pointer by 2.<br>
13. ADD: Puts the sum of 2 registers into a register.<br>
14. SUB: Puts the subtraction result of 2 registers into a register.<br>
15. MUL: Puts the product of 2 registers into a register.<br>
16. DIV: Puts the quotient of 2 registers into a register.<br>
17. IGT: Sets a register to 1 if the second register is greater than the third register; otherwise 0.<br>
18. ILT: Sets a register to 1 if the second register is less than the third register; otherwise 0.<br>
19. AND: performs bitwise AND on two registers and puts the result to a register.<br>
20. OR:  performs bitwise OR on two registers and puts the result to a register.<br>
21. NOR:  performs bitwise NOR on two registers and puts the result to a register.<br>
22. NOT: performs bitwise NOT on two registers and puts the result to a register.<br> 
23. XOR: performs bitwise XOR on two registers and puts the result to a register.<br>
24. LOD: loads a word from memory to a register. (bytewise)<br>
25. STR: stores a value to a memory address from a register. (bytewise)<br><br>

## Interrupts
Because the Luna L2 is a primitive CPU, it does not support directly interacting with things like VRAM or input devices from raw instructions. Instead, you must use an interrupt and allow the BIOS to carry out the tasks. (Note: these are for the integrated BIOS, other BIOSes may have different interrupts.)<br><br>

1. Print character to screen (char in r1, foreground in r2, background in r3)<br>
2. Sleep (seconds in r1)<br>
3. Write to VRAM (bytewise) (address in r1, value in r2)<br>
4. Toggle keyboard echo (mode in r1, 1 for echo char back, 0 for no echo)<br>
5. Reserved; do not use<br><br>

## Assembly
The L2 architecture has a custom assembler (`las`) to convert programs from assembly language (.asm, .s, .S) to machine code (.o) that can then be linked and then run on L2.<br>
# Syntax specifications
The syntax of L2 assembly is similar to that of Intel assembly syntax. An instruction consists of a mnemonic, then the operands. Above, there were no specifications on which instructions use which registers, since every instruction that uses registers can use any register.<br>
Except for STR, the destination register is always the first register in the instruction.<br>
You can use a label name followed by a colon to make a label, which gets turned into a numerical offset at assembly time. Therefore you can treat them as numbers as well. These can also be used as functions with `call` and `ret`.<br>
# Custom directives
There are some directives in LAS that do not correspond to any instruction on L2. They are as follows:<br>
`call`: calculates the return address, pushes it onto the stack, and jumps to the label specified (`call mylabel`)<br>
`ret`: jumps to the value in register `re1`<br>
`.ascii`: defines a sequence of ASCII bytes, wrapped in quotation marks<br>
`.asciz`: defines a sequence of ASCII bytes, wrapped in quotation marks (null terminated)<br>
# Examples
`mov r1, 5` (destination: r1, source: 5)<br>
`pop r1` (destination: r1)<br>
`add r3, r1, r2` (destination: r3, source 1: r1, source 2: r2)<br>
`mylabel:
    mov r1, 1
    ret`<br>
`call mylabel`<br>
`.ascii "Hello world!"`<br>
# Assembling a program
To assemble a program, use the following: `las <flags> <input file(s)> -o <output file>`<br>
The flags are as follows:<br>
`-v`: shows the version of LAS and exits.<br>
`-c`: do not invoke linker (`l2ld`) after assembly is complete.<br>
Note: you may also use the Luna Compiler Collection frontend (`lcc`) with the same syntax to do this.<br><br>

## Linking
Linking is the process of resolving offsets in an object file (.o) to turn it into an executable file (.bin). L2 also has a custom linker (`l2ld`) to convert L2 object format (L2O) to L2 executable format (L2E).<br>
# Linking a program
To link a program, use the following: `l2ld <flags> <input file(s)> -o <output file>`<br>
The flags are as follows:<br>
`-v`: shows the version of L2LD and exits.<br>
Note: you may also use the Luna Compiler Collection frontend (`lcc`) with the same syntax to do this.<br><br>

## Frontend
L2 has a simple compiler frontend (`lcc`) which is similar to that of `gcc` or `clang`. LCC can also automatically detect file type and use the relevant programs to compile it to an executable.<br>
# Compiling a program
To compile a program, use the following: `lcc <flags> <input file(s)> -o <output file>`<br>
The flags are as follows:<br>
`-c`: do not invoke linker (`l2ld`) after assembly is complete.<br>
`-v`: shows the version of LCC and exits.<br>
`-s`: do not invoke assembler (`las`) after compilation is complete.<br>
Supported file types: (subject to change)<br>
`.s`: assembly
`.S`: assembly
`.asm`: assembly
`.o`: object file
`.c`: C (upcoming)
