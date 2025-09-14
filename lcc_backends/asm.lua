local Bindings = {}
local UnresolvedBindings = {}
local UBID = 1

local CurrentSection = "text"
local DataBuffer = ""
local CodeBuffer = ""
local ExtendedDataBuffer = ""
local CurrentFileName = ""
local compile

function table.find(haystack, needle)
    for Index, Value in pairs(haystack) do
        if Value == needle then
            return Index
        end
    end
end

local registers = {
    {0x0000, "R0", 0},
	{0x0001, "R1", 0},
	{0x0002, "R2", 0},
	{0x0003, "R3", 0},
	{0x0004, "R4", 0},
	{0x0005, "R5", 0},
	{0x0006, "R6", 0},
	{0x0007, "R7", 0},
	{0x0008, "R8", 0},
	{0x0009, "R9", 0},
	{0x000a, "R10", 0},
	{0x000b, "R11", 0},
	{0x000c, "R12", 0},
	{0x000d, "T1", 0},
	{0x000e, "T2", 0},
	{0x000f, "T3", 0},
	{0x0010, "T4", 0},
	{0x0011, "T5", 0},
	{0x0012, "T6", 0},
	{0x0013, "T7", 0},
	{0x0014, "T8", 0},
	{0x0015, "T9", 0},
	{0x0016, "T10", 0},
	{0x0017, "T11", 0},
	{0x0018, "T12", 0},
	{0x0019, "SP", 0},
	{0x001a, "PC", 0},
	{0x001b, "RE1", 0},
	{0x001c, "RE2", 0},
	{0x001d, "RE3", 0},
}

local errors = {
    [1] = "assembler: invalid register",
    [2] = "assembler: invalid instruction",
    [3] = "assembler: string not wrapped by quotes",
    [4] = "assembler: string not terminated by null char",
    [5] = "assembler: comment is not closed",
    [6] = "assembler: function is not closed",
    [7] = "assembler: invalid preprocessor argument",
    [8] = "assembler: invalid number",
    [9] = "assembler: '[' expected",
    [10] = "assembler: output machine code too big",
    [11] = "assembler: string too long for register (max 2 bytes)",
    [12] = "assembler: putting more than one character to an immediate may have undesirable results",
    [13] = "assembler: unknown section label",
    [14] = "linker: undefined symbol",
    [15] = "no input files",
    [16] = "assembler: no such file or directory",
    [17] = "linker: entry point not found",
    [18] = "assembler: unable to create file"
}

local instructions = { 
    MOV = 0x01,
    HLT = 0x02,
    JMP = 0x03,
    INT = 0x04,
    JNZ = 0x05,
    NOP = 0x06,
    CMP = 0x07,
    JZ = 0x08,
    INC = 0x09,
    DEC = 0x0a,
    PUSH = 0x0b,
    POP = 0x0c,
    ADD = 0x0d,
    SUB = 0x0e,
    MUL = 0x0f,
    DIV = 0x10,
    IGT = 0x11,
    ILT = 0x12,
    AND = 0x13,
    OR = 0x14,
    NOR = 0x15,
    NOT = 0x16,
    XOR = 0x17,
    LOD = 0x18,
    STR = 0x19,
}

local function throw(typeo, err, args)
	local etext = errors[err]
    local filename = CurrentFileName
    if filename ~= "" then
        filename = filename .. ": "
    else
        filename = "lcc: "
    end
	if typeo == "warning" then
        print("\27[1;37m" .. filename .. "\27[1;33mwarning:\27[1;37m " .. (etext or "") .. " " .. (args or "") .. "\27[0m")
	elseif typeo == "error" then
        print("\27[1;37m" .. filename .. "\27[1;31merror:\27[1;37m " .. (etext or "") .. " " .. (args or "") .. "\27[0m")
        if table.find(arg, "--error-extra-info") then
            print(debug.traceback())
        end
		os.exit(err)
	end
end

local function getRegisterFromName(rname, silent)
    for _, register in pairs(registers) do
        if register[2] == string.upper(rname) then
            return register[1]
        end
    end
    if silent ~= true then
        throw("error", 1, "'" .. rname .. "'")
    end
end

local TrackWrites = false
local WriteAdd = 0
local function writeToBufRaw(text)
    if TrackWrites == true then
        WriteAdd = WriteAdd + string.len(text)
    end
    if CurrentSection == "text" then
        CodeBuffer = CodeBuffer .. text
    elseif CurrentSection == "data" then
        DataBuffer = DataBuffer .. text
    elseif CurrentSection == "edata" then
        ExtendedDataBuffer = ExtendedDataBuffer .. text
    end
end

local function writeToBuf(text)
    if type(text) == "table" then
        if text[2] == "NUMBER" then
            writeToBufRaw(tostring(text[1]))
            return
        elseif text[2] == "REGISTER" then
            text = text[1]
        elseif text[2] == "SYMBOL" then
            writeToBufRaw(text[1])
            return
        end
    end
    if CurrentSection == "text" then
        writeToBufRaw(string.char(text))
    else
        writeToBufRaw(string.char(text))
    end
end

local function ToTable(text)
    local chars = {}
    for i = 1, #text do
        table.insert(chars, string.sub(text, i, i)) 
    end
    return chars
end

local function getInstructionFromName(ins)
    ins = string.upper(ins)
    if instructions[ins] then
        return instructions[ins]
    else
        throw("error", 2, "'" .. ins .. "'")
    end
end

local function parse(token, start, tokens)
    if token ~= nil then
        if string.find(token, ",") then
            token = string.gsub(token, ",", "")
        end
    end
    if tokens ~= nil then
        if string.find(tokens[start], '"') then
            local index = string.find(tokens[start], '"')
            if index == 1 then
                local ending = 0
                for i = start + 1, #tokens do
                    if string.find(tokens[i], '"') and string.find(tokens[i], '"') == #tokens[i] then
                        ending = i
                        break
                    end
                end
                local onetoken = false
                if ending == 0 then
                    if #tokens == 1 then
                        onetoken = true
                    else
                        throw("error", 3, "'" .. tokens[#tokens] .. "'")
                    end
                end
                tokens[start] = string.gsub(tokens[start], '"', "")
                if not onetoken then
                    tokens[ending] = string.gsub(tokens[ending], '"', "")
                    local str = ""
                    for i = start, ending do 
                        if i == start then
                            str = str .. tokens[i]
                        else
                            str = str .. " " .. tokens[i]
                        end
                    end
                    return str
                else 
                    return tokens[start]
                end
            end
        elseif string.find(string.lower(tokens[start]), "db") then
            local range = start + 1
            local str = "db"
            for i = range, #tokens do
                str = str .. " " .. tokens[i]
            end
            local result = compile(str)
            print(result)
        end
    end
    if tonumber(token) then
        return {token, "NUMBER"}
    end
    if getRegisterFromName(token, true) then
        return {getRegisterFromName(token), "REGISTER"}
    end

    return {token, "SYMBOL"}
end

local function removeComma(str)
    return string.gsub(str, ",", "")
end

local function UInt16(n)
    n = n & 0xFFFF
    local hi = (n >> 8) & 0xFF
    local lo = n & 0xFF
    return string.char(hi), string.char(lo)
end

compile = function(text)
    if not text then return end
    local after = 0
    local tokens = {}

    for token in string.gmatch(text, "%S+") do
        table.insert(tokens, token)
    end

    if tokens[1] == nil then
        return
    end

    if string.upper(tokens[1]) == "MOV" or string.upper(tokens[1]) == "JNZ" or string.upper(tokens[1]) == "JZ" or string.upper(tokens[1]) == "JMP" or string.upper(tokens[1]) == "PUSH" then
        local Noto = false
        if string.upper(tokens[1]) == "JMP" or string.upper(tokens[1]) == "PUSH" then
            Noto = true
        end

        local to = tokens[2]
        to = removeComma(to)
        local mode 

        if Noto == true then
            from = parse(tokens[2])
        else
            from = parse(tokens[3])
        end

        from[1] = string.gsub(from[1], "\\n", "\n")
        from[1] = string.gsub(from[1], "\\0", "\0")
        from[1] = string.gsub(from[1], "\\r", "\r")
        from[1] = string.gsub(from[1], "\\27", "\27")
        from[1] = string.gsub(from[1], "\\20", " ")

        local Mode = nil
        local DNR = false 
 
        local FH
        local FL
 
        if from[2] == "REGISTER" then
            Mode = 2 
        elseif from[2] == "NUMBER" then
            Mode = 1
            FH, FL = UInt16(tonumber(from[1])) 
        elseif from[2] == "SYMBOL" then
            Mode = 1 
            if string.sub(from[1], 1, 1) ~= "\"" or string.sub(from[1], #from[1], #from[1]) ~= "\"" then
                DNR = true
                local Found = false
                writeToBuf(getInstructionFromName(tokens[1]))
                writeToBufRaw(string.char(0x01))
                if not Noto then
                    writeToBuf(getRegisterFromName(to))
                end
                for _, Binding in pairs(Bindings) do
                    local Location = Binding[2]
                    local Name = Binding[1]
                    if Name == from[1] then
                        Found = true
                        local H, L = UInt16(Location)
                        writeToBufRaw(H)
                        writeToBufRaw(L) 
                    end
                end
                if Found == false then 
                    table.insert(UnresolvedBindings, {UBID, fname})
                    writeToBufRaw("UNRESOLVED_BINDING_" .. UBID)
                    UBID = UBID + 1
                end        
            else
                from[1] = string.gsub(from[1], "\"", "")
                if string.len(from[1]) > 2 then
                    throw("error", 11)
                end
                if string.len(from[1]) == 2 then
                    throw("warning", 12)
                    local Num = "0x" 
                    Num = Num .. string.format("%02x", string.byte(string.sub(from[1], 1, 1)))
                    Num = Num .. string.format("%02x", string.byte(string.sub(from[1], 2, 2)))
                    Num = tonumber(Num)
                    FH, FL = UInt16(Num) 
                else
                    local Num = "0x00"
                    Num = Num .. string.format("%02x", string.byte(string.sub(from[1], 1, 1))) 
                    Num = tonumber(Num)
                    FH, FL = UInt16(Num)
                end 
            end 
        end

        if DNR == false then
            writeToBuf(getInstructionFromName(tokens[1]))
            writeToBufRaw(string.char(Mode)) 
            if not Noto then
                writeToBuf(getRegisterFromName(to))
            end
            if FH and FL then
                writeToBufRaw(FH)
                writeToBufRaw(FL)
            else
                writeToBuf(from)
            end
        end
        if Noto == false then
            after = 4
        else
            after = 3
        end
    elseif string.upper(tokens[1]) == "HLT" then
        writeToBuf(getInstructionFromName(tokens[1]))
        after = 2 
    elseif string.upper(tokens[1]) == "INT" then
        local number = tokens[2]
        number = tonumber(number)
        local H, L = UInt16(number)
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBufRaw(H)
        writeToBufRaw(L)
        after = 3
    elseif string.upper(tokens[1]) == "NOP" then
        writeToBuf(getInstructionFromName(tokens[1]))
        after = 2
    elseif string.upper(tokens[1]) == "INC" or string.upper(tokens[1]) == "DEC" or string.upper(tokens[1]) == "POP" then
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(tokens[2]))
        after = 3
    elseif string.upper(tokens[1]) == "ADD" or string.upper(tokens[1]) == "SUB" or string.upper(tokens[1]) == "MUL" or string.upper(tokens[1]) == "DIV" or string.upper(tokens[1]) == "IGT" or string.upper(tokens[1]) == "ILT" or string.upper(tokens[1]) == "AND" or string.upper(tokens[1]) == "OR" or string.upper(tokens[1]) == "NOR" or string.upper(tokens[1]) == "XOR" then
        tokens[2] = removeComma(tokens[2])
        tokens[3] = removeComma(tokens[3])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(tokens[2]))
        writeToBuf(getRegisterFromName(tokens[3]))
        writeToBuf(getRegisterFromName(tokens[4]))
        after = 5
    elseif string.upper(tokens[1]) == "NOT" or string.upper(tokens[1]) == "LOD"  or string.upper(tokens[1]) == "STR" then
        tokens[2] = removeComma(tokens[2])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(tokens[2]))
        writeToBuf(getRegisterFromName(tokens[3]))
        after = 4
    elseif string.upper(tokens[1]) == "PCL" then
        -- Add 2 to stack pointer to set up new stack frame
        -- Callee's job to pop everything off before or else things
        -- will go wrong.
        -- Initial 11
        writeToBuf(getInstructionFromName("MOV"))
        writeToBufRaw(string.char(0x01))
        writeToBuf(getRegisterFromName("R0"))
        writeToBufRaw("LASM_JLOC")
        compile([[ 
        add re1, pc, r0
        push re1
        ]])
        TrackWrites = true
        after = 2
    elseif string.upper(tokens[1]) == "JCL" then
        local fname = tokens[2] 
        local H, L = UInt16(11 + WriteAdd)
        DataBuffer = string.gsub(DataBuffer, "LASM_JLOC", H .. L)
        CodeBuffer = string.gsub(CodeBuffer, "LASM_JLOC", H .. L)
        ExtendedDataBuffer = string.gsub(ExtendedDataBuffer, "LASM_JLOC", H .. L)
        TrackWrites = false
        WriteAdd = 0
        writeToBuf(getInstructionFromName("JMP"))
        writeToBufRaw(string.char(0x01))
        local Found = false
        for _, Binding in pairs(Bindings) do
            local Location = Binding[2]
            local Name = Binding[1]
            if Name == fname then
                Found = true
                local H, L = UInt16(Location)
                writeToBufRaw(H)
                writeToBufRaw(L) 
            end
        end
        if Found == false then 
            table.insert(UnresolvedBindings, {UBID, fname})
            writeToBufRaw("UNRESOLVED_BINDING_" .. UBID)
            UBID = UBID + 1
        end
        after = 3
    elseif string.upper(tokens[1]) == "RET" then
        -- Subtract 2 from stack pointer to restore old stack frame
        compile([[ 
        pop re1
        ]])
        writeToBuf(getInstructionFromName("JMP"))
        writeToBufRaw(string.char(0x02))
        writeToBuf(getRegisterFromName("RE1"))
        after = 2
    elseif string.upper(tokens[1]) == "DB" then
        local ending = 0
        local tokensToParse = {}
        for i = 2, #tokens do
            tokens[i] = string.gsub(tokens[i], "\\n", "\n")
            if string.find(tokens[i], [[\0]]) then
                ending = i
                tokens[i] = string.gsub(tokens[i], [[\0]], string.char(0x0))
                break
            end
        end
        for i = 2, ending do
            tokens[i] = string.gsub(tokens[i], [[\n]], "\n")
            tokens[i] = string.gsub(tokens[i], [[\r]], "\r")
            tokens[i] = string.gsub(tokens[i], [[\27]], "\27") 
            tokens[i] = string.gsub(tokens[i], [[\033]], "\27")
            tokens[i] = string.gsub(tokens[i], [[\x1b]], "\27")
        end
        if ending == 0 then
            throw("error", 4, "'" .. tokens[#tokens] .. "'")
        end
        for i = 2, ending do
            table.insert(tokensToParse, tokens[i])
        end
        local parsed = parse(nil, 1, tokensToParse)
        writeToBufRaw(parsed)
        after = ending + 2 
    elseif string.upper(tokens[1]) == "NOP" then
        writeToBuf(instructions.NOP)
        after = 2
    elseif string.upper(tokens[1]) == "CMP" then
        local register = tokens[2]
        local first = tokens[3]
        local second = tokens[4]
        register = removeComma(register)
        first = removeComma(first)
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(register))
        writeToBuf(getRegisterFromName(first))
        writeToBuf(getRegisterFromName(second))
        after = 5
    elseif string.upper(tokens[1]) == ";" then
        local _end = 0
        for i = 2, #tokens do
            if tokens[i] == ";" then
                _end = i
                break
            end
        end

        if _end == 0 then
            throw("error", 5, "")
        end

        after = _end + 1
    elseif string.upper(tokens[1]) == "%" then
        if tokens[2] == "include" then
            local filename = tokens[3]
            local file = io.open(filename, 'r')
            if not file then
                print("\27[31mFile not found '" .. filename .. "'\27[0m")
                os.exit(1)
            end
            local contents = file:read("a")
            file:close()
            local LFN = CurrentFileName
            CurrentFileName = filename
            compile(contents)
            CurrentFileName = LFN
            after = 4
        elseif tokens[2] == "embed" then
            local filename = tokens[3]
            local file = io.open(filename, 'r')
            if not file then
                print("\27[31mFile not found '" .. filename .. "'\27[0m")
                os.exit(1)
            end
            local contents = file:read("a")
            file:close()
            writeToBufRaw(contents)
            after = 4 
        elseif tokens[2] == "section" then
            local section = tokens[3]
            if section ~= "data" and section ~= "text" and section ~= "edata" then
                throw("error", 13, section) 
            end
            CurrentSection = section
            after = 4
        else
            throw("error", 7, tokens[2])
        end
    elseif string.find(tokens[1], ":") and string.find(tokens[1], ":") == #tokens[1] then
        local ending = 0
        for i = 2, #tokens do
            if string.find(tokens[i], ":") and string.find(tokens[i], ":") == #tokens[i] then
                ending = i
                break
            end
        end

        if ending == 0 then
            ending = #tokens + 1
        end

        local toParse = ""
        for i = 2, ending - 1 do 
            if i == 2 then
                toParse = toParse .. tokens[i]
            else
                toParse = toParse .. " " .. tokens[i]
            end
        end

        local varname = string.gsub(tokens[1], ":", "")
        local location
        if CurrentSection == "data" then
            location = 2 + #DataBuffer
        elseif CurrentSection == "text" then
            location = 2 + #DataBuffer + #CodeBuffer
        elseif CurrentSection == "edata" then
            location = 2 + #DataBuffer + #CodeBuffer + #ExtendedDataBuffer
        end
        local H, L = UInt16(location)
        table.insert(Bindings, {varname, location})

        compile(toParse) 

        -- Resolve bindings
        for j, Binding in pairs(UnresolvedBindings) do
            local ID = Binding[1]
            local Name = Binding[2] 

            if Name == varname then
                DataBuffer = string.gsub(DataBuffer, "UNRESOLVED_BINDING_" .. ID, H .. L)
                CodeBuffer = string.gsub(CodeBuffer, "UNRESOLVED_BINDING_" .. ID, H .. L)
                ExtendedDataBuffer = string.gsub(ExtendedDataBuffer, "UNRESOLVED_BINDING_" .. ID, H .. L)
                table.remove(UnresolvedBindings, j)
            end
        end

        after = ending
        if not tokens[after] then
            after = 0
        end
    end

    if after > 0 and tokens[after] ~= nil then
        local str = ""
        for i = after, #tokens do
            if i == after then
                str = str .. tokens[i]
            else
                str = str .. " " .. tokens[i]
            end
        end
        
        compile(str)
    end
end

local infiles = {}
local outfile = nil

-- Parse flags

local Operands = 0
local i = 1
while i <= #arg do
    local argument = arg[i]

    if argument == "-o" then
        outfile = arg[i + 1]
        i = i + 2
    elseif argument == "--error-extra-info" then
        i = i + 1
    else
        table.insert(infiles, argument)
        i = i + 1
    end
end

if #infiles < 1 then
    throw("error", 15)
end

for i = 1, #infiles do
    local infile = infiles[i]
    local file = io.open(infile, "r")
    if not file then throw("error", 16, "'" .. infile .. "'") end
    local content = file:read("*a")
    file:close()
    CurrentFileName = infile
    compile(content)
end

local entry = nil
local Found = false
for _, Binding in pairs(Bindings) do
    local Location = Binding[2]
    local Name = Binding[1]
    if Name == "_start" then
        Found = true
        entry = Location 
    end
end
if Found == false then 
    throw("error", 17) 
end

local H, L = UInt16(entry)
local buffer = H .. L .. DataBuffer .. CodeBuffer .. "\0" .. ExtendedDataBuffer

if #UnresolvedBindings > 0 then
    for i = 1, #UnresolvedBindings do
        local Binding = UnresolvedBindings[i]
        throw("error", 14, "'" .. Binding[2] .. "'")
    end
end

if outfile == nil then
    outfile = "a.bin"
end

local file = io.open(outfile, "w")
if not file then throw("error", 18, "'" .. outfile .. "'") end
file:write(buffer)
file:close()

