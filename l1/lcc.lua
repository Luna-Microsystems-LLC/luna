--[[
 * Luna L-Series License
 * 
 * This file is part of the Luna CPU architecture and toolchain,
 * developed by Luna Microsystems LLC.
 *
 * Licensed under the CERN Open Hardware Licence Version 2 – Strongly Reciprocal
 * (CERN-OHL-S v2) with Additional Terms 
 * (Non-Commercial Use, Attribution, Patent Waiver).
 * 
 * You may use and modify this work with attribution for non-commercial purposes.
 * Commercial use requires written permission from Luna Microsystems LLC.
 *
 * Full license text: see the LICENSE file in the repository root.
 *
]]--

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
local code = ""

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
    ["else"] = true,
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
    [">"] = true,
    ["<"] = true,
}

local operator_d = {
    ["=="] = true, 
    ["!="] = true,
    [">="] = true,
    ["<="] = true,
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
    [32] = "Consider adding an explicit return value to 'main' function",
    [33] = "Reference to undefined/implicit variable/function",
    [34] = "Unused variable/function",
    [35] = "LASM not found",
    [36] = "Assembling program failed",
    [37] = "Attempt redeclare undefined variable",
    [38] = "Length overflow"
}

local reserved_varnames = {
    "__setregister__",
    "__jmp__",
    "__stack_size__",
    "__extern__",
    "__nop__",
    "goto"
}

local STACK_SIZE = 0
local STACK_OFFSET = 0

local function write(where, text, prepend)
    if where == "defs" then
        if not prepend then
            defs = defs .. text .. "\n"
        else
            defs = text .. "\n" .. defs
        end
    elseif where == "funcs" then
        funcs = funcs .. text .. "\n"
    else
        code = code .. text .. "\n"
    end
end

local variables = {}

-- insert registers
for i = 0, 12 do
    table.insert(variables, {
        name = "r" .. i,
        type = "any",
        ignoreunused = true,
        scope = 0,
    })
    if i ~= 0 then
        table.insert(variables, {
            name = "t" .. i,
            type = "any",
            ignoreunused = true,
            scope = 0,
        })
    end
end
table.insert(variables, { name = "pc", type = "any", ignoreunused = true, scope = 0, })
table.insert(variables, { name = "ptr", type = "any", ignoreunused = true, scope = 0, })
table.insert(variables, { name = "sp", type = "any", ignoreunused = true, scope = 0, })

local function tohex16(n)
    return string.format("%04X", n & 0xFFFF)
end

local function throw(_error, eargs, _type) 
    if _type == "error" or _type == nil then
        print("\27[31mError " .. tostring((_error or "(no error number)")) .. ": " .. (errors[_error] or "(no error text)") .. " " .. (eargs or "") .. "\27[0m") 
        if table.find(arg, "--error-extra-info") then
            print("Buffer dump: " .. buffer)
            print(debug.traceback())
        end
        os.exit(tonumber(_error) or 1)
    elseif _type == "warning" then
        print("\27[33mWarning " .. tostring((_error or "(no warning number)")) .. ": " .. (errors[_error] or "(no error text)") .. " " .. (eargs or "") .. "\27[0m")
    elseif _type == "info" then
        print("\27[34mInfo " .. tostring((_error or "(no info number)")) .. ": " .. (errors[_error] or "(no info text)") .. " " .. (eargs or "") .. "\27[0m") 
    end
end

local function tokenize(text)
    local tokens = {}
    local i = 1
    local quotes = false
    local comment = false
    local current_str = ""
    local lastslash = true

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
        elseif string.match(c, "[%a_@*#\\]") then
            local start = i
            local matched = false
            while string.match(peek(), "[%w_*@\\]") do
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
        elseif c == "/" then
            if string.sub(text, i + 1, i + 1) == "/" then
                while peek() ~= "\n" do advance() end
                advance()
            else
                table.insert(tokens, { type = "operator", value = c }) 
            end
        elseif operator_d[peek(1)] then
            table.insert(tokens, { type = "operator", value = peek(1) })
            advance(2)
        elseif operator_s[c] then
            local function insert()
                table.insert(tokens, { type = "operator", value = c })
                advance()
            end
            insert() 
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

local function hoist(buffer)
    local tokens = {}
    local rc = 1
    local funcs = {}
    local ordered = {}

    for token in string.gmatch(buffer, "[^\n]+") do
        table.insert(tokens, token)
    end

    local depth = 0
    local depths = {}
    local current = nil

    for i, token in ipairs(tokens) do
        if string.sub(token, 1, 1) == ":" and string.sub(token, #token, #token) == ":" then
            if not funcs[token] then
                funcs[token] = {}
                table.insert(funcs[token], token)
                table.insert(ordered, funcs[token])
                current = funcs[token]
                depth = depth + 1
                depths[depth] = funcs[token]
            else
                table.insert(funcs[token], token)
                depth = depth - 1
                if depth > 0 then
                    current = depths[depth]
                else
                    current = nil
                end
            end
        else
            if current ~= nil then
                table.insert(current, token)
            else
                funcs[rc] = {}
                table.insert(funcs[rc], token)
                table.insert(ordered, funcs[rc])
                rc = rc + 1
            end
        end
    end

    local reconstruct = ""
    for _, array in ipairs(ordered) do
        for i = 1, #array do
            reconstruct = reconstruct .. array[i] .. "\n"
        end
    end

    return reconstruct
end

local function rebuildBuffer()
    local funcs_ = hoist(funcs)
    buffer = buffer .. "; Definitions ;\n\n" .. defs .. "\n\n; Functions ;\n\n" .. funcs_ .. "\n\n; Code ;\n\n" .. code 
end

local scopes = {
    {0, nil}
}

local cscope = 1

local function createScope(parent)
    table.insert(scopes, {cscope, parent})
    local _scope = cscope
    cscope = cscope + 1
    return _scope
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

local function findVariable(name, scope)
    local function findParent(scope)
        for i = 1, #scopes do
            if scopes[i][1] == scope then
                return scopes[i][2]
            end
        end
    end
    for _, variable in pairs(variables) do
        if variable.name == name then
            if variable.scope == scope then
                variable.used = true
                return variable
            else
                local parent = findParent(scope)
                while parent ~= nil do
                    if variable.scope == parent then
                        variable.used = true
                        return variable
                    else
                        parent = findParent(parent)
                    end
                end
            end
        end
    end
    return nil
end

local function resolveVariables(args, scope)
    for i, argument in pairs(args) do
        local variable = findVariable(argument.value, scope)
        if variable then
            if variable.type == "char_ptr" then
                args[i].value = variable.address
            elseif variable.type == "func" then
                args[i].value = "t9"
            end
        elseif not variable and argument.type == "identifier" then
            if not table.find(arg, "--allow-implicit") then
                throw(33, "'" .. argument.value .. "'")
            else
                throw(33, "'" .. argument.value .. "'", "warning")
            end
        end
    end
    return args
end

local tokens_ = {}
local level = 0
local TORETURN = nil
local function compile(start, finish, _tokens, where, scope)
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

        if token.type == "identifier" and tokens[i + 1].value == "=" then
            local _end = 0
            local vtokens = {}
            for j = i + 2, finish do
                if tokens[j].value == ";" then
                    _end = j
                    break
                else
                    table.insert(vtokens, tokens[j])
                end 
            end

            local var_name = token.value
            local variable = findVariable(var_name, scope) 

            if not variable then
                throw(37, "'" .. var_name .. "'")
            end

            if variable.type == "char_ptr" then 
                local value = string.gsub(parseDef(vtokens), "\\0", "") .. "\0"
                if string.len(value) > variable.length then
                    throw(38, "for variable '" .. var_name .. "'", "warning") 
                end
                write(where, "mov r0, " .. variable.address)
                write(where, "add r0, r0, sp")
                write(where, "mbyte r0, [ " .. value .. " ]")
            end
            next = _end + 1
        elseif token.type == "keyword" and vtype[token.value] then
            if level ~= 0 then
                -- throw(23)
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

                if token.value == "int" then
                    table.insert(variables, {
                        name = var_name,
                        type = "int",
                        scope = scope
                    })
                    write("defs", var_name .. ": dw" .. tohex16(tokens[i + 3].value) .. " :" .. var_name)
                elseif token.value == "char*" then
                    local content = string.gsub(parseDef(vtokens), "\\0", "")
                    content = string.gsub(content, "\0", "")
                    content = content .. "\0"
                    table.insert(variables, {
                        name = var_name,
                        type = "char_ptr",
                        address = STACK_OFFSET,
                        scope = scope,
                        length = string.len(content)
                    })
                    write(where, "mov r0, " .. STACK_OFFSET)
                    write(where, "add r0, r0, sp")
                    write(where, "mbyte r0, [ " .. content .. " ]")
                    STACK_OFFSET = STACK_OFFSET + string.len(content)
                elseif token.value == "void" then
                    if tokens[i + 4].value == "(" then
                        table.insert(variables, {
                            name = var_name,
                            type = "func",
                            scope = scope
                        })
                        local ftokens = {}
                        local __end = 0
                        for j = i + 3, _end do
                            table.insert(ftokens, tokens[j]) 
                            if tokens[j].value == ";" then
                                __end = j
                                break
                            end
                        end
                        compile(1, #ftokens, ftokens, where, scope) 
                    end
                end 
 
                next = _end + 1
                goto continue
            end

            -- Function
            if tokens[i + 2].value ~= "(" then
                throw(4)
            end

            local _end = 0
            for j = i + 3, finish do
                if tokens[j].value == ")" then
                    _end = j
                    break 
                end
            end
            if _end == 0 then
                throw(5)
            end 

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

            table.insert(variables, {
                name = var_name,
                type = "function",
                used = (var_name == "main" and true or false)
            })

            if var_name ~= "main" then
                write("funcs", ":" .. var_name .. ":")
            end

            if var_name == "main" and token.value ~= "int" then
                throw(11)
            end

            local cwhere = "funcs"

            if var_name == "main" then
                cwhere = "code"
            end

            local newScope = createScope(scope)
            level = 1
            compile(1, #ftokens, ftokens, cwhere, newScope)
            level = 0

            if var_name ~= "main" then
                write("funcs" ,":" .. var_name .. ":")
            else 
                write("code", "mov r1, 4")
                write("code", "mov r2, " .. (TORETURN or "0"))
                write("code", "int")
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
                    write("defs", ";- include " .. filename)
                elseif string.find(filename, ".c$") or string.find(filename, ".h$") then
                    local file = io.open(filename, 'r')
                    if not file then
                        error("File could not be opened '" .. filename .. "'")
                    end
                    local contents = file:read("a")
                    file:close()
                    local __tokens = tokenize(contents)
                    compile(1, #__tokens, __tokens, nil, 0)
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
                    throw(9, ", near" .. table.concat(vtokens, ' '))
                end

                local value = parseDef(vtokens)

                if tonumber(value) then
                    write("defs", var_name .. ": dw " .. tostring(value) .. " :" .. var_name)
                else
                    write("defs", var_name .. ": db \"" .. value .. "\" :" .. var_name)
                end

                table.insert(variables, {
                    name = var_name,
                    type = "label",
                    scope = 0,
                })

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
                elseif tokens[i + 1].value == "__extern__" then
                    table.insert(variables, {
                        name = tokens[i + 2].value,
                        type = "any"
                    })
                    _end = i + 3
                elseif tokens[i + 1].value == "__force_size__" then
                    write("defs" ,";- size " .. tokens[i + 2].value) 
                    _end = i + 3
                elseif tokens[i + 1].value == "point" then
                    write(where, "mov t12, pc")
                    write(where, "nop")
                    _end = i + 2
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
                local exp = ""
                for j = 1, #condtokens do
                    if operator_d[condtokens[j].value] or condtokens[j].value == ">" or condtokens[j].value == "<" then
                        split = j
                        exp = condtokens[j].value
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

                exptokens = resolveVariables(exptokens, scope)
                restokens = resolveVariables(restokens, scope)

                -- Handle code part
                local cdepth = 0
                if tokens[condend + 1].value ~= "{" then
                    throw(6)
                else
                    cdepth = 1
                end

                local tend = 0
                local fend = 0
                local eend = 0

                local ftokens = {}
                local etokens = {}

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

                local isElse = false

                if tokens[fend + 1] and tokens[fend + 1].value == "else" then
                    isElse = true
                    local cdepth = 0
                    if tokens[fend + 2].value ~= "{" then
                        throw(6)
                    else
                        cdepth = 1
                    end
                    
                    for k = fend + 3, finish do
                        if tokens[k].value == "{" then
                            cdepth = cdepth + 1
                            table.insert(etokens, tokens[k])
                        elseif tokens[k].value == "}" then
                            cdepth = cdepth - 1
                            if cdepth == 0 then
                                eend = k
                                break
                            else
                                table.insert(etokens, tokens[k])
                            end
                        else
                            table.insert(etokens, tokens[k])
                        end
                    end

                    if eend == 0 then
                        throw(7)
                    end

                    tend = eend
                else
                    tend = fend
                end

                local method = ""
                local op = ""
                local op_reverse = ""

                if exp == "==" then
                    method = "cmp"
                    op = "jnz"
                    op_reverse = "jz"
                elseif exp == "!=" then
                    method = "cmp"
                    op = "jz"
                    op_reverse = "jnz"
                elseif exp == ">" then
                    method = "igt"
                    op = "jnz"
                    op_reverse = "jz"
                elseif exp == "<" then
                    method = "ilt"
                    op = "jnz"
                    op_reverse = "jz"
                elseif exp == ">=" then
                    method = "iget"
                    op = "jnz"
                    op_reverse = "jz"
                elseif exp == "<=" then
                    method = "ilet"
                    op = "jnz"
                    op_reverse = "jz"
                end

                local ifLoc = tempcounter
                local elseLoc
                write("funcs", ":lcc_" .. tostring(tempcounter) .. ":")
                local newScope = createScope(scope)
                compile(1, #ftokens, ftokens, "funcs", newScope)
                write("funcs", ":lcc_" .. tostring(tempcounter) .. ":")

                if isElse == true then
                    tempcounter = tempcounter + 1
                    elseLoc = tempcounter
                    write("funcs", ":lcc_" .. tostring(tempcounter) .. ":")
                    local newScope = createScope(scope)
                    compile(1, #etokens, etokens, "funcs", newScope)
                    write("funcs", ":lcc_" .. tostring(tempcounter) .. ":")
                end

                write(where, "mov t7, " .. exptokens[1].value)
                write(where, "mov t8, " .. restokens[1].value)
                write(where, "mov r4, lcc_" .. tostring(ifLoc))
                write(where, method .. " t7, t8")
                write(where, op .. " r5")
                if isElse == true then
                    write(where, "mov r4, lcc_" .. tostring(elseLoc))
                    write(where, op_reverse .. " r5")
                end

                tempcounter = tempcounter + 1
                next = tend + 1
            end
        elseif tokens[i].type == "identifier" and tokens[i + 1].value == "(" then
            -- Function call
            -- ( is prechecked
            
            if not findVariable(tokens[i].value) and not table.find(reserved_varnames, tokens[i].value) then
                if not table.find(arg, "--allow-implicit") then
                    throw(33, "'" .. tokens[i].value .. "'")
                else
                    throw(33, "'" .. tokens[i].value .. "'", "warning")
                end
            end

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

            if not tokens[aend + 1] then
                throw(9)
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

            args = resolveVariables(args, scope)

            if #args > 6 then
                throw(20)   
            end

            if not table.find(reserved_varnames, tokens[i].value) then
                for j = 1, #args do
                    write(where, "mov t" .. j .. ", " .. args[j].value)
                end

                write(where, "mov r4, " .. tokens[i].value)
                write(where, "jmp")
            else 
                if tokens[i].value == "__setregister__" then
                    write(where, "mov " .. args[1].value .. ", " .. args[2].value)
                elseif tokens[i].value == "__jmp__" then
                    write(where, "jmp")
                elseif tokens[i].value == "__nop__" then
                    write(where, "nop")
                elseif tokens[i].value == "goto" then
                    write(where, "mov r4, t12")
                    write(where, "jmp")
                end
            end

            next = aend + 2
        else
            -- print(tokens[i - 1].value or "")
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

compile(1, #tokens_, tokens_, nil, 0)

if STACK_SIZE > 0 then
    write("defs", "stack " .. STACK_SIZE, true)
else 
    throw(29, "", "warning")
end

if (STACK_OFFSET - 1) > STACK_SIZE then
    throw(30, "(allocated: " .. STACK_SIZE .. ", actual: " .. STACK_OFFSET - 1 .. ")", "warning")
end

for _, variable in pairs(variables) do
    if variable.used ~= true and variable.ignoreunused ~= true then
        throw(34, "'" .. variable.name .. "'", "warning")
    end
end

rebuildBuffer()

if table.find(arg, "-e") then
    if os.getenv("OS") ~= "Windows_NT" then
        local _, __, returnvalue = os.execute("which lasm > /dev/null")
        if tonumber(returnvalue) ~= 0 then
            throw(35)
        end
    else
        local _, __, returnvalue = os.execute("where lasm > nul")
        if tonumber(returnvalue) ~= 0 then
            throw(35)
        end
    end

    math.randomseed(os.time())
    local filename = "lcc_temporary" .. math.random(10000, 99999) .. ".asm"
    local file = io.open(filename, 'w')
    if not file then
        error("Could not create output file")
    end
    file:write(buffer)
    file:close()
    local function deleteTemp()
        if os.getenv("OS") ~= "Windows_NT" then
            os.execute("rm " .. filename)
        else
            os.execute("del " .. filename)
        end
    end
    local _, __, returnvalue = os.execute("lasm " .. filename .. " " .. outfile)
    if tonumber(returnvalue) ~= 0 then
        deleteTemp()
        throw(36)
    end
    deleteTemp() 
else
    local _outfile = io.open(outfile, "w")
    if not _outfile then
        error("Could not create output file")
    end
    _outfile:write(buffer)
    _outfile:close()
end

