# cmd/shortener

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/shortener` содержит:
- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения

- github.com/hardvlad/ypshort/internal/handler.GetShortCode

C:\Projects\ypshort\internal\handler\handlers.go

Total:           0     656926 (flat, cum)  3.66%
430            .          .           	var shortLink string
431            .          .           	var urlAlreadyExisted bool
432            .          .            
433            .          .           	for i := 0; i < maxAttempts; i++ {
434            .     262148           		shortLink = GenerateRandomString(data.Conf)
435            .     394778           		code, urlExisted, err := data.Store.Set(shortLink, body, userID)
436            .          .           		if err != nil {
437            .          .           			if errors.Is(err, repository.ErrorKeyExists) {
438            .          .           				continue
439            .          .           			} else {
440            .          .           				return false, "", false, err 
