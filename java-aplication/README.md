# Sistema de Gestão Empresarial

Sistema de gestão empresarial desenvolvido em Java 21 com Spring Boot.

## Requisitos

- Java 21
- Maven 3.8+
- PostgreSQL 15+

## Configuração do Ambiente

1. Clone o repositório
2. Configure o banco de dados PostgreSQL:
   - Crie um banco de dados chamado `sistema_gestao`
   - Atualize as credenciais no arquivo `application.yml` se necessário

3. Execute o projeto:
```bash
mvn spring-boot:run
```

## Documentação da API

A documentação da API está disponível em:
- Swagger UI: http://localhost:8080/swagger-ui.html
- OpenAPI: http://localhost:8080/api-docs

## Estrutura do Projeto

```
src/
├── main/
│   ├── java/
│   │   └── com/example/sistemagestao/
│   │       ├── config/         # Configurações da aplicação
│   │       ├── controller/     # Controladores REST
│   │       ├── model/          # Entidades JPA
│   │       ├── repository/     # Repositórios JPA
│   │       ├── service/        # Lógica de negócio
│   │       └── security/       # Configurações de segurança
│   └── resources/
│       └── application.yml     # Configurações da aplicação
└── test/                       # Testes unitários e de integração
```

## Funcionalidades

- Autenticação JWT
- CRUD de usuários
- Gestão de permissões
- Documentação Swagger
- Validação de dados
- Tratamento de exceções

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request 