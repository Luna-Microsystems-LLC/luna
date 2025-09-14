#!/usr/bin/env lua
local infiles = {}
local outfile = nil

local invocation = false
local error_extra_info = false
local noassemble = false

local intermediates = {}
local args = {}

local cleanup

-- Helper functions

local function stderr(message) 
    io.stderr:write(message .. "\n")
end

local function execute(command, outsidecommand)
    local _, __, code = os.execute(command)
   
    if invocation == true then
        print(command)
    end

    if tonumber(code) ~= 0 then
        if outsidecommand == true then
            stderr("\27[1;37mlcc: \27[1;31merror: \27[1;37mcompilation command failed with exit code " .. code .. " (use -v to see invocation)\27[0m")
            cleanup()
            os.exit(1)
        else
            cleanup()
            os.exit(1)
        end
    end
end

cleanup = function()
    for i = 1, #intermediates do
        if intermediates[i][2] == true then
            execute("rm -f " .. intermediates[i][1])
        end
    end
end
-- Parse flags

local Operands = 0
local i = 1
while i <= #arg do
    local argument = arg[i]

    if argument == "-o" then
        outfile = arg[i + 1]
        i = i + 2
    elseif argument == "-v" then
        invocation = true
        i = i + 1
    elseif argument == "-s" then
        noassemble = true
        i = i + 1
    elseif argument == "--error-extra-info" then
        table.insert(args, "--error-extra-info")
        i = i + 1
    elseif argument == "--allow-implicit" then
        table.insert(args, "--allow-implicit")
        i = i + 1
    else
        table.insert(infiles, argument)
        i = i + 1
    end
end

if #infiles < 1 then
    stderr("\27[1;37mlcc: \27[1;31merror: \27[1;37mno input files\27[0m")
    os.exit(1)
end

-- First pass: compile high-level languages to assembly
for i = 1, #infiles do
    local file = infiles[i]
    local name, extension = string.match(file, "(.+)%.([^.]+)$")

    if extension == "c" or extension == "h" then
        execute("lcc1 -s " .. table.concat(args, ' ') .. file .. " -o " .. name .. ".s")
        table.insert(intermediates, {name .. ".s", true})
    elseif extension == "s" or extension == "asm" then
        table.insert(intermediates, {name .. "." .. extension, false})
        goto continue
    else
        stderr("\27[1;37mlcc: \27[1;31merror: \27[1;37munknown file type '" .. extension .. "'\27[0m")
        cleanup()
        os.exit(1)
    end

    ::continue::
end

if noassemble == true then
    os.exit(0)
end
-- Second pass: assemble and link all files compiled down to assembly

local command = "las"
command = command .. table.concat(args, ' ') .. ' '
for i = 1, #intermediates do
    local filename = intermediates[i][1]
    command = command .. " " .. filename
end

if outfile == nil then
    outfile = "a.bin"
end
command = command .. " -o " .. outfile
execute(command)

-- Third pass: clean up all intermediary files

cleanup()
