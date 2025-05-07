import os
import json
import sys
from colorama import Fore, Style
import requests

REPO_BASE_URL = "https://raw.githubusercontent.com/allan-bispo/fg/main"

FG_HOME = os.environ.get('FG_HOME', os.path.expanduser('~/.fg'))

MAVEN_CENTRAL = 'https://repo1.maven.org/maven2'


OS_TYPE = 'windows'  
ARCH = 'x64'       


def run(version):
    try:
        meta_path = f'java-aplication/src/dependencies/version-{version}.json'
        remote_url = f"{REPO_BASE_URL}/java-aplication/src/dependencies/version-{version}.json"
        if not os.path.exists(meta_path):
         
            print(Fore.YELLOW + f"Arquivo de metadados não encontrado localmente. Baixando de {remote_url}..." + Style.RESET_ALL)
            r = requests.get(remote_url)
            if r.status_code == 200:
                os.makedirs(os.path.dirname(meta_path), exist_ok=True)
                with open(meta_path, 'w', encoding='utf-8') as f:
                    f.write(r.text)
                print(Fore.GREEN + f"Metadados salvos em {meta_path}" + Style.RESET_ALL)
            else:
                print(Fore.RED + f"Não foi possível baixar o arquivo de metadados: {remote_url}" + Style.RESET_ALL)
                sys.exit(1)
        with open(meta_path, 'r', encoding='utf-8') as f:
            meta = json.load(f)
        print(Fore.GREEN + f'Metadados carregados para a versão {version}.' + Style.RESET_ALL)

        version_dir = os.path.join(FG_HOME, version)
        os.makedirs(version_dir, exist_ok=True)
        libs_dir = os.path.join(version_dir, 'libs')
        os.makedirs(libs_dir, exist_ok=True)
        jdk_dir = os.path.join(version_dir, 'jdk')
        os.makedirs(jdk_dir, exist_ok=True)

        
        for dep_group in meta.get('dependencies', {}):
            dep = meta['dependencies'][dep_group]
            if isinstance(dep, dict) and 'dependencies' in dep:
                for d in dep['dependencies']:
                    if 'url' in d:
                        url = d['url']
                    else:
                        url = maven_url(d['groupId'], d['artifactId'], d['version'])
                    local_name = f"{d['artifactId']}-{d['version']}.jar"
                    dest = os.path.join(libs_dir, local_name)
                    download_file(url, dest)
            elif isinstance(dep, dict) and 'groupId' in dep:
                if 'url' in dep:
                    url = dep['url']
                else:
                    url = maven_url(dep['groupId'], dep['artifactId'], dep['version'])
                local_name = f"{dep['artifactId']}-{dep['version']}.jar"
                dest = os.path.join(libs_dir, local_name)
                download_file(url, dest)

    
        jdk_version = str(meta.get('java', {}).get('version', '21'))
        jdk_url = adoptium_url(jdk_version, OS_TYPE, ARCH)
        jdk_zip = os.path.join(jdk_dir, f'jdk-{jdk_version}.zip')
        download_file(jdk_url, jdk_zip)

  
        config_path = os.path.join(version_dir, 'config.yaml')
        with open(config_path, 'w', encoding='utf-8') as f:
            f.write(f'# Configuração padrão para FHIR Guard {version}\n')
        print(Fore.GREEN + f'Versão {version} instalada com sucesso em {version_dir}.' + Style.RESET_ALL)
    except Exception as e:
        print(Fore.RED + f'Falha ao instalar versão {version}: {e}' + Style.RESET_ALL)
        sys.exit(1)

def maven_url(group_id, artifact_id, version):
    path = f"{group_id.replace('.', '/')}/{artifact_id}/{version}/{artifact_id}-{version}.jar"
    return f"{MAVEN_CENTRAL}/{path}"

def adoptium_url(jdk_version, os_type, arch):
    return f"https://api.adoptium.net/v3/binary/latest/{jdk_version}/ga/{os_type}/{arch}/jdk/hotspot/normal/eclipse"

def download_file(url, dest):
    try:
        print(f'Baixando {url}...')
        with requests.get(url, stream=True) as r:
            r.raise_for_status()
            total = int(r.headers.get('content-length', 0))
            with open(dest, 'wb') as f:
                downloaded = 0
                for chunk in r.iter_content(chunk_size=8192):
                    if chunk:
                        f.write(chunk)
                        downloaded += len(chunk)
                        if total:
                            done = int(50 * downloaded / total)
                            sys.stdout.write(f'\r[{"="*done}{{" "*(50-done)}}] {downloaded//1024}KB/{total//1024}KB')
                            sys.stdout.flush()
        print(f'\nSalvo em {dest}')
    except Exception as e:
        print(Fore.RED + f'Erro ao baixar {url}: {e}' + Style.RESET_ALL)
        sys.exit(1) 