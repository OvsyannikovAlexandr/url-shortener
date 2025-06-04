#requires -Version 5.0
Import-Module InvokeBuild

task Run {
    Write-Host "Запуск приложения локально..."
    go run ./cmd/main.go
}

task Build {
    Write-Host "Сборка бинарника..."
    go build -o url-shortener ./cmd/main.go
}

task Test {
    Write-Host "Тесты..."
    go test -v ./internal/...
}

task Lint {
    Write-Host "Линтинг кода..."
    golangci-lint run ./...
}

task Tidy {
    Write-Host "Очистка зависимостей..."
    go mod tidy
}

task All Test, Lint, Tidy

# --- Docker задачи ---

task DockerRedisUp {
    Write-Host "Запуск Redis в Docker..."
    docker run -d --name redis-local -p 6379:6379 redis:7
}

task DockerRedisDown {
    Write-Host "Остановка Redis в Docker..."
    docker stop redis-local
    docker rm redis-local
}

task DockerUp {
    Write-Host "Запуск приложения и Redis через docker-compose..."
    docker-compose up -d
}

task DockerDown {
    Write-Host "Остановка docker-compose сервисов..."
    docker-compose down
}

# --- Тесты с поднятым Redis ---

task TestWithRedis {
    Write-Host "Тесты с поднятым Redis..."
    Invoke-Build DockerRedisUp
    Start-Sleep -Seconds 3
    try {
        go test -v ./internal/...
    } finally {
        Invoke-Build DockerRedisDown
    }
}

# --- CI задача ---

task CITest {
    Write-Host "Запуск CI pipeline: tidy, lint, тесты с Redis..."
    Invoke-Build Tidy
    Invoke-Build Lint
    Invoke-Build TestWithRedis
}
