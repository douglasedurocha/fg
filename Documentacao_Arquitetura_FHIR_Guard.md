# Documentação de Arquitetura - FHIR Guard

## 1. Visão Geral do Sistema

O FHIR Guard é uma aplicação para gerenciamento e execução de serviços relacionados ao padrão FHIR (Fast Healthcare Interoperability Resources), um padrão para troca de informações em saúde. A ferramenta "fg" (FHIR Guard CLI) é um utilitário de linha de comando que facilita a instalação, configuração e gerenciamento de instâncias do FHIR Guard.

### 1.1. Diagrama de Arquitetura

```
+---------------------------------------------------------------+
|                      FHIR Guard CLI (fg)                      |
+----------------------+-------------------+--------------------+
        |                      |                      |
        v                      v                      v
+----------------+    +----------------+    +------------------+
| Módulo Config  |    |  Módulo Cmd    |    | Serviços FHIR    |
| (config.go)    |<-->| (comandos CLI) |<-->| (instâncias JAR) |
+----------------+    +----------------+    +------------------+
        |                      |                      |
        v                      v                      v
+----------------+    +----------------+    +------------------+
|  config.yaml   |    |   Operações    |    |    Processos     |
| (configurações)|    |(install/start) |    |     Java/JVM     |
+----------------+    +----------------+    +------------------+
        |                      |                      |
        v                      v                      v
+-----------------------------------------------------------+
|                      Sistema de Arquivos                   |
|  ($FG_HOME: configs, versões, logs, processos ativos)      |
+-----------------------------------------------------------+
```

## 2. Objetivos do Sistema

- Permitir a instalação e gerenciamento de múltiplas versões do FHIR Guard
- Facilitar o início e parada de instâncias do FHIR Guard
- Monitorar instâncias em execução
- Configurar parâmetros de execução como porta, host e argumentos JVM
- Gerenciar logs de forma centralizada

## 3. Componentes Arquiteturais

### 3.1. Estrutura de Diretórios

O sistema utiliza a seguinte estrutura de diretórios:

```
$FG_HOME/
├── config/
│   └── config.yaml
├── versions/
│   └── {versão}/
│       ├── fhir-guard-{versão}.jar
│       ├── deps/
│       └── config/
├── logs/
│   ├── fg.log
│   └── {versão}/
└── active_pids.json
```

- **$FG_HOME**: Diretório principal do FHIR Guard, normalmente localizado em `~/.fhir-guard`
- **config/**: Armazena configurações globais
- **versions/**: Contém os arquivos JAR e dependências de cada versão instalada
- **logs/**: Armazena logs do sistema e de cada instância em execução
- **active_pids.json**: Armazena informações sobre processos em execução

### 3.2. Módulos Principais

#### 3.2.1. Módulo de Configuração (`config`)

Este módulo gerencia as configurações globais do sistema, incluindo:
- Configurações da JVM
- Configurações de servidor (porta, host)
- Informações de versões instaladas
- Processos ativos

#### 3.2.2. Módulo de Comando (`cmd`)

Implementa os comandos disponíveis na CLI:
- **root**: Configuração principal da CLI
- **install**: Instalação de versões específicas do FHIR Guard
- **start**: Inicialização de instâncias
- **stop**: Parada de instâncias em execução
- **status**: Monitoramento de instâncias ativas
- **list**: Listagem de versões instaladas
- **logs**: Visualização de logs
- **update**: Atualização de versões
- **config**: Configuração do sistema
- **version**: Exibição da versão da CLI

## 4. Fluxos de Operação

### 4.1. Instalação de uma Versão

1. O comando `fg install` é acionado com uma versão específica
2. O sistema valida o formato da versão
3. As informações da versão são obtidas remotamente
4. O arquivo JAR principal é baixado e verificado (checksum)
5. As dependências são baixadas (se necessário)
6. As configurações padrão são criadas
7. As informações da versão são atualizadas no arquivo de configuração

#### 4.1.1. Diagrama de Fluxo de Instalação

```
+-------------+     +---------------+     +----------------+
| Início:     |     | Validação     |     | Busca info     |
| fg install  +---->+ da versão     +---->+ da versão      |
| <versão>    |     | (formato)     |     | (metadata.json)|
+-------------+     +---------------+     +----------------+
                                                 |
+----------------+     +------------------+      |
| Atualização    |     | Criação de       |      |
| de config      +<----+ configs padrão   |<-----+
| (config.yaml)  |     | (se necessário)  |      |
+----------------+     +------------------+      |
      ^                                          |
      |                                          v
+----------------+     +------------------+     +----------------+
| Download       |     | Verificação      |     | Download do    |
| dependências   +<----+ de checksum      |<----+ arquivo JAR    |
| (se necessário)|     | (SHA-256)        |     | principal      |
+----------------+     +------------------+     +----------------+
```

### 4.2. Inicialização de uma Instância

1. O comando `fg start` é acionado com uma versão específica
2. O sistema verifica se a versão está instalada
3. Configurações como porta e host são definidas
4. O sistema verifica se não há conflito com instâncias em execução
5. O processo Java é iniciado com os argumentos apropriados
6. As informações do processo são registradas no arquivo de PIDs ativos
7. Os logs são redirecionados para o arquivo apropriado

#### 4.2.1. Diagrama de Fluxo de Inicialização

```
+-------------+     +---------------+     +----------------+
| Início:     |     | Verificação   |     | Configuração   |
| fg start    +---->+ da versão     +---->+ de parâmetros  |
| <versão>    |     | (instalada?)  |     | (porta, host)  |
+-------------+     +---------------+     +----------------+
                                                 |
+----------------+     +------------------+      |
| Registro do    |     | Inicialização    |      |
| processo ativo +<----+ do processo Java |<-----+
| (PID)          |     | (JVM, args)      |      |
+----------------+     +------------------+      |
      ^                                          |
      |                                          v
+----------------+     +------------------+     +----------------+
| Monitoramento  |     | Redirecionamento |     | Verificação de |
| do processo    +<----+ de logs          |<----+ conflitos      |
| (background)   |     | (arquivo)        |     | (portas em uso)|
+----------------+     +------------------+     +----------------+
```

### 4.3. Monitoramento de Instâncias

1. O comando `fg status` é acionado
2. O sistema lê o arquivo de PIDs ativos
3. Cada processo é verificado para confirmar se ainda está em execução
4. Informações como PID, versão, porta e tempo de execução são exibidas

## 5. Aspectos Técnicos

### 5.1. Tecnologias Utilizadas

- **Linguagem**: Go (Golang)
- **Frameworks**:
  - Cobra: Framework para criação de aplicações CLI
  - Viper: Gerenciamento de configurações
  - Logrus: Sistema de logging
  - Retryable-HTTP: Cliente HTTP com capacidade de retry
  - YAML/JSON: Formatos de configuração e metadados

### 5.2. Gerenciamento de Processos

O sistema utiliza o pacote `os/exec` para criar e gerenciar processos externos (JVM). O módulo `gopsutil` é usado para verificar o estado de processos em execução.

### 5.3. Logging

- **Centralizado**: Todos os logs da CLI são armazenados em `$FG_HOME/logs/fg.log`
- **Por Instância**: Cada instância FHIR Guard possui seus próprios logs em `$FG_HOME/logs/{versão}/`
- **Formato**: Os logs utilizam o formato JSON para facilitar a análise

### 5.4. Configuração

- O arquivo principal de configuração é armazenado em YAML
- Configurações podem ser sobrescritas via linha de comando
- Variáveis de ambiente são suportadas (ex: `FG_HOME`)

## 6. Segurança e Resiliência

### 6.1. Verificação de Integridade

- Checksums SHA-256 são utilizados para verificar a integridade dos arquivos baixados
- Conexões HTTP utilizam mecanismos de retry para lidar com falhas temporárias de rede

### 6.2. Isolamento de Versões

- Cada versão do FHIR Guard é instalada em seu próprio diretório
- Configurações e dependências são isoladas por versão

## 7. Extensibilidade

O sistema foi projetado com extensibilidade em mente:

- Novos comandos podem ser facilmente adicionados ao módulo `cmd`
- O formato de configuração permite a adição de novos parâmetros
- O sistema de versões suporta metadados específicos por versão

## 8. Limitações e Considerações

- O sistema depende da JVM para execução do FHIR Guard
- Requer privilégios suficientes para gerenciar processos no sistema operacional
- O monitoramento de processos é limitado às capacidades do sistema operacional

## 9. Conclusão

O FHIR Guard CLI é uma ferramenta poderosa para gerenciar instâncias FHIR Guard, oferecendo uma interface de linha de comando intuitiva e recursos abrangentes para instalação, configuração e monitoramento. Sua arquitetura modular permite extensibilidade e manutenção simplificada. 