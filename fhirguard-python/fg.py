import argparse
import sys
from cli import available, gui, install, update, uninstall, list as list_cmd, config, start, stop, status, logs


def main():
    parser = argparse.ArgumentParser(prog='fg', description='FHIR Guard CLI')
    subparsers = parser.add_subparsers(dest='command')

    # fg available
    subparsers.add_parser('available', help='Lista todas as versões disponíveis do FHIR Guard')

    # fg gui
    subparsers.add_parser('gui', help='Inicia a interface gráfica')

    # fg install [versão]
    install_parser = subparsers.add_parser('install', help='Instala uma versão específica do FHIR Guard')
    install_parser.add_argument('version', help='Versão para instalar')

    # fg update
    subparsers.add_parser('update', help='Atualiza para a última versão disponível')

    # fg uninstall [versão]
    uninstall_parser = subparsers.add_parser('uninstall', help='Remove uma versão específica')
    uninstall_parser.add_argument('version', help='Versão para desinstalar')

    # fg list
    subparsers.add_parser('list', help='Lista todas as versões instaladas')

    # fg config [versão]
    config_parser = subparsers.add_parser('config', help='Mostra a configuração de uma versão')
    config_parser.add_argument('version', help='Versão para mostrar configuração')

    # fg start [versão]
    start_parser = subparsers.add_parser('start', help='Inicia uma versão específica')
    start_parser.add_argument('version', help='Versão para iniciar')

    # fg stop [pid]
    stop_parser = subparsers.add_parser('stop', help='Para uma instância em execução')
    stop_parser.add_argument('pid', help='PID da instância para parar')

    # fg status
    subparsers.add_parser('status', help='Mostra o status das instâncias em execução')

    # fg logs [pid] [--tail n] [--follow]
    logs_parser = subparsers.add_parser('logs', help='Mostra os logs de uma instância')
    logs_parser.add_argument('pid', help='PID da instância')
    logs_parser.add_argument('--tail', type=int, default=None, help='Mostra as últimas n linhas do log')
    logs_parser.add_argument('--follow', action='store_true', help='Segue o log em tempo real')

    args = parser.parse_args()

    if args.command == 'available':
        available.run()
    elif args.command == 'gui':
        gui.run()
    elif args.command == 'install':
        install.run(args.version)
    elif args.command == 'update':
        update.run()
    elif args.command == 'uninstall':
        uninstall.run(args.version)
    elif args.command == 'list':
        list_cmd.run()
    elif args.command == 'config':
        config.run(args.version)
    elif args.command == 'start':
        start.run(args.version)
    elif args.command == 'stop':
        stop.run(args.pid)
    elif args.command == 'status':
        status.run()
    elif args.command == 'logs':
        logs.run(args.pid, tail=args.tail, follow=args.follow)
    else:
        parser.print_help()
        sys.exit(1)

if __name__ == '__main__':
    main() 