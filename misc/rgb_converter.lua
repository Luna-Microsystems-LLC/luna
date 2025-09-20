#!/usr/bin/env lua

function convert(hex)
    local r = (hex >> 16) & 0xFF
    local g = (hex >> 8) & 0xFF
    local b = hex & 0xFF

    local r3 = (r * 7) // 255
    local g3 = (g * 7) // 255
    local b2 = (b * 3) // 255
 
    return (r3 << 5) | (g3 << 2) | b2
end

print(convert(tonumber(arg[1])) .. " / " .. string.format("0x%02x", convert(tonumber(arg[1]))))
