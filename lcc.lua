#!/usr/bin/env lua
--[[
-- Calling convention:
-- T1 - T6 = function arguments
-- T7 - T8 = if statement things
-- T9 - T12 = other
--]]
function table.find(haystack, needle)
    for Index, Value in pairs(haystack) do
        if Value == needle then
            return Index
        end
    end
end
local compile
local buffer = ""
local tempcounter = 1

local defs = ""
local funcs = ""
local code = "" -- To keep things organized

local keywords = {
    ["if"] = true,
    ["else"] = true,
    ["void"] = true,
    ["int"] = true,
    ["char*"] = true,
    ["define"] = true,
    ["include"] = true,
    ["pragma"] = true,
    ["return"] = true,
}

local vtype = {
    ["void"] = true,
    ["int"] = true,
    ["char*"] = true,
}

local operator_s = {
    ["="] = true,
    ["+"] = true,
    ["-"] = true,
    ["*"] = true,
    ["/"] = true, 
}

local operator_d = {
    ["=="] = true, 
    ["!="] = true
}

local symbols = {
    ["("] = true,
    [")"] = true,
    ["{"] = true,
    ["}"] = true,
    ["["] = true,
    ["]"] = true,
    [";"] = true,
    [","] = true,
    ['"'] = true,
    ["'"] = true,
}

local errors = {
    [1] = "Unexpected character",
    [2] = "No main function found",
    [3] = "Expected function declaration",
    [4] = "'(' expected",
    [5] = "')' expected",
    [6] = "'{' expected",
    [7] = "'}' expected",
    [8] = "Variable or function declaration expected",
    [9] = "';' expected",
    [10] = "Closing '\" expected",
    [11] = "'main' function required to return to integer",
    [12] = "Unnecessary arguments to main fucntion",
    [13] = "Identifier expected",
    [14] = "',' expected",
    [15] = "Unexpected ','",
    [16] = "Unexpected '*/'",
    [17] = "Integer representation too large",
    [18] = "Unknown file type to include",
    [19] = "Operator expected",
    [20] = "Too many arguments passed to function (max 6)",
    [21] = "Invalid operator to 'if' statement",
    [22] = "Unclosed pair",
    [23] = "Functions can only be declared in the global scope",
    [24] = "Action can only be used inside of a function",
    [25] = "Comparing more than one statement is not supported",
    [26] = "Cannot edit readonly variable",
    [27] = "Number expected",
    [28] = "Unknown pragma directive",
    [29] = "No stack memory allocated to program",
    [30] = "Stack exceeds stack size",
    [31] = "Undefined variable",
    [32] = "Consider adding an explicit return value to 'main' function"
}

local reserved_varnames = {
    "__setbytes__",
    "__setregister__",
    "__jmploc__", 
    "__stack_size__",
    "__write__"
}

local STACK_SIZE = 0
local STACK_OFFSET = 0

local function write(text, prepend)
    if prepend ~= true then
        buffer = buffer .. text .. "\n"
    else
        buffer = text .. "\n" .. buffer
    end
end

local variables = {
    -- we will store it like <identifier><value>
    -- in the table it will be stored like
    -- {
    --    identifier = 1
    --    value 
    -- }
    -- ids: int (1)
}

local function tohex16(n)
    return string.format("%04X", n & 0xFFFF)
end

local function throw(_error, eargs, _type) 
    if _type == "error" or _type == nil then
        print("\27[31mError " .. tostring((_error or "(no error number)")) .. ": " .. (errors[_error] or "(no error text)") .. " " .. (eargs or "") .. "\27[0m")
        print("Buffer dump: " .. buffer)
        print(debug.traceback())
        os.exit(tonumber(_error) or 1)
    elseif _type == "warning" then
        print("\27[33mWarning " .. tostring((_error or "(no warning number)")) .. ": " .. (errors[_error] or "(no error text)") .. " " .. (eargs or "") .. "\27[0m")
    elseif _type == "info" then
        print("\27[34mInfo " .. tostring((_error or "(no info number)")) .. ": " .. (errors[_error] or "(no info text)") .. " " .. (eargs or "") .. "\27[0m") 
    end
end

local cstrings = {}

local function tokenize(text)
    local tokens = {}
    local i = 1
    local quotes = false
    local comment = false
    local current_str = ""

    local function peek(n)
        return string.sub(text, i, i + (n or 0))
    end

    local function advance(n)
        i = i + (n or 1)
    end

    while i <= #text do
        local c = peek()

        if string.match(c, "%s") then
            advance()
        elseif string.match(c, "[%a_@*#]") then
            local start = i
            local matched = false
            while string.match(peek(), "[%w_*@]") do
                advance()
                matched = true
            end
            if matched then
                local word = string.sub(text, start, i - 1)
                table.insert(tokens, { type = keywords[word] and "keyword" or "identifier", value = word })
            else
                advance()
            end
        elseif string.match(c, "%d") then
            local start = i
            while string.match(peek(), "%d") do
                advance()
            end
            table.insert(tokens, { type = "number", value = string.sub(text, start, i - 1) })
        elseif operator_d[peek(1)] then
            table.insert(tokens, { type = "operator", value = peek(1) })
            advance(2)
        elseif operator_s[c] then
            table.insert(tokens, { type = "operator", value = c })
            advance()
        elseif symbols[c] then
            if c == "\"" then
                if quotes == false then
                    quotes = true 
                else
                    quotes = false 
                end
            end

            table.insert(tokens, { type = "symbol", value = c })
            advance()
        elseif string.match(c, "//") then
            while peek() ~= "\n" do advance() end
            advance()
        else
            if quotes == false then
                throw(1, "'" .. c .. "'")
            else
                table.insert(tokens, { type = "strval", value = string.char(0) .. c })
                advance()
            end
        end

        ::continue::
    end

    return tokens
end

local function parseDef(tokens)
    -- assume structure is like ", Hello, world, "
    local function rebuild(Table)
        local str = ""
        for i = 1, #Table do
            if i == 1 then
                str = str .. Table[i]
            else
                local found = string.find(Table[i], string.char(0))
                if found ~= 1 then
                    str = str .. " " .. Table[i]
                else
                    str = str .. string.sub(Table[i], 2, #Table[i])
                end
            end
        end
        return str
    end
    if tokens[1].type == "symbol" and tokens[1].value == "\"" then
        local _end = 0
        for i = 2, #tokens do
            if tokens[i].value == "\"" then
                _end = i
                break
            end
        end
        if _end == 0 then
            throw(10)
        end
        local words = {}
        for i = 2, _end - 1 do
            table.insert(words, tokens[i].value)
        end

        return rebuild(words) .. "\\0"
    elseif #tokens == 1 and tonumber(tokens[1].value) then
        local number = tonumber(tokens[1].value)
        if number > 32767 or number < -32768 then
            throw(17)
        end
        return tonumber(tokens[1].value)
    end
end

local function findVariable(name)
    for _, variable in pairs(variables) do
        if variable.name == name then
            return variable
        end
    end
    return nil
end

local function findVariableAddress(variable)
    variable = findVariable(variable)
    local offset = variable.address
    if variable.type == "int" then
        return offset, offset + 4
    elseif variable.type == "char*" then
        return offset, (offset + variable.length) - 1
    end
    return nil
end

local tokens_ = {}
local level = 0
local TORETURN = nil
function compile(start, finish, _tokens, where)
    local tokens = {}
    where = where or "code"
    if _tokens == nil then
        tokens = tokens_
    else
        tokens = _tokens
    end
    local next = 0
    for i = 1, #tokens do
        if next ~= 0 then
            if i < next then
                goto continue
            else
                next = 0
            end
        end
        if tokens[i] == nil then
            break
        end
        local token = tokens[i]

        if token.type == "keyword" and vtype[token.value] then
            if level ~= 0 then
                throw(23)
            end
            local var_name = tokens[i + 1].value

            if table.find(reserved_varnames, var_name) then
                throw(26, var_name)
            end

            if tokens[i + 2].value == "=" then 
                -- Variable
                local _end = 0
                local vtokens = {}
                for j = i + 3, finish do
                    if tokens[j].value == ";" then
                        _end = j
                        break
                    else
                        table.insert(vtokens, tokens[j])
                    end
                end
                if _end == 0 then
                    throw(9)
                end

                local offset = nil
                local variable = findVariable(var_name)
                local content = ""
                local _type = token.value
                local sizeof = 0
                local added = false
                if variable then
                    offset = variable.address
                else
                    offset = STACK_OFFSET
                    table.insert(variables, { name = var_name, address = STACK_OFFSET, type = _type })
                    variable = findVariable(var_name)
                    added = true
                end
                write("mov t11, " .. offset)
                write("add r0, sp, t11")

                if token.value == "int" then
                    local number = tonumber(tokens[i + 3].value)
                    if not number then
                        throw(27)
                    end
                    content = tohex16(number)
                    sizeof = 4
                elseif token.value == "char*" then
                    content = string.gsub(parseDef(vtokens), "\\0", "")
                    sizeof = string.len(content)
                    variable.length = sizeof
                    sizeof = sizeof - 1
                end

                write("mbyte r0, [ " .. content .. " ]")

                if added then
                    STACK_OFFSET = STACK_OFFSET + sizeof + 1
                end
 
                next = _end + 1
                goto continue
            end

            -- Function
            if tokens[i + 2].value ~= "(" then
                throw(4)
            end
            --[[
            local args = {}
            ]]--
            local _end = 0
            for j = i + 3, finish do
                if tokens[j].value == ")" then
                    _end = j
                    break
                else
                    --table.insert(args, tokens[j].value)
                end
            end
            if _end == 0 then
                throw(5)
            end

            --[[
            local last_ = {}
            for j = 1, #args do
                if args[j].type == "identifier" then
                    if last_.type == "identifier" then
                        throw(14)
                    end
                    last_ = args[j]
                elseif args[j].type == "symbol" and args[j].value == "," then
                    if last_.type == "symbol" then
                        throw(1, ",")
                    end
                    last_ = args[j]
                else
                    throw(1, "'" .. args[j].value .. "'")
                end
            end
            for j = 1, #args do
                if args[j].value == "," then
                    table.remove(args, j)
                end
            end
            ]]--

            local cdepth = 0
            local __end = 0
            local ftokens = {}
            if tokens[_end + 1].value ~= "{" then
                throw(6)
            else
                cdepth = 1
            end

            for j = _end + 2, finish do
                if tokens[j].value == "{" then
                    cdepth = cdepth + 1
                    table.insert(ftokens, tokens[j])
                elseif tokens[j].value == "}" then
                    cdepth = cdepth - 1
                    if cdepth == 0 then
                        __end = j
                        break
                    else
                        table.insert(ftokens, tokens[j])
                    end
                else
                    table.insert(ftokens, tokens[j])
                end
            end

            if __end == 0 then
                throw(7)
            end

            if var_name ~= "main" then
                write(":" .. var_name .. ":")
            end

            if var_name == "main" and token.value ~= "int" then
                throw(11)
            end

            level = 1
            compile(1, #ftokens, ftokens)
            level = 0

            if var_name ~= "main" then
                if TORETURN ~= nil then
                    write("mov t9, " .. TORETURN)
                end
                write(":" .. var_name .. ":")
            else 
                write("mov r1, 4")
                write("mov r2, " .. (TORETURN or "0"))
                write("syscall")
                if TORETURN == nil then
                    throw(32, "", "info")
                end
            end
           
            TORETURN = nil
            next = __end + 1
        elseif token.type == "keyword" and not vtype[token.value] then 
            if token.value == "include" then
                local vtokens = {}

                local _end = 0
                for j = i + 1, finish do
                    table.insert(vtokens, tokens[j])
                    if tokens[j].value == "\"" and j > i + 1 then
                        _end = j
                        break 
                    end
                end
                if _end == 0 then
                    throw(9)
                end

                local filename = parseDef(vtokens) 
                filename = string.gsub(filename, "[\\0%s]", "")
                if string.find(filename, ".asm$") then
                    write(";- include " .. filename, "funcs")
                elseif string.find(filename, ".c$") then
                    local file = io.open(filename, 'r')
                    if not file then
                        error("File could not be opened '" .. filename .. "'")
                    end
                    local contents = file:read("a")
                    file:close()
                    local __tokens = tokenize(contents)
                    compile(1, #__tokens, __tokens)
                else
                    throw(18, filename)
                end
                next = _end + 1
            elseif token.value == "define" then
                local vtokens = {}
                local var_name = tokens[i + 1].value

                local _end = 0
                for j = i + 2, finish do
                    table.insert(vtokens, tokens[j])
                    if tokens[j].value == "\"" and j > i + 2 then
                        _end = j 
                        break
                    end
                end
                if _end == 0 then
                    throw(9)
                end

                local value = parseDef(vtokens)

                if tonumber(value) then
                    write(var_name .. ": dw " .. tostring(value) .. " :" .. var_name)
                else
                    write(var_name .. ": db \"" .. value .. "\" :" .. var_name)
                end

                next = _end + 1
            elseif token.value == "pragma" then
                local _end = 0
                -- This one is a fun one since pragma can recognize ANY sequence of tokens
                for j = i + 1, finish do 
                    if tokens[j].type == "keyword" then
                        _end = j
                        break
                    end
                end
               if tokens[i + 1].value == "__stack_size__" then
                    if not tonumber(tokens[i + 2].value) then throw(27) end
                    STACK_SIZE = tonumber(tokens[i + 2].value) 
                    _end = i + 3
                else
                    throw(28, "'" .. tokens[i + 1].value .. "'", "info") 
                end
                next = _end
            elseif token.value == "return" then
                TORETURN = tokens[i + 1].value
                if tokens[i + 2].value ~= ";" then
                    throw(9)
                end
                next = i + 3
            elseif token.value == "if" then
                if level == 0 then
                    throw(24)
                end

                if tokens[i + 1].value ~= "(" then
                    throw(4)
                end

                local condend = 0

                local condtokens = {}
                for j = i + 2, finish do
                    if tokens[j].value == ")" then
                        condend = j
                        break
                    elseif tokens[j].value == "(" then
                        throw(1, '(')
                    else
                        table.insert(condtokens, tokens[j])
                    end
                end

                if condend == 0 then
                    throw(5)
                end

                -- Handle condition
                local exptokens = {}
                local restokens = {}
                local split = 0
                for j = 1, #condtokens do
                    if condtokens[j].value == "==" or condtokens[j].value == "!=" then
                        split = j
                        break
                    else
                        table.insert(exptokens, condtokens[j])
                    end
                end
                if split == 0 then
                    throw(19)
                end

                for j = split + 1, #condtokens do
                    table.insert(restokens, condtokens[j])
                end

                if #exptokens > 1 or #restokens > 1 then
                    throw(25)
                end

                -- Handle code part
                local cdepth = 0
                if tokens[condend + 1].value ~= "{" then
                    throw(6)
                else
                    cdepth = 1
                end

                local fend = 0
                local ftokens = {}
                for j = condend + 2, finish do
                    if tokens[j].value == "}" then
                        cdepth = cdepth - 1

                        if cdepth == 0 then
                            fend = j
                            break
                        else
                            table.insert(ftokens, tokens[j])
                        end
                    elseif tokens[j].value == "{" then
                        cdepth = cdepth + 1
                        table.insert(ftokens, tokens[j])
                    else
                        table.insert(ftokens, tokens[j])
                    end
                end
                if fend == 0 then
                    throw(7)
                end

                write(":lcc_" .. tostring(tempcounter) .. ":")
                for j = 1, #ftokens do
                    print("FTOKEN:", ftokens[j].value)
                end
                compile(1, #ftokens, ftokens)
                write(":lcc_" .. tostring(tempcounter) .. ":")

                write("mov t7, " .. exptokens[1].value)
                write("mov t8, " .. restokens[1].value)
                write("mov r4, lcc_" .. tostring(tempcounter))
                write("cmp t7, t8")
                write("jnz r5")

                tempcounter = tempcounter + 1
                next = fend + 1
            end
        elseif tokens[i].type == "identifier" and tokens[i + 1].value == "(" then
            -- Function call
            -- ( is prechecked
            if level < 1 then
                throw(24)
            end

            local aend = 0
            local args = {}
            for j = i + 2, finish do
                if tokens[j].value == ")" then
                    aend = j
                    break
                else
                    table.insert(args, tokens[j])
                end
            end
            if aend == 0 then
                throw(5)
            end
            if tokens[aend + 1].value ~= ";" then
                throw(9) -- cheap shot but it works LOL
            end

            local last_ = {}
            for j = 1, #args do
                if args[j].type == "identifier" or args[j].type == "number" then
                    if last_.type == "identifier" then
                        throw(14)
                    end
                    last_ = args[j]
                elseif args[j].type == "symbol" and args[j].value == "," then
                    if last_.type == "symbol" then
                        throw(1, ",")
                    end
                    last_ = args[j]
                else
                    throw(1, "'" .. args[j].value .. "'")
                end
            end
            for j = 1, #args do
                if args[j] == nil then goto continue end
                if args[j].value == "," then
                    table.remove(args, j)
                end

                ::continue::
            end

            

            if not table.find(reserved_varnames, tokens[i].value) then
                for j = 1, #args do
                    write("mov t" .. j .. ", " .. args[j].value)
                end

                write("mov r4, " .. tokens[i].value)
                write("jmp")
            else
                if tokens[i].value == "__setbytes__" then
                    local bytes = ""
                    local found = false
                    for i = 1, #cstrings do
                        if cstrings[i][1] == args[2].value then
                            bytes = cstrings[i][2]
                            break
                        end
                    end
                    bytes = string.gsub(bytes, "\\0", "")
                    write("mov t1, " .. args[1].value)
                    write("stn t1")
                    write("mbyte t1, [ " .. bytes .. " ]")
                elseif tokens[i].value == "__setregister__" then
                    write("mov " .. args[1].value .. ", " .. args[2].value)
                elseif tokens[i].value == "__write__" then
                    write("mov r1, 1")
                    local start, _end = findVariableAddress(args[1].value)
                    if not start then
                        throw(31, "'" .. args[1].value .. "'")
                    end
                    write("mov r2, " .. start)
                    write("mov r3, " .. _end)
                    write("add r2, r2, sp")
                    write("add r3, r3, sp")
                    write("syscall")
                end
            end

            next = aend + 2
        else
            print(tokens[i - 1].value or "")
            print(tokens[i].value)
            print(tokens[i + 1].value or "")
            throw(3)
        end

        ::continue::
    end
end

local infile = arg[1]
local outfile = arg[2]

if not infile or not outfile then
    error("Please provide a source and output file!")
end

local _infile = io.open(infile, 'r')
if not _infile then
    error("Source file not found")
end
local contents = _infile:read("a")
_infile:close()
tokens_ = tokenize(contents)

compile(1, #tokens_, tokens_)

if STACK_SIZE > 0 then
    write("stack " .. STACK_SIZE, true)
else 
    throw(29, "", "warning")
end

if STACK_OFFSET > STACK_SIZE then
    throw(30, "(allocated: " .. STACK_SIZE .. ", actual: " .. STACK_OFFSET .. ")", "warning")
end

local _outfile = io.open(outfile, "w")
if not _outfile then
    error("Could not create output file")
end

_outfile:write(buffer)
_outfile:close()
