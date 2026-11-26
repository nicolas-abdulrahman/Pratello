# Pratello

O **Pratello** é uma plataforma de menu digital para múltiplos estabelecimentos (*multi-tenant*), totalmente customizável para atender à identidade e regras de negócio de cada restaurante, pizzaria ou lanchonete. Ele permite gerenciar cardápios, itens, pedidos e apresentação visual de forma independente por tenant, proporcionando uma experiência rápida, moderna e intuitiva tanto para os clientes quanto para a administração.

A aplicação conta com um backend robusto desenvolvido em **Go** (utilizando `net/http`), um frontend reativo em **Svelte** com **Tailwind CSS** e **Vite**, além de persistência em **PostgreSQL** e suporte a cache e sessões com **Redis**.

---

## 📁 Estrutura do Projeto

```text
pratello/
├── flake.nix                  # Definição do ambiente de desenvolvimento Nix
├── go.mod                     # Módulo Go 1.22
├── package.json               # Scripts do monorepo (server, dev, test_e2e, test_backend)
├── README.md                  # Documentação do projeto
├── docs/                      # Especificações OpenAPI, casos de teste e arquitetura
│   ├── Casos de teste.md      # Catálogo e descrição de casos de teste
│   ├── openapi.yaml           # Especificação OpenAPI 3.0 dos endpoints
│   └── diagrams/              # Diagramas de fluxo e arquitetura
├── e2e/                       # Testes ponta a ponta (E2E) com Playwright
│   ├── package.json           # Dependências do Playwright (@playwright/test)
│   ├── playwright.config.cjs  # Configuração do Playwright (servidor e browser)
│   └── module-7.spec.cjs      # Especificações e cenários dos testes E2E
├── frontend/                  # Aplicação cliente em Svelte + Vite + Tailwind CSS
│   ├── package.json           # Dependências do frontend
│   ├── vite.config.js         # Configuração do Vite com proxy da API
│   └── src/
│       ├── App.svelte         # Componente raiz e roteamento SPA
│       ├── admin/             # Painel administrativo, gestão de cardápio e equipe
│       ├── auth/              # Telas e fluxos de autenticação
│       ├── common/            # Componentes reutilizáveis (Navbar, Footer, etc.)
│       ├── delivery/          # Cálculo e gerenciamento de entrega
│       ├── marketplace/       # Vitrine de restaurantes e geolocalização
│       ├── menu/              # Exibição de cardápios, categorias e pratos
│       ├── order/             # Carrinho, checkout e histórico de pedidos
│       ├── payment/           # Fluxos e páginas de pagamento
│       ├── restaurant/        # Páginas dinâmicas customizáveis por restaurante
│       ├── review/            # Avaliações e notas dos clientes
│       └── user/              # Perfil de usuário e restaurantes do proprietário
├── scripts/                   # Scripts auxiliares de desenvolvimento e testes
│   ├── start-dev-redis.mjs    # Inicializador do container Redis de desenvolvimento
│   └── test-backend-cases.mjs # Executor dos casos de teste do backend
└── server/                    # Backend RESTful em Go (net/http)
    ├── cmd/api/               # Ponto de entrada do servidor (main.go)
    └── internal/              # Módulos de domínio e regras de negócio
        ├── auth/              # Autenticação, tokens e sessões
        ├── common/            # Middlewares, utilitários e validações comuns
        ├── db/                # Conexão PostgreSQL, migrações e fixtures
        ├── integration/       # Testes de integração de fluxos completos
        ├── media/             # Armazenamento e entrega de mídias e imagens
        ├── menu/              # Gestão de cardápios, categorias e pratos
        ├── order/             # Gestão de ciclo de vida e status de pedidos
        ├── restaurants/       # Lógica multi-tenant de restaurantes
        ├── review/            # Avaliações e comentários
        ├── sales/             # Relatórios e estatísticas de vendas
        ├── transaction/       # Processamento e registro de transações
        └── user/              # Gestão de contas e perfis de usuários
```

---

## 🛠️ Ambiente de Desenvolvimento

Para entrar no ambiente com todas as ferramentas e variáveis necessárias configuradas, execute:

```bash
nix shell .
```

### Dependências do Projeto

O projeto utiliza as seguintes dependências e ferramentas (providas no ambiente):

- **Go** (1.22+)
- **Node.js** (v22) & **npm**
- **PostgreSQL**
- **Docker** & **Docker Compose**
- **Git**

---

## 🚀 Como Executar o Projeto

### 1. Iniciar o Backend

Para iniciar o servidor backend com banco de dados e Redis de desenvolvimento (carregando a base de demonstração do *João Burger*):

```bash
npm run server
```

> **Nota:** O comando acima inicia/reutiliza o container Redis local (`pratello-dev-redis`) na porta `6379` e roda a API Go com `AUTH_FIXED_CODE=1`. Para iniciar diretamente sem fixtures:
> ```bash
> go run ./server/cmd/api
> ```

### 2. Iniciar o Frontend

Em outro terminal, execute o servidor de desenvolvimento do frontend (com *Hot Module Replacement*):

```bash
npm run dev
```

O frontend estará disponível em `http://localhost:5173`, com proxy automático para as rotas da API em `http://localhost:8080`.

---

## 🧪 Comandos de Testes

### Testes do Backend (Go)

Para executar os testes do backend:

```bash
npm run test_backend
```

Ou diretamente via Go com as variáveis de ambiente:

```bash
AUTH_TEST_DATABASE_URL="postgres://nick@localhost:5432/pratello?sslmode=disable" AUTH_TEST_REDIS_URL="redis://localhost:6379/0" go test ./server/...
```
*(Para ver a saída detalhada de cada teste, adicione `-v -count=1`).*

### Testes End-to-End (E2E) com Playwright

Para rodar os testes ponta a ponta:

```bash
npm run test_e2e
# ou simplesmente:
npm test
```
*(O comando e parâmetros específicos podem ser consultados no `package.json`).*

---

## 🌐 Visão Geral de Rotas

| Rota | Descrição |
| :--- | :--- |
| `/` | Página inicial do Pratello |
| `/{restaurante}` | Visualização dinâmica do cardápio do estabelecimento (ex: `/bobpizzeria`, `/deliciadeprazeres`) |
| `/api/health` | Verificação de integridade (*health check*) |
| `/api/restaurants/{restaurant}` | Dados do restaurante e cardápio servidos pela API Go |

---

## 📖 Documentação da API

Para rodar e visualizar a documentação interativa da API:

```bash
npm run docs
```

Acesse em seu navegador: <http://localhost:8081>
*(O comando específico executado pode ser consultado no `package.json`).*
