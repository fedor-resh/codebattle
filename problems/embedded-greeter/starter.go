package solution

type Greeter interface {
	Greet() string
}

type Base struct {
	Name string
}

func (base Base) Greet() string {
	// Напишите решение
	return ""
}

type Shouty struct {
	Base
}

func (shouty Shouty) Greet() string {
	// Напишите решение
	return ""
}

type Polite struct {
	Base
}

func (polite Polite) Greet() string {
	// Напишите решение
	return ""
}

func greeterFor(kind, name string) (Greeter, bool) {
	base := Base{Name: name}
	switch kind {
	case "base":
		return base, true
	case "shouty":
		return Shouty{Base: base}, true
	case "polite":
		return Polite{Base: base}, true
	default:
		return nil, false
	}
}

// Solve адаптирует интерфейс Greeter к JSON-совместимому контракту judge.
func Solve(kinds []string, names []string) []string {
	count := len(kinds)
	if len(names) < count {
		count = len(names)
	}
	result := make([]string, 0, count)
	for index := 0; index < count; index++ {
		greeter, ok := greeterFor(kinds[index], names[index])
		if !ok {
			continue
		}
		result = append(result, greeter.Greet())
	}
	return result
}
