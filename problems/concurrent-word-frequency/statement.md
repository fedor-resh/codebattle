# Конкурентная частота слов

Подсчитайте частоту каждого слова с помощью `workers` горутин. Все workers должны обновлять общую map, защищённую `sync.Mutex` или `sync.RWMutex`, а `Solve` — дождаться их завершения через `sync.WaitGroup`.

Если `workers <= 0`, используйте одного worker. Регистр символов учитывается. Для пустого ввода верните пустую map.

Реализуйте функцию `Solve(words []string, workers int) map[string]int` в пакете `solution`.
