if os.getenv("OS") ~= "Windows_NT" then
    io.write("\27]0;Luna VM Runtime\7")
    io.flush()
else
    os.execute("title \"Luna VM Runtime\"")
end

-- Variables
local execute
local memory = {} -- 65,535 byte maximum
local ogimg = {}
local registers = {
    {0x0, "R0", nil},
    {0x1, "R1", nil},
    {0x2, "R2", nil},
    {0x3, "R3", nil},
    {0x4, "R4", nil},
    {0x5, "R5", nil},
    {0x6, "R6", nil},
    {0x7, "R7", nil},
    {0x8, "R8", nil},
    {0x9, "R9", nil},
    {0xa, "R10", nil},
    {0xb, "R11", nil},
    {0xc, "R12", nil},
    {0xd, "SP", 0},
    {0xe, "PC", 0},
    {0xf, "T1", nil},
    {0x10, "T2", nil},
    {0x11, "T3", nil},
    {0x12, "T4", nil},
    {0x13, "T5", nil},
    {0x14, "T6", nil},
    {0x15, "T7", nil},
    {0x16, "T8", nil},
    {0x17, "T9", nil},
    {0x19, "T10", nil},
    {0x1a, "T11", nil},
    {0x1b, "T12", nil},
    {0x1c, "PTR", 0}
}
local PC = 0xe
local instructions = {
    NULL = 0x00;
    MOV = 0x01;
    JMP = 0x02;
    JNZ = 0x03;
    SYSCALL = 0x04;
    ADD = 0x05;
    SUB = 0x06;
    MUL = 0x07;
    DIV = 0x08;
    PUSH = 0x09;
    AND = 0x0b;
    OR = 0x0c;
    NOT = 0x0e;
    XOR = 0x0f;
    DB = 0x10;
    NOP = 0x11;
    DSTART = 0x12;
    DSEP = 0x13;
    DEND = 0x14;
    CMP = 0x15;
    FSTART = 0x16;
    FSEP = 0x17;
    FEND = 0x18;
    JZ = 0x19;
    DW = 0x1a;
    MBYTE = 0x1c; 
    STN = 0x1d;
    STACK = 0x1e;
    LDA = 0x1f;
    IGT = 0x7f;
    ILT = 0x80;
    IET = 0x81;
    IGET = 0x82;
    ILET = 0x83;
}

local ignorelist = {
    instructions.DSTART,
    instructions.DSEP,
    instructions.DEND,
    instructions.NULL,
    instructions.NOP,
    --instructions.FSTART,
    --instructions.FSEP,
    --instructions.FEND,
    instructions.DW,
    instructions.BSTART,
    instructions.BEND
}
local RunningLog = false
local Verbose = false

local function sleep(_time)
    if not tonumber(_time) then return end
    if os.getenv("OS") ~= "Windows_NT" then
        os.execute("sleep " .. _time)
    else
        os.execute("timeout /t " .. _time .. " > NUL")
    end
end

local function Dump(file)
    if Verbose == true then
        print("===============\nMemory dump:")
        for i = 1, #memory do
            if memory[i] ~= nil then
                io.write(memory[i])
            end
        end
        io.write("\n")
        print(#memory .. " bytes in memory.\n===============")
        for i = 1, #registers do
            print("Register " .. registers[i][2] .. ": " .. (registers[i][3] or "nil"))
        end
    end
    if RunningLog == true then
		local handle = io.open(file, "a")
		if not handle then
			error("Could not open running log file.")
		end
		handle:write("=====LOG AT " .. registers[15][3] .. "=====\nMemory dump:\n")
		for i = 1, #memory do
			if memory[i] ~= nil then
				handle:write(memory[i])
			end
		end
		handle:write("\n" .. #memory .. " bytes in memory.\n==========\n")
		for i = 1, #registers do
			handle:write("Register " .. registers[i][2] .. ": " .. (registers[i][3] or "nil") .. "\n")
		end
		handle:close()
    end
end

local function is_byte(str, exp)
    if string.byte(str) == exp then
        return true
    else
        return false
    end
end

local function getValueFromRegister(addr)
    for _, register in pairs(registers) do
        if register[1] == addr then
            return register[3]
        end
    end
    return "REGISTER_NOT_EXISTENT"
end

local function _tonumber(number, silent)
    local n = tonumber(number)

    if not n then
        if not silent then
            print(debug.traceback()); print("[FATAL]: Invalid number '" .. number .. "'")
            Dump()
            os.exit(1)
        end
        return nil
    end
 
    if n > 0x7FFF then
        n = n - 0x10000
    elseif n < -0x8000 then
        n = -0x8000
    end

    return n
end


local function lookupSymbol(name, grace)
    local offset_ = 0
    for i = 1, #memory do
        if is_byte(memory[i], instructions.DSTART) then
            for j = i, #memory do
                if is_byte(memory[j], instructions.DSEP) then
                    local str = ""
                    for k = i + 1, j - 1 do
                        str = str .. memory[k]
                    end
    
                    if name == str then
                        offset_ = j
                    end
                end
            end
        end
    end
    if offset_ == 0 then
        if not grace then
            print(debug.traceback()); print("[FATAL]: Unknown symbol '" .. (name or "(no name)") .. "'")
            Dump()
            os.exit(1)
        else
            return false
        end
    end
    for i = offset_ + 1, #memory do
        if is_byte(memory[i], instructions.DEND) then
            local str = ""
            for j = offset_ + 1, i - 1 do
                if getValueFromRegister(string.byte(memory[j])) and offset_ + 1 == i - 1 then
                    return getValueFromRegister(string.byte(memory[j]))
                else
                    str = str .. tostring(memory[j])
                end
            end
            if string.sub(str, 1, 1) == string.char(instructions.DW) then
                str = _tonumber(string.sub(str, 2, #str))
            end
            return str
        end
    end
end

local function DumpBytes(str)
    for i = 1, #str do
        print(tostring(string.byte(string.sub(str, i, i))))
    end
end 

local function setRegister(addr, value)
    for _, register in pairs(registers) do
        if register[1] == addr then
            register[3] = value
            return
        end
    end
    print(debug.traceback()); print("[FATAL]: Invalid register at " .. getValueFromRegister(PC))
    Dump()
    os.exit(1)
end

local function execLocation(name)
    local offset = 0
    for i = 1, #memory do
        if string.byte(memory[i]) == instructions.FSTART then
            local _end = 0
            for j = i + 1, #memory do
                if string.byte(memory[j]) == instructions.FSEP then
                    offset = j + 1
                    _end = j
                    break
                end
            end
            if _end == 0 then
                return
            end
            local name_ = ""
            for j = i + 1, _end - 1 do
                name_ = name_ .. memory[j]
            end

            if name_ == name then
                local _end = 0
                for j = offset, #memory do
                    if string.byte(memory[j]) == instructions.FEND then
                        _end = j
                        break
                    end
                end
                if _end == 0 then
                    return
                end 

                local start = offset
                local finish = _end - 1
                local startreal = offset
                ::execfunc:: 
                local newstart = execute(start, finish, startreal)
                
                if tonumber(newstart) then
                    start = newstart
                    finish = finish
                    startreal = startreal
                    if Verbose == true then
                        Dump()
                        print("Press ENTER to continue...")
                        io.read("*l")
                    end
                    goto execfunc
                end
                -- execute(offset, _end - 1, offset, true)
            end
        end
    end
end

local function loadMemory(character)
    if #memory < 65535 then
        table.insert(memory, character)
    end
end

local function convertToWord(str)
    math.randomseed(os.time())
    local chars = { "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z" }
    local name = ""
    for i = 1, 10 do
        name = name .. chars[math.random(1, #chars)]
    end
    local total = ""
    total = total .. string.char(instructions.DSTART)
    total = total .. name
    total = total .. string.char(instructions.DSEP)
    total = total .. str
    total = total .. string.char(instructions.NULL)
    total = total .. string.char(instructions.DEND)
    return {name, total}
end

local function syscallHandler()
    local func = getValueFromRegister(0x01)
    if func == 0x1 then
        -- Printing to stdout
        local msg = ""
        if not _tonumber(getValueFromRegister(0x02), true) or not _tonumber(getValueFromRegister(0x03), true) then
            local msgSymbol = getValueFromRegister(0x02) 
            msg = lookupSymbol(msgSymbol)
            if #tostring(msg) == 1 and getValueFromRegister(string.byte(msg)) ~= "REGISTER_NOT_EXISTENT" then
                msg = getValueFromRegister(string.byte(msg))
            end
        elseif _tonumber(getValueFromRegister(0x02), true) and _tonumber(getValueFromRegister(0x03), true) then
            for i = _tonumber(getValueFromRegister(0x02)), _tonumber(getValueFromRegister(0x03)) do 
                msg = msg .. (memory[i] or "\0")
            end
        end
        io.write(msg) 
    elseif func == 0x2 then
        -- Allocating memory
        local start = getValueFromRegister(0x02)
        local _end = getValueFromRegister(0x03)

        if not start then
            start = lookupSymbol(getValueFromRegister(0x02))
        end
        if not _end then
            _end = lookupSymbol(getValueFromRegister(0x03))
        end

        if not _tonumber(start) or not _tonumber(_end) then
            print(debug.traceback()); print("[FATAL]: Invalid number at " .. getValueFromRegister(PC))
            Dump()
            os.exit(1)
        end
        start = _tonumber(start)
        _end = _tonumber(_end)
        for i = start, _end do
            memory[i] = string.char(0x25)
        end
    elseif func == 0x3 then
        -- Reading user input
        local input = io.read("*l")
        local word = convertToWord(input)
        local str = word[2]
        local name = word[1]
        for i = 1, #str do
            loadMemory(string.sub(str, i, i))
        end
        local start = #memory + 1
        
        loadMemory(string.char(instructions.MOV))
        loadMemory(string.char(0x2))
        for i = 1, #name do
            loadMemory(tostring(string.sub(name, i, i)))
        end
        execute(start, #memory, start)
    elseif func == 0x4 then
        local code = _tonumber(lookupSymbol(getValueFromRegister(0x2), true), true)
        if not code then code = _tonumber(getValueFromRegister(0x2), true) end
        if not code then
            code = 0
        end
        os.exit(code)
    elseif func == 0x5 then
        -- Save to disk
        local contents = getValueFromRegister(0x02)
        local startaddr = _tonumber(getValueFromRegister(0x03))
        if lookupSymbol(contents, true) then
            contents = lookupSymbol(contents)
        end

        local fnull = #ogimg + 1

        for i = fnull, startaddr - 1 do
            ogimg[i] = "\0"
        end

        for i = 1, #contents do
            ogimg[startaddr + (i - 1)] = string.sub(contents, i, i)
            memory[startaddr + (i - 1)] = string.sub(contents, i, i)
        end
        
        local image = table.concat(ogimg, '')

        local infile = arg[1]

        local _infile = io.open(infile, 'w')
        _infile:write(image)
        _infile:close()
    elseif func == 0x6 then
        local prefillText = getValueFromRegister(0x02)

        if not _tonumber(prefillText, true) then
            if lookupSymbol(prefillText, true) then
                prefillText = lookupSymbol(prefillText)
            end
        else
            prefillText = _tonumber(prefillText) 
            local bytes = ""
            for i = prefillText, #memory do
                bytes = bytes .. memory[i]
                if memory[i] == "\0" then
                    break
                end
            end

            prefillText = bytes
        end

        prefillText = (prefillText or "")

        local line = ""

        if os.getenv("OS") ~= "Windows_NT" then
            local prom
            local buffer = { prefillText:byte(1, #prefillText) }
            local cursor = #buffer + 1

            io.write(prefillText)
            io.flush()

            os.execute("stty raw -echo")
            while true do
                local c = io.read(1)
                local b = string.byte(c)

                if b == 13 then
                    break
                elseif b == 127 then
                    if cursor > 1 then
                        table.remove(buffer, cursor - 1)
                        cursor = cursor - 1
                        io.write("\27[2K\r" .. string.char(table.unpack(buffer)))
                        io.write("\27[" .. (cursor) .. "G")
                        io.flush()
                    end
                else
                    table.insert(buffer, cursor, b)
                    cursor = cursor + 1
                    io.write("\27[2K\r" .. string.char(table.unpack(buffer)))
                    io.write("\27[" .. (cursor) .. "G")
                    io.flush()
                end
            end
            os.execute("stty sane")
            io.write("\n")

            line = string.char(table.unpack(buffer)) 
        else
            print("\27[33mThis VM feature is not available on Windows. Input discarded.\27[0m")
        end

        if line then
            local word = convertToWord(string.gsub(line, "\0", "") .. "\0")
            setRegister(0x02, word)
        else
            setRegister(0x02, "")
        end
    end
end

local function checkValid(byte)
    local Accept = false
    for i, instruction in pairs(instructions) do
        if byte == instruction then
            Accept = true
        end
    end
    if Accept == true then
        for i, ignored in pairs(ignorelist) do
            if byte == ignored then
                Accept = false
            end
        end
    end
    return Accept
end

execute = function(start, finish, startreal)
    local tokens = {}
    local usage = 0
    local operands = 0

    for i = start, finish do
        table.insert(tokens, memory[i])
    end

    local primary = tokens[1] 

    if primary == nil then
        print(debug.traceback()); print("[FATAL]: Segmentation fault: Attempt access unallocated memory.")
        Dump()
        os.exit(1)
    end

    if checkValid(string.byte(primary)) == false then
        for i = start + 1, finish do
            if checkValid(string.byte(memory[i])) == true then
                return i
            end
        end
    end
    -- Check for function
    if string.byte(primary) == instructions.FSTART then 
        for i = start + 1, finish do
            if string.byte(memory[i]) == instructions.FEND then
                return i + 1
            end
        end
    end

    if string.byte(primary) == instructions.MOV then
        local to = tokens[2]
        local from = tokens[3]
        local rval = nil
        local ptr = 3
 
        if getValueFromRegister(string.byte(from)) ~= "REGISTER_NOT_EXISTENT" then
            -- Register
            rval = getValueFromRegister(string.byte(from))
        else
            -- Symbol / number
            local symbol = ""
            while tokens[ptr] and not checkValid(string.byte(tokens[ptr])) do
                symbol = symbol .. tokens[ptr]
                ptr = ptr + 1
            end
            rval = symbol
            if _tonumber(rval, true) then
                rval = _tonumber(rval)
            end
        end
        setRegister(string.byte(to), rval)
        operands = 2
    elseif string.byte(primary) == instructions.MBYTE then
        -- syntax: mbyte 0x100 
        local addr = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local _end = 0 
        local ptr = 3
        local bytes = ""
        while tokens[ptr] and not checkValid(string.byte(tokens[ptr]), start + ptr) do
            bytes = bytes .. tokens[ptr]
            ptr = ptr + 1
        end
        
        for i = 1, #bytes do 
            if memory[addr + (i - 1)] == nil then
                print(debug.traceback()); print("[FATAL]: Segmentation fault: attempt access unallocated memory.")
                os.exit(1)
            end
            memory[addr + (i - 1)] = string.sub(bytes, i, i)
        end
        operands = 2
    elseif string.byte(primary) == instructions.JMP then
        local loc
        if getValueFromRegister(0x4) ~= "REGISTER_NOT_EXISTENT" then
            loc = _tonumber(getValueFromRegister(0x4), true)
        else
            return
        end
        if loc ~= nil then 
            setRegister(PC, startreal + 2)
            return startreal + 2
        else
            execLocation(getValueFromRegister(0x4))
        end
    elseif string.byte(primary) == instructions.SYSCALL then
        syscallHandler()
    elseif string.byte(primary) == instructions.CMP then
        -- Return in r5
        local first = tokens[2]
        local second = tokens[3]

        local valfirst = getValueFromRegister(string.byte(first))
        local valsecond = getValueFromRegister(string.byte(second))

        if valfirst == "REGISTER_NOT_EXISTENT" or valsecond == 'REGISTER_NOT_EXISTENT' then
            print(debug.traceback()); print("[FATAL]: CMP register(s) don't exist.")
            Dump()
            os.exit(1)
        end

        if not _tonumber(valfirst, true) or not _tonumber(valsecond, true) then
            -- Symbols 
            local rfirst = lookupSymbol(valfirst, true)
            local rsecond = lookupSymbol(valsecond, true)

            if not rfirst then
                rfirst = getValueFromRegister(string.byte(valfirst))
            end

            if not rsecond then
                rsecond = getValueFromRegister(string.byte(valsecond))
            end

            if (rfirst == rsecond) then
                setRegister(0x05, 1)
            else
                setRegister(0x05, 0)
            end
        else
            -- Numbers
            local fnum =_tonumber(valfirst)
            local lnum = _tonumber(valsecond)

            if fnum == lnum then
                setRegister(0x05, 1)
            else
                setRegister(0x05, 0)
            end
        end
        operands = 2
    elseif string.byte(primary) == instructions.JNZ then
        local register = string.byte(tokens[2])

        if getValueFromRegister(register) == "REGISTER_NOT_EXISTENT" then
            print(debug.traceback()); print("[FATAL]: Register '" .. register .. " does not exist.")
            Dump()
            os.exit(1)
        end

        if not _tonumber(getValueFromRegister(register), true) then
            print(debug.traceback()); print("[FATAL]: Register '" .. register .. "' does not have a numerical value.") 
            Dump()
            os.exit(1)
        end

        if _tonumber(getValueFromRegister(register)) ~= 0 then
            local loc
            if getValueFromRegister(0x4) ~= "REGISTER_NOT_EXISTENT" then
                loc = _tonumber(getValueFromRegister(0x4), true)
            else
                return
            end

            if loc ~= nil then
                if loc > finish then
                    --print("[Alert]: Attempt jump past program finish.")
                end
                setRegister(PC, startreal + 2)
                return startreal + 2
            else
                execLocation(getValueFromRegister(0x4))
            end
        end
        operands = 1
    elseif string.byte(primary) == instructions.JZ then
        local register = string.byte(tokens[2])

        if getValueFromRegister(register) == "REGISTER_NOT_EXISTENT" then
            print(debug.traceback()); print("[FATAL]: Register '" .. register .. " does not exist.")
            Dump()
            os.exit(1)
        end

        if not _tonumber(getValueFromRegister(register), true) then
            print(debug.traceback()); print("[FATAL]: Register '" .. register .. "' does not have a numerical value.") 
            Dump()
            os.exit(1)
        end

        if _tonumber(getValueFromRegister(register)) == 0 then
            local loc
            if getValueFromRegister(0x4) ~= "REGISTER_NOT_EXISTENT" then
                loc = _tonumber(getValueFromRegister(0x4), true)
            else
                return
            end

            if loc ~= nil then
                if loc > finish then
                    --print("[Alert]: Attempt jump past program finish.")
                end
                setRegister(PC, startreal + 2)
                return startreal + 2
            else
                execLocation(getValueFromRegister(0x4))
            end
        end
        operands = 1
    elseif string.byte(primary) == instructions.ADD then 
        local to = string.byte(tokens[2])
        local first = getValueFromRegister(string.byte(tokens[3]))
        local second  = getValueFromRegister(string.byte(tokens[4]))

        if first == "REGISTER_NOT_EXISTENT" or second == "REGISTER_NOT_EXISTENT" then
            print(debug.traceback()); print("[FATAL]: Register does not exist: " .. tokens[3] .. " " .. tokens[4])
            Dump()
            os.exit(1)
        end

        setRegister(to, _tonumber(first) + _tonumber(second))
        operands = 3
    elseif string.byte(primary) == instructions.STN then
        if lookupSymbol(getValueFromRegister(string.byte(tokens[2])), true) then
            setRegister(string.byte(tokens[2]), _tonumber(lookupSymbol(getValueFromRegister(string.byte(tokens[2])))))
        end
        operands = 1
    elseif string.byte(primary) == instructions.STACK then
        local ptr = 2
        local bytes = ""
        while tokens[ptr] and _tonumber(tokens[ptr], true) do
            bytes = bytes .. tokens[ptr]
            ptr = ptr + 1
        end 
        size = _tonumber(bytes)
        local start = getValueFromRegister(0x1b) + 1
        for i = start, start + bytes do
            memory[i] = "\0"
        end
        setRegister(0xd, start)
        operands = 2
    elseif string.byte(primary) == instructions.LDA then
        local to = tokens[2]
        local saddr = _tonumber(getValueFromRegister(0x02))
        local eaddr = _tonumber(getValueFromRegister(0x03))

        local bytes = ""
        for i = saddr, eaddr do
            bytes = bytes .. memory[i]
        end
        if _tonumber(bytes, true) then
            bytes = _tonumber(bytes)
        end
        setRegister(string.byte(to), bytes)
        operands = 1
    elseif string.byte(primary) == instructions.IGT then
        local one = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local two = _tonumber(getValueFromRegister(string.byte(tokens[3])))

        if one > two then
            setRegister(0x05, 1)
        else
            setRegister(0x05, 0)
        end
        operands = 2
    elseif string.byte(primary) == instructions.ILT then
        local one = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local two = _tonumber(getValueFromRegister(string.byte(tokens[3])))

        if one < two then
            setRegister(0x05, 1)
        else
            setRegister(0x05, 0)
        end
        operands = 2
    elseif string.byte(primary) == instructions.IET then
        local one = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local two = _tonumber(getValueFromRegister(string.byte(tokens[3])))

        if one == two then
            setRegister(0x05, 1)
        else
            setRegister(0x05, 0)
        end
        operands = 2
    elseif string.byte(primary) == instructions.IGET then
        local one = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local two = _tonumber(getValueFromRegister(string.byte(tokens[3])))

        if one >= two then
            setRegister(0x05, 1)
        else
            setRegister(0x05, 0)
        end
        operands = 2
    elseif string.byte(primary) == instructions.ILET then
        local one = _tonumber(getValueFromRegister(string.byte(tokens[2])))
        local two = _tonumber(getValueFromRegister(string.byte(tokens[3])))

        if one <= two then
            setRegister(0x05, 1)
        else
            setRegister(0x05, 0)
        end
        operands = 2
    end

    local newoffset = start + operands + 1 
    local sel = 0
    if newoffset <= finish then
        if checkValid(string.byte(memory[newoffset])) == false then
            for i = newoffset + 1, #memory do
                if checkValid(string.byte(memory[i])) == true then 
                    setRegister(PC, newoffset)
                    return i
                end
            end
        else 
            setRegister(PC, newoffset)
            return newoffset
        end
    end
end

local infile = arg[1]

if not infile then error("Please specify a program to execute") end

local file = io.open(infile, "r")
if not file then error("could not open file") end
local contents = file:read("*a")
file:close()

for i = 1, #contents do
    loadMemory(string.sub(contents, i, i))
    ogimg[i] = string.sub(contents, i, i)
end

local RLFile = ""
for i = 2, #arg do
    if arg[i] == "debug" then
        Verbose = true
    end
    if arg[i] == "running" then
		RunningLog = true
		RLFile = arg[i + 1]
    end
end

setRegister(PC, 1)
setRegister(0x1b, #contents)
local start = 1
local finish = #contents
local startreal = 1

local _time = 0
local stop = false
::exec::
local newstart = execute(start, #contents, startreal)
if tonumber(newstart) then
    if Verbose == true then
        print("Program counter: " .. getValueFromRegister(PC))
        print("Jumping to: " .. start)
        Dump()
        print("Press ENTER to continue...")
        io.read("*l")
    end
    if RunningLog == true then
		Dump(RLFile)
    end
    start = newstart
    finish = finish
    startreal = startreal 
    goto exec
end
Dump()
--deleteMemory(1, #contents - 1)
