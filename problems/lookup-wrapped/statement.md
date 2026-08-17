# Обёрнутая ошибка

Допишите `Lookup(scores map[string]int, name string) (int, error)`.

Если имя есть в карте, верните очки и `nil`. Если имени нет, верните `0` и ошибку, обёрнутую через `fmt.Errorf` с `%w` вокруг sentinel `ErrNotFound`. В тексте ошибки должно встречаться имя игрока.

Проверяющая система вызывает `Lookup` напрямую и использует `errors.Is`, поэтому не удаляйте `ErrNotFound` и не меняйте сигнатуру.

`Solve(scores map[string]int, names []string) []string` уже вызывает `Lookup` для каждого имени и возвращает число либо `err.Error()`.
