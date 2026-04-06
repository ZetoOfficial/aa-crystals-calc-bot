package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrEmpty          = errors.New("empty command")
	ErrUnknownCommand = errors.New("unknown command")
	ErrUsage          = errors.New("usage: N")
	ErrInvalidNumber  = errors.New("N must be a positive integer")
	ErrTooLarge       = errors.New("N is too large")
)

type Command struct {
	Name string
	SHK  int
}

func Parse(input string, maxSHK int) (Command, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return Command{}, ErrEmpty
	}

	if strings.HasPrefix(fields[0], "@") {
		fields = fields[1:]
	}
	if len(fields) == 0 {
		return Command{}, ErrEmpty
	}

	command := "calc"
	value := fields[0]
	if strings.ToLower(fields[0]) == "куса" {
		command = "куса"
		if len(fields) != 2 {
			return Command{}, ErrUsage
		}
		value = fields[1]
	} else if len(fields) != 1 {
		return Command{}, ErrUnknownCommand
	}

	shk, err := strconv.Atoi(value)
	if err != nil || shk <= 0 {
		return Command{}, ErrInvalidNumber
	}
	if maxSHK > 0 && shk > maxSHK {
		return Command{}, fmt.Errorf("%w: max is %d", ErrTooLarge, maxSHK)
	}

	return Command{Name: command, SHK: shk}, nil
}

func UserMessage(err error) string {
	switch {
	case errors.Is(err, ErrEmpty), errors.Is(err, ErrUsage):
		return "Формат: 100"
	case errors.Is(err, ErrInvalidNumber):
		return "Количество ШК должно быть положительным целым числом. Пример: 100"
	case errors.Is(err, ErrTooLarge):
		return "Слишком большое количество ШК. Попробуй меньшее значение."
	default:
		return "Не понял команду. Пример: 100"
	}
}
