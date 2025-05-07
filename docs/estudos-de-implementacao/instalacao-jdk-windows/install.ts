const shell = require('shelljs');

interface JavaConfig {
  version: string;
  vendor: string;
}


export async function installJdk(javaCfg: JavaConfig): Promise<void> {
  const pkgEnv = process.env.JDK_PACKAGE;
  const pkgName = pkgEnv
    ? pkgEnv
    : (javaCfg.vendor.toLowerCase().includes('openjdk')
        ? `openjdk${javaCfg.version}`
        : `jdk${javaCfg.version}`);

  const shouldInstall = process.env.CHOCOLATEY_INSTALL !== 'false';
  if (!shouldInstall) {
    console.log('Instalação via Chocolatey desabilitada. Pulando instalação do JDK.');
    return;
  }

  console.log(`> Instalando JDK via Chocolatey: ${pkgName}...`);
  const result = shell.exec(`choco install ${pkgName} -y`, { silent: false });
  if (result.code !== 0) {
    throw new Error(`Falha ao instalar JDK (código ${result.code})`);
  }
  console.log('✔ JDK instalado com sucesso.');
}


export function resolveJavaHome(javaCfg: JavaConfig): string {
  const envJavaHome = process.env.JAVA_HOME;
  if (envJavaHome) {
    return envJavaHome;
  }

  const base = javaCfg.vendor.toLowerCase().includes('openjdk')
    ? 'C:\\Program Files\\OpenJDK'
    : 'C:\\Program Files\\Java';

  const home = `${base}\\jdk-${javaCfg.version}`;
  console.log(`> Definindo JAVA_HOME para ${home}`);
  return home;
}
