local registers = {
    {0x0, "R0"},
    {0x1, "R1"},
    {0x2, "R2"},
    {0x3, "R3"},
    {0x4, "R4"},
    {0x5, "R5"},
    {0x6, "R6"},
    {0x7, "R7"},
    {0x8, "R8"},
    {0x9, "R9"},
    {0xa, "R10"},
    {0xb, "R11"},
    {0xc, "R12"}, 
    {0xd, "SP"},
    {0xe, "PC"},
    {0xf, "T1"},
    {0x10, "T2"},
    {0x11, "T3"},
    {0x12, "T4"},
    {0x13, "T5"},
    {0x14, "T6"},
    {0x15, "T7"},
    {0x16, "T8"},
    {0x17, "T9"},
    {0x18, "T10"},
    {0x19, "T11"},
    {0x1a, "T12"},
    {0x1b, "PTR"}
}

local errors = {
    [1] = "Invalid register",
    [2] = "Invalid instruction",
    [3] = "String not closed by quotes",
    [4] = "String not terminated by null char",
    [5] = "Comment is not closed",
    [6] = "Function is not closed",
    [7] = "Invalid preprocessor argument",
    [8] = "Invalid number",
    [9] = "[ expected"
}

local outbuffer = ""

local function throwNew(typeo, err, args)
	local etext = errors[err]
	if typeo == "warning" then
		print("Warning: " .. etext .. " " .. args .. " (Warning " .. tostring(err) .. ")")
	elseif typeo == "error" then
		print("Error: " .. etext .. " " .. args .. " (Error " .. tostring(err) .. ")")
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
        throwNew("error", 1, "'" .. rname .. "'")
    end
end

local function writeToBufRaw(text)
    outbuffer = outbuffer .. text
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
    outbuffer = outbuffer .. string.char(text)
end

local function ToTable(text)
    local chars = {}
    for i = 1, #text do
        table.insert(chars, string.sub(text, i, i)) 
    end
    return chars
end

--[[
Instruction binary:
MOV = 0x01
LDI = 0x02
JMP = 0x03
JNZ = 0x04
CALL = 0x05
]]

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
    MBYTE = 0x1b;
    BSTART = 0x1c;
    BEND = 0x1d;
    STN = 0x1e;
}

local function getInstructionFromName(ins)
    ins = string.upper(ins)
    if instructions[ins] then
        return instructions[ins]
    else
        throwNew("error", 2, "'" .. ins .. "'")
    end
end

local compile

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
                        throwNew("error", 3, "'" .. tokens[#tokens] .. "'")
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

compile = function(text, args)
    if not text then return end
    local after = 0
    local tokens = {}

    for token in string.gmatch(text, "%S+") do
        table.insert(tokens, token)
    end

    if tokens[1] == nil then
        return
    end

    if string.upper(tokens[1]) == "MOV" then
        local to = tokens[2]
        to = removeComma(to)
        local from = tokens[3]
        from = parse(from)
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(to))
        writeToBuf(from)
        after = 4
    elseif string.upper(tokens[1]) == ";" then
        local _end = 0
        for i = 2, #tokens do
            if tokens[i] == ";" then
                _end = i
                break
            end
        end

        if _end == 0 then
            throwNew("error", 5, "")
        end

        after = _end + 1
    elseif string.upper(tokens[1]) == ";-" then
        if tokens[2] == "include" then
            local filename = tokens[3]
            local file = io.open(filename, 'r')
            if not file then
                error("File not found '" .. filename .. "'")
            end
            local contents = file:read("a")
            file:close()
            compile(contents, { filename = filename })
            after = 4
        else
            throwNew("error", 7, tokens[2])
        end
    elseif string.upper(tokens[1]) == "JMP" then
        writeToBuf(getInstructionFromName(tokens[1]))
        after = 2
    elseif string.upper(tokens[1]) == "JNZ" then
        local register = getRegisterFromName(tokens[2])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(register)
        after = 3
    elseif string.upper(tokens[1]) == "JZ" then
        local register = getRegisterFromName(tokens[2])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(register)
        after = 3
    elseif string.upper(tokens[1]) == "SYSCALL" then
        writeToBuf(getInstructionFromName(tokens[1]))
        after = 2
    elseif string.upper(tokens[1]) == "ADD" then

        local to = tokens[2]
        local first = tokens[3]
        to = removeComma(to)
        first = removeComma(first)
        
        to = getRegisterFromName(to)
        first = getRegisterFromName(first)
        local second = getRegisterFromName(tokens[4])

        writeToBuf(instructions.ADD)
        writeToBuf(to)
        writeToBuf(first)
        writeToBuf(second)

        after = 5
    elseif string.upper(tokens[1]) == "SUB" then
        local to = tokens[2]
        to = removeComma(to)
        local first = parse(tokens[3])
        local second = parse(tokens[4])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(to))
        writeToBuf(first)
        writeToBuf(second)
        after = 5
    elseif string.upper(tokens[1]) == "MUL" then
        local to = tokens[2]
        to = removeComma(to)
        local first = parse(tokens[3])
        local second = parse(tokens[4])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(to))
        writeToBuf(first)
        writeToBuf(second)
        after = 5
    elseif string.upper(tokens[1]) == "DIV" then
        local to = tokens[2]
        to = removeComma(to)
        local first = parse(tokens[3])
        local second = parse(tokens[4])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(getRegisterFromName(to))
        writeToBuf(first)
        writeToBuf(second)
        after = 5
    elseif string.upper(tokens[1]) == "DB" then
        local ending = 0
        local tokensToParse = {}
        for i = 2, #tokens do
            if string.find(tokens[i], [[\0]]) then
                ending = i
                tokens[i] = string.gsub(tokens[i], [[\0]], string.char(0x0))
                break
            end
        end
        for i = 2, ending do
            if string.find(tokens[i], [[\n]]) then
                tokens[i] = string.gsub(tokens[i], [[\n]], "\n")
            end
        end
        if ending == 0 then
            throwNew("error", 4, "'" .. tokens[#tokens] .. "'")
        end
        for i = 2, ending do
            table.insert(tokensToParse, tokens[i])
        end
        local parsed = parse(nil, 1, tokensToParse)
        return parsed
    elseif string.upper(tokens[1]) == "DW" then
        if tonumber(tokens[2]) then
            return string.char(instructions.DW) .. tokens[2]
        end
    elseif string.upper(tokens[1]) == "NOP" then
        writeToBuf(instructions.NOP)
        after = 2
    elseif string.upper(tokens[1]) == "CMP" then
        local first = tokens[2]
        local second = tokens[3]
        first = removeComma(first)
        writeToBuf(instructions.CMP)
        writeToBuf(getRegisterFromName(first))
        writeToBuf(getRegisterFromName(second))
        after = 4
    elseif string.upper(tokens[1]) == "STN" then
        writeToBuf(instructions.STN)
        writeToBuf(getRegisterFromName(tokens[2]))
        after = 3
    elseif string.upper(tokens[1]) == "MBYTE" then
        -- Syntax: mbyte 3000 [ hello world! ]
        writeToBuf(instructions.MBYTE) 
        tokens[2] = removeComma(tokens[2])
        writeToBuf(getRegisterFromName(tokens[2]))
        local _end = 0
        if tokens[3] ~= "[" then
            throwNew("error", 9)
        end
        local str = ""

        local vtokens = {}
        for i = 4, #tokens do
            if tokens[i] == "]" then
                _end = i
                break
            else
                table.insert(vtokens, tokens[i])
            end
        end
        str = table.concat(vtokens, ' ')

        if _end == 0 then
            throwNew("error", 10)
        end

        writeToBufRaw(str)
        after = _end + 1
    elseif string.find(tokens[1], ":") and string.find(tokens[1], ":") == #tokens[1] then
        tokens[1] = string.gsub(tokens[1], ":", "")
        if string.find(tokens[1], ":") then
            return
        end
        local ending = 0
        for i = 2, #tokens do
            if tokens[i] == ":" .. tokens[1] then
                ending = i
                break
            end
        end
        if ending == 0 then
            return
        end
        local toParse = ""
        for i = 2, ending - 1 do
            if i == 2 then
                toParse = toParse .. tokens[i]
            else
                toParse = toParse .. " " .. tokens[i]
            end
        end
        local varname = tokens[1]
        local value = compile(toParse)
        writeToBuf(instructions.DSTART)
        writeToBufRaw(varname)
        writeToBuf(instructions.DSEP)
        writeToBufRaw(value)
        writeToBuf(instructions.DEND)

        after = ending + 1
        if not tokens[after] then
            after = 0
        end
    elseif ToTable(tokens[1])[1] == ":" and ToTable(tokens[1])[#tokens[1]] then
        local _end = 0
        local name = ""
        for i = 2, #tokens do
            if tokens[i] == tokens[1] then
                _end = i
                break
            end
        end
        
        if _end == 0 then
            throwNew("error", 6, "")
        end

        local ntoken = ToTable(tokens[1])

        for i = 2, #ntoken - 1 do
            name = name .. ntoken[i]
        end

        local ftokens = {}

        for i = 2, _end - 1 do
            table.insert(ftokens, tokens[i])
        end

        writeToBuf(instructions.FSTART)
        writeToBufRaw(name)
        writeToBuf(instructions.FSEP)
        compile(table.concat(ftokens, " "), {})
        writeToBuf(instructions.FEND)

        after = _end + 1
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

local infile = arg[1]
local outfile = arg[2]

local file = io.open(infile, "r")
if not file then error("File not found") end
local content = file:read("*a")
file:close()
compile(content, { filename = infile })
print("Program compiled to: " .. outbuffer)

local file = io.open(outfile, "w")
if not file then error("Could not create file") end
file:write(outbuffer)
file:close()
