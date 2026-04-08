package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// MaxSHK — верхний предел количества ШК, разрешённый в одной команде.
// Подобран с запасом, чтобы калькулятор не уходил в многомегабайтный DP.
const MaxSHK = 100000

var (
	ErrEmpty          = errors.New("empty command")
	ErrUnknownCommand = errors.New("unknown command")
	ErrUsage          = errors.New("usage: N")
	ErrInvalidNumber  = errors.New("n must be a positive integer")
	ErrTooLarge       = errors.New("n is too large")
)

const (
	CommandCalc = "calc"
	CommandHelp = "help"
	CommandKusa = "куса"
)

type Command struct {
	Name string
	SHK  int
}

// Service — реализация парсинга команд бота.
// Безсостоятельная: можно использовать через нулевое значение.
type Service struct{}

// Parse разбирает входное сообщение пользователя в Command.
func (Service) Parse(input string) (Command, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return Command{}, ErrEmpty
	}

	first := strings.ToLower(fields[0])

	if first == "помощь" || first == "help" {
		if len(fields) != 1 {
			return Command{}, ErrUnknownCommand
		}
		return Command{Name: CommandHelp}, nil
	}

	command := CommandCalc
	value := fields[0]
	if first == CommandKusa {
		command = CommandKusa
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
	if shk > MaxSHK {
		return Command{}, fmt.Errorf("%w: max is %d", ErrTooLarge, MaxSHK)
	}

	return Command{Name: command, SHK: shk}, nil
}

// UserMessage переводит ошибку парсера в текст для пользователя.
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
