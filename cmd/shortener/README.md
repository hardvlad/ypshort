# cmd/shortener

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/shortener` содержит:
- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения

# Результаты оптимизации

Получилось оптимизировать использование памяти (`7.11%`) за счет увеличения начального размера мапы хранения данных и переноса сохранения базы данных из синхронного в асинхронный режим 

Showing nodes accounting for `-677.03kB`, `7.11%` of 9528.40kB total
      flat  flat%   sum%        cum   cum%
-1301.24kB 13.66% 13.66% -1301.24kB 13.66%  bytes.growSlice
  618.49kB  6.49%  7.17%  -682.75kB  7.17%  encoding/json.mapEncoder.encode
  516.76kB  5.42%  1.74%   516.76kB  5.42%  runtime.procresize
  513.69kB  5.39%  3.65%   513.69kB  5.39%  golang.org/x/text/internal/language.map.init.1
    -513kB  5.38%  1.73%     -513kB  5.38%  runtime.allocm
  512.31kB  5.38%  3.64%   512.31kB  5.38%  net.newFD (inline)
  512.05kB  5.37%  9.02%   512.05kB  5.37%  time.NewTicker
 -512.05kB  5.37%  3.64%  -512.05kB  5.37%  github.com/hardvlad/ypshort/internal/handler.deleteWorker
 -512.05kB  5.37%  1.73%  -512.05kB  5.37%  runtime.acquireSudog
 -512.01kB  5.37%  7.11%  -512.01kB  5.37%  runtime.(*timers).addHeap
         0     0%  7.11% -1301.24kB 13.66%  bytes.(*Buffer).Write

# Для передачи значений переменных сборки

используйте команду 

go build -ldflags \
"-X main.buildVersion=v1.2.3 \
-X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
-X main.buildCommit=$(git rev-parse HEAD)" \
./cmd/shortener

Или, например,

go build -ldflags "-X main.buildVersion=v1.2.3 -X main.buildDate=04.03.2026 -X main.buildCommit=0a0b0c" ./cmd/shortener


