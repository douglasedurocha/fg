# FHIR Guard CLI (fg)

Ferramenta de linha de comando para gerenciar e executar diferentes versões da aplicação FHIR Guard.

## Instalação

1. Clone o repositório
2. Instale as dependências:
   ```bash
   pip install -r requirements.txt
   ```
3. Execute a CLI:
   ```bash
   python fg.py --help
   ```

## Comandos principais

- `fg available` — Lista todas as versões disponíveis
- `fg gui` — Inicia a interface gráfica
- `fg install [versão]` — Instala uma versão específica
- `fg update` — Atualiza para a última versão
- `fg uninstall [versão]` — Remove uma versão
- `fg list` — Lista todas as versões instaladas
- `fg config [versão]` — Mostra a configuração de uma versão
- `fg start [versão]` — Inicia uma versão
- `fg stop [pid]` — Para uma instância
- `fg status` — Mostra o status das instâncias
- `fg logs [pid] [--tail n] [--follow]` — Mostra os logs de uma instância

## Observações
- Todos os comandos podem ser executados com `--help` para mais informações.
- O projeto é multiplataforma (Windows, Linux, macOS). 