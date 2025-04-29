# Guia do Usuário - FHIR Guard CLI

## Introdução

FHIR Guard CLI (`fg`) é uma ferramenta de linha de comando para gerenciar instalações do FHIR Guard, uma aplicação Java para serviços FHIR (Fast Healthcare Interoperability Resources). O CLI permite instalar, iniciar, parar e monitorar diferentes versões do FHIR Guard, facilitando o gerenciamento em ambientes de desenvolvimento e produção.

## Instalação

### Pré-requisitos

- **Go 1.21+**: necessário para compilar o projeto
- **Java 11+**: necessário para executar as instâncias do FHIR Guard
- **Git**: para obter o código fonte (opcional)

### Compilação do Projeto

1. Clone o repositório (ou baixe o código fonte):
   ```
   git clone https://github.com/fhir-guard/fg.git
   cd fg
   ```

2. Baixe as dependências:
   ```
   go mod tidy
   ```

3. Compile o projeto:
   ```
   go build -o bin/fg.exe main.go
   ```

4. (Opcional) Adicione o diretório `bin` ao seu PATH para executar o comando de qualquer local.

## Como Utilizar

Para utilizar o FHIR Guard CLI, navegue até a pasta `bin` e execute o comando `fg.exe` seguido do subcomando desejado:

```
cd bin
fg.exe [comando]
```

### Comandos Disponíveis

| Comando | Descrição | Exemplo de Uso |
|---------|-----------|----------------|
| `version` | Exibe a versão atual do CLI | `fg.exe version` |
| `config` | Gerencia configurações do sistema | `fg.exe config --get server.port` |
| `list` | Lista versões instaladas ou disponíveis | `fg.exe list --remote` |
| `install` | Instala uma versão específica | `fg.exe install 1.2.3` |
| `start` | Inicia uma instância | `fg.exe start 1.2.3` |
| `stop` | Para instâncias em execução | `fg.exe stop 1.2.3` |
| `status` | Exibe o status das instâncias ativas | `fg.exe status` |
| `logs` | Exibe logs das instâncias | `fg.exe logs 1.2.3` |
| `update` | Atualiza para uma nova versão | `fg.exe update 1.2.3` |

### Fluxo de Trabalho Típico

1. **Listar versões disponíveis**:
   ```
   fg.exe list --remote
   ```

2. **Instalar uma versão**:
   ```
   fg.exe install 1.2.3
   ```

3. **Iniciar a instância**:
   ```
   fg.exe start 1.2.3
   ```

4. **Verificar o status**:
   ```
   fg.exe status
   ```

5. **Visualizar logs**:
   ```
   fg.exe logs 1.2.3
   ```

6. **Parar a instância**:
   ```
   fg.exe stop 1.2.3
   ```

## Detalhes da Implementação

### Estrutura do Projeto

- `cmd/`: Contém os subcomandos da CLI
  - `install.go`: Comandos para instalar versões
  - `start.go`: Inicia instâncias do FHIR Guard
  - `stop.go`: Para instâncias em execução
  - `status.go`: Exibe status de instâncias
  - `logs.go`: Acessa os logs das instâncias
  - `list.go`: Lista versões disponíveis
  - `update.go`: Atualiza para novas versões
  - `config.go`: Gerencia configurações
  - `root.go`: Comando raiz e configuração base
  - `version.go`: Exibe informações de versão

- `config/`: Definições de configuração
  - `config.go`: Estrutura e funções de configuração

- `main.go`: Ponto de entrada da aplicação

### Diretórios e Arquivos

O FHIR Guard CLI cria uma estrutura de diretórios no computador:

```
$HOME/.fhir-guard/
├── config/
│   └── config.yaml     # Configurações do CLI
├── logs/
│   ├── fg.log          # Logs do próprio CLI
│   └── <versão>/       # Logs de cada versão
├── versions/
│   └── <versão>/       # Instalações de cada versão
```

## Configuração

O arquivo de configuração principal está localizado em:
- Windows: `C:\Users\<usuário>\.fhir-guard\config\config.yaml`
- Linux/macOS: `$HOME/.fhir-guard/config/config.yaml`

### Configurações Principais

| Configuração | Descrição | Exemplo |
|--------------|-----------|---------|
| `server.port` | Porta HTTP padrão | `8080` |
| `server.host` | Host de escuta padrão | `localhost` |
| `server.maxMemory` | Memória máxima para a JVM | `1g` |
| `java.jvmArgs` | Argumentos JVM padrão | `["-Xms256m", "-Xmx1g"]` |
| `downloadUrl` | URL base para downloads | `https://releases.fhir-guard.org` |

### Modificar Configurações

Para visualizar uma configuração:
```
fg.exe config --get server.port
```

Para modificar uma configuração:
```
fg.exe config --set server.port=8081
```

Para exibir o arquivo de configuração completo:
```
fg.exe config
```

## Gerenciamento de Versões

### Instalar uma Versão

```
fg.exe install 1.2.3
```

Opções:
- `--force` (`-f`): Força reinstalação mesmo se já existir
- `--skip-deps`: Ignora download de dependências

### Iniciar uma Versão

```
fg.exe start 1.2.3
```

Opções:
- `--port` (`-p`): Especifica a porta HTTP
- `--host`: Especifica o host de escuta
- `--jvm-args`: Argumentos JVM adicionais
- `--background` (`-b`): Executa em segundo plano

### Verificar Status

```
fg.exe status
```

Exibe todas as instâncias em execução, seus PIDs, versões, portas, uso de CPU/memória e tempo de execução.

### Visualizar Logs

```
fg.exe logs 1.2.3
```

Opções:
- `--follow` (`-f`): Acompanha em tempo real
- `--lines` (`-n`): Número de linhas a exibir
- `--port` (`-p`): Filtra por porta específica

### Parar Instâncias

```
fg.exe stop 1.2.3
```

Opções:
- `--port` (`-p`): Para apenas a instância na porta especificada
- `--force` (`-f`): Encerramento imediato (SIGKILL)
- `--timeout` (`-t`): Tempo de espera antes de matar

## Solução de Problemas

### Problemas Comuns

1. **Erro "versão não encontrada"**:
   - Verifique sua conexão de internet
   - Verifique a URL de download com `fg.exe config --get downloadUrl`

2. **Erro ao iniciar uma versão**:
   - Verifique se o Java está instalado e acessível
   - Verifique se a porta não está em uso

3. **Logs não aparecem**:
   - Verifique o diretório de logs em `$HOME/.fhir-guard/logs/<versão>/`

### Verificação do Ambiente

```
fg.exe version        # Versão do CLI
java -version         # Versão do Java
fg.exe config --file  # Localização do arquivo de configuração
```

## Exemplos Práticos

### Configurar e Iniciar em uma Nova Porta

```
fg.exe config --set server.port=8090
fg.exe install 1.2.3
fg.exe start 1.2.3
```

### Atualizar para uma Nova Versão

```
fg.exe list --remote
fg.exe update 1.3.0
fg.exe stop 1.2.3
fg.exe start 1.3.0
```

### Executar em Segundo Plano

```
fg.exe start 1.2.3 --background
fg.exe status
fg.exe logs 1.2.3 --follow
```

## Considerações de Segurança

- Certifique-se de que o FHIR Guard esteja executando em um ambiente seguro
- Ao expor para redes externas, configure o host apropriadamente
- Considere usar HTTPS para comunicação segura

## Desinstalação

Para remover completamente o FHIR Guard CLI e todos os seus dados:

1. Pare todas as instâncias em execução:
   ```
   fg.exe stop
   ```

2. Remova o diretório `.fhir-guard`:
   ```
   # Windows
   rmdir /s /q %USERPROFILE%\.fhir-guard

   # Linux/macOS
   rm -rf ~/.fhir-guard
   ```