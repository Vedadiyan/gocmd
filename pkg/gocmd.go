package gocmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Token struct {
	must     map[string]*string
	optional map[string]string
	flags    map[string]bool
	help     map[string]string
}

type Command struct {
	commands map[string]Token
	help     map[string]string
}

func Sort(input map[string]string) (longest int, sorted *[]string) {
	sortedKeys := make([]string, 0)
	for key, value := range input {
		_ = value
		len := len(key)
		if len > longest {
			longest = len
		}
		sortedKeys = append(sortedKeys, key)
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})
	return longest, &sortedKeys
}

func New() *Command {
	command := &Command{}
	command.commands = make(map[string]Token)
	command.help = make(map[string]string)
	return command
}

func (token Token) GetMust(command string) *string {
	return token.must[command]
}

func (token Token) GetOptional(command string) string {
	return token.optional[command]
}

func (token Token) GetFlag(command string) bool {
	return token.flags[command]
}

func (token Token) PrintHelp() {
	longest, sortedKeys := Sort(token.help)
	for _, key := range *sortedKeys {
		_, isMust := token.must[key]
		_, isOptional := token.must[key]
		if isMust || isOptional {
			fmt.Printf("-")
			fmt.Print(key)
			for i := 0; i < (longest-len(key)-1)+10; i++ {
				fmt.Print(" ")
			}
		} else if _, ok := token.flags[key]; ok {
			fmt.Printf("--")
			fmt.Print(key)
			for i := 0; i < (longest-len(key)-2)+10; i++ {
				fmt.Print(" ")
			}
		} else {
			panic("Unknown Case")
		}
		fmt.Println(token.help[key])
	}
}

func (token *Token) RegisterCommand(cmd string, help string, defaultValue *string) *Token {
	if defaultValue != nil {
		token.optional[cmd] = *defaultValue
	} else {
		token.must[cmd] = nil
	}
	token.help[cmd] = help
	return token
}
func (token *Token) RegisterFlag(cmd string, help string) *Token {
	token.flags[cmd] = false
	token.help[cmd] = help
	return token
}
func (command *Command) RegisterGroup(group string, help string) *Token {
	token := Token{}
	token.must = make(map[string]*string)
	token.optional = make(map[string]string)
	token.flags = make(map[string]bool)
	token.help = make(map[string]string)
	command.commands[group] = token
	command.help[group] = help
	return &token
}
func (command *Command) Parse() (string, *Token, error) {
	commands := make(map[string]string)
	flags := make(map[string]bool)
	var group string
	var prev *string
	for i := 1; i < len(os.Args); i++ {
		val := os.Args[i]
		if strings.HasPrefix(val, "--") {
			_val := strings.TrimPrefix(val, "--")
			flags[_val] = true
			continue
		}
		if strings.HasPrefix(val, "-") {
			_val := strings.TrimPrefix(val, "-")
			prev = &_val
			continue
		}
		if i == 1 {
			group = val
			continue
		}
		if prev == nil {
			panic("Invalid Command Line Argument")
		}
		commands[*prev] = val
		prev = nil
	}
	value, ok := command.commands[group]
	if !ok {
		command.PrintHelp()
		return "", nil, errors.New("Command group not found")
	}
	errs := make([]string, 0)
	for key, val := range value.must {
		_ = val
		_value, ok := commands[key]
		if !ok {
			errs = append(errs, fmt.Sprintf("-%s is missing", key))
			continue
		}
		value.must[key] = &_value
	}
	for key, val := range value.optional {
		_ = val
		value.optional[key] = commands[key]
	}
	for key, val := range value.flags {
		_ = val
		_ = val
		_, ok := flags[key]
		if ok {
			value.flags[key] = true
		}
	}
	if len(errs) != 0 {
		value.PrintHelp()
		return "", nil, errors.New(strings.Join(errs, "\r\n"))
	}
	return group, &value, nil
}

func (command Command) PrintHelp() {
	longest, sortedKeys := Sort(command.help)
	for _, key := range *sortedKeys {
		fmt.Print(key)
		for i := 0; i < (longest-len(key))+10; i++ {
			fmt.Print(" ")
		}
		fmt.Println(command.help[key])
	}
}
