package keyboard

import "strings"

var CharTable = [][2]rune{
    {'`', '~'},
    {'1', '!'},
    {'2', '@'},
    {'3', '#'},
    {'4', '$'},
    {'5', '%'},
    {'6', '^'},
    {'7', '&'},
    {'8', '*'},
    {'9', '('},
    {'0', ')'},
    {'-', '_'},
    {'=', '+'},
    {'[', '{'},
    {']', '}'},
    {'\\', '|'},
    {';', ':'},
    {'\'', '"'},
    {',', '<'},
    {'.', '>'},
    {'/', '?'},
}

var Shift bool = false 

func Lower(str string) string {
    if len(str) == 0 {
        return ""
    }
    char := rune(str[0])
    for _, pair := range CharTable {
        if pair[1] == char {
            return string(pair[0])
        }
    }

	char = rune(strings.ToLower(string(char))[0])
    return string(char)
}

func Upper(str string) string {
    if len(str) == 0 {
        return ""
    }
    char := rune(str[0])
    for _, pair := range CharTable {
        if pair[0] == char {
            return string(pair[1])
        }
    }

	char = rune(strings.ToUpper(string(char))[0])
    return string(char)
}
