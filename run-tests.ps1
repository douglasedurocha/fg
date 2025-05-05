# Script para executar testes específicos no Windows
# Uso: .\run-tests.ps1 -Package "cmd" -Test "TestIsJavaVersionValid"

param (
    [string]$Package = "cmd",
    [string]$Test = "",
    [int]$Timeout = 30
)

Write-Host "Executando teste: $Test no pacote $Package com timeout de ${Timeout}s"

if ($Test -eq "") {
    # Execute todos os testes no pacote
    go test -v -timeout "${Timeout}s" "./$Package/..."
} else {
    # Execute apenas o teste específico
    go test -v -timeout "${Timeout}s" "./$Package/..." -run="^$Test$"
}

Write-Host "Teste concluído" 