# FHIR Guard - Interface Desktop

## Visão Geral
FHIR Guard é uma aplicação desktop desenvolvida em Go utilizando o framework Fyne para interface gráfica. Esta aplicação permite gerenciar e monitorar recursos FHIR de forma segura e eficiente.

## Requisitos
- Go 1.21 ou superior
- Dependências do Fyne (serão instaladas automaticamente via `go mod tidy`)

## Instalação
1. Clone o repositório
2. Navegue até o diretório `desktop`
3. Execute `go mod tidy` para instalar as dependências
4. Execute `go run main.go` para iniciar a aplicação

## Estrutura da Interface
A interface atual inclui:
- Uma janela principal com título "FHIR Guard"
- Um rótulo de boas-vindas
- Um botão de demonstração

## Funcionalidades Atuais
- Exibição de mensagem de boas-vindas
- Botão interativo que altera o texto do rótulo ao ser clicado
- Interface responsiva com tamanho de janela configurável

## Desenvolvimento
Para contribuir com o desenvolvimento:
1. Certifique-se de ter todas as dependências instaladas
2. Faça suas alterações no código
3. Teste localmente com `go run main.go`
4. Envie suas alterações para revisão

## Próximos Passos
- Implementação de funcionalidades específicas do FHIR
- Adição de menus e navegação
- Integração com serviços backend
- Melhorias na interface do usuário

## Suporte
Para suporte ou dúvidas, entre em contato com a equipe de desenvolvimento. 