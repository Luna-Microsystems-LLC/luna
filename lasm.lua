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
    {0x19, "T10"},
    {0x1a, "T11"},
    {0x1b, "T12"},
    {0x1c, "PTR"}
}

local errors = {
    [1] = "Invalid register",
    [2] = "Invalid instruction",
    [3] = "String not wrapped by quotes",
    [4] = "String not terminated by null char",
    [5] = "Comment is not closed",
    [6] = "Function is not closed",
    [7] = "Invalid preprocessor argument",
    [8] = "Invalid number",
    [9] = "[ expected",
    [10] = "Output machine code too big",
    [11] = "String too long for register (max 2 bytes)",
    [12] = "Putting more than one character to an immediate may have undesirable results",
    [13] = "Unknown section label"
}

local LabelLocations = {}

local CurrentSection = "text"
local DataBuffer = ""
local CodeBuffer = ""

local function throw(typeo, err, args)
	local etext = errors[err]
	if typeo == "warning" then
        print("\27[33mWarning " .. (tostring(err) or "") .. ": " .. (etext or "") .. " " .. (args or "") .. "\27[0m")
	elseif typeo == "error" then
        print("\27[31mError " .. (tostring(err) or "") .. ": " .. (etext or "") .. " " .. (args or "") .. "\27[0m")
        print(debug.traceback())
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

local function writeToBufRaw(text)
    if CurrentSection == "text" then
        CodeBuffer = CodeBuffer .. text
    elseif CurrentSection == "data" then
        DataBuffer = DataBuffer .. text
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
        CodeBuffer = CodeBuffer .. string.char(text)
    else
        DataBuffer = DataBuffer .. string.char(text)
    end
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
    MOV = 0x01,
    HLT = 0x02,
    JMP = 0x03,
    INT = 0x04,
    JNZ = 0x05,
    NOP = 0x06,

}

local padto = 0

local function getInstructionFromName(ins)
    ins = string.upper(ins)
    if instructions[ins] then
        return instructions[ins]
    else
        throw("error", 2, "'" .. ins .. "'")
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
        local mode
        from = parse(from)

        local Mode = nil
 
        local FH
        local FL

        if from[2] == "REGISTER" then
            Mode = 2 
        elseif from[2] == "NUMBER" then
            Mode = 1
            FH, FL = UInt16(tonumber(from[1]))
            print(string.byte(FH))
            print(string.byte(FL))
        elseif from[2] == "SYMBOL" then
            Mode = 1
            if not string.sub(from[1], 1, 1) == "\"" or not string.sub(from[1], #from[1], #from[1]) == "\"" then
                throw("error", 3) 
            end
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
 
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBufRaw(string.char(Mode)) 
        writeToBuf(getRegisterFromName(to))
        if FH and FL then
            writeToBufRaw(FH)
            writeToBufRaw(FL)
        else
            writeToBuf(from)
        end
        after = 4 
    elseif string.upper(tokens[1]) == "HLT" then
        writeToBuf(getInstructionFromName(tokens[1]))
        after = 2
    elseif string.upper(tokens[1]) == "JMP" then
        local register = getRegisterFromName(tokens[2])
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBuf(register)
        after = 3
    elseif string.upper(tokens[1]) == "INT" then
        local number = tokens[2]
        number = tonumber(number)
        local H, L = UInt16(number)
        writeToBuf(getInstructionFromName(tokens[1]))
        writeToBufRaw(H)
        writeToBufRaw(L)
        after = 3
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
        end
        if ending == 0 then
            throw("error", 4, "'" .. tokens[#tokens] .. "'")
        end
        for i = 2, ending do
            table.insert(tokensToParse, tokens[i])
        end
        local parsed = parse(nil, 1, tokensToParse)
        return parsed 
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
    elseif string.upper(tokens[1]) == "IGT" or string.upper(tokens[1]) == "ILT" or string.upper(tokens[1]) == "IET" or string.upper(tokens[1]) == "IGET" or string.upper(tokens[1]) == "ILET" then
        writeToBuf(getInstructionFromName(tokens[1]))
        local one = tokens[2]
        local two = tokens[3]
        one = removeComma(one)

        writeToBuf(getRegisterFromName(one))
        writeToBuf(getRegisterFromName(two))
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
            throw("error", 5, "")
        end

        after = _end + 1
    elseif string.upper(tokens[1]) == "%" then
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
        elseif tokens[2] == "size" then
            padto = (tonumber(tokens[3]) or throw("error", 8))
            after = 4
        elseif tokens[2] == "section" then
            local section = tokens[3]
            if section ~= "data" and section ~= "text" then
                throw("error", 12, section) 
            end
            CurrentSection = section
            after = 4
        else
            throw("error", 7, tokens[2])
        end
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
            throw("error", 6, "")
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

if padto ~= 0 then
    if string.len(CodeBuffer) > padto then
        throw("error", 10) 
    elseif string.len(CodeBuffer) < padto then
        for i = string.len(CodeBuffer), padto do
            CodeBuffer = CodeBuffer .. "\0"
        end
    end
end

local buffer = DataBuffer .. CodeBuffer
print("Data:", #DataBuffer)
local entry = #DataBuffer + 2
local H, L = UInt16(entry)
buffer = H .. L .. DataBuffer .. CodeBuffer

local file = io.open(outfile, "w")
if not file then error("Could not create file") end
file:write(buffer)
file:close()

