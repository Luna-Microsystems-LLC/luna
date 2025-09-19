## Documentation
[Jump to preamble](#preamble)<br>

## Preamble
The Luna L2 is a simple, lightweight, RISC CPU that aims to be clean while also leveraging some luxuries from CISC, with the ultimate end goal of being easy to teach and learn.<br><br>

## Instructions
The Luna L2 has 25 unique instructions that allow the CPU to interact with registers, memory, and the BIOS<br><br>

1. MOV: moves a value from the source to the destination; source can be register or immediate.<br>
2. HLT: stops the CPU from executing instructions.<br>
3. JMP: sets the program counter to the specified address; address can be register or immediate.<br>
4. INT: calls a BIOS interrupt. (Jump to interrupts)[#interrupts]<br> 
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

