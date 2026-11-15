# Pratello API

Base URL: `http://localhost:8080/api`

## Endpoints

| Method | Route | Body | Description |
| :--- | :--- | :--- | :--- |
| **Autenticação & Contas** | | | |
| `POST` | `/auth/register` | `{ email, password, phone }` | Criar nova conta global (`account`) |
| `POST` | `/auth/login` | `{ email, password }` | Autenticar usuário e gerar sessão/token JWT |
| `GET` | `/auth/me` | — | Obter dados da conta autenticada |
| `GET` | `/sysadmins` | — | Listar superadministradores globais (`sysadmin`) |
| `POST` | `/sysadmins` | `{ account_id }` | Promover conta a superadministrador |
| `GET` | `/restaurants/{slug}/owners` | — | Listar proprietários da entidade lógica (`OWNER`) |
| `POST` | `/restaurants/{slug}/owners` | `{ account_id }` | Promover/vincular conta como proprietário (`OWNER`) |
| `DELETE` | `/restaurants/{slug}/owners/{account_id}` | — | Revogar proprietário (protegido contra orfandade da entidade) |
| `GET` | `/restaurants/{slug}/managers` | — | Listar gerentes operacionais do restaurante (`MANAGER`) |
| `POST` | `/restaurants/{slug}/managers` | `{ account_id }` | Vincular conta como gerente operacional (`MANAGER`) |
| `DELETE` | `/restaurants/{slug}/managers/{account_id}` | — | Desvincular gerente operacional do restaurante |
| `GET` | `/users` | — | Listar consumidores finais (`users`) |
| `GET` | `/users/{cpf}` | — | Obter dados do cliente por CPF |
| `POST` | `/users` | `{ cpf, name, phone }` | Cadastrar novo cliente consumidor |
| `PUT` | `/users/{cpf}` | `{ name, phone }` | Atualizar dados cadastrais do cliente |
| **Estabelecimentos & Unidades Físicas** | | | |
| `GET` | `/restaurants` | — | Listar marcas/estabelecimentos (`restaurant`) |
| `GET` | `/restaurants/{slug}` | — | Obter detalhes da marca por slug |
| `POST` | `/restaurants` | `{ name, slug, description }` | Cadastrar nova marca/restaurante |
| `PUT` | `/restaurants/{slug}` | `{ name, description }` | Atualizar dados da marca |
| `DELETE` | `/restaurants/{slug}` | — | Remover marca/restaurante |
| `GET` | `/restaurants/{slug}/branches` | — | Listar filiais físicas (`physical_restaurant`) |
| `GET` | `/restaurants/{slug}/branches/{branch_id}` | — | Obter dados da filial física |
| `POST` | `/restaurants/{slug}/branches` | `{ name, cnpj, owner_cnpj, address: { cep, street, number, apt, complement } }` | Criar filial física validando CNPJ da filial e do proprietário |
| `PUT` | `/restaurants/{slug}/branches/{branch_id}` | `{ name }` | Atualizar informações da filial |
| `DELETE` | `/restaurants/{slug}/branches/{branch_id}` | — | Desativar filial física |
| `PUT` | `/restaurants/{slug}/branches/{branch_id}/address` | `{ cep, street, number, apt, complement }` | Atualizar endereço da filial (`Address`) |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/phones` | — | Listar telefones da filial (`p_restaurant_phone`) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/phones` | `{ phone_number, phone_type }` | Adicionar telefone à filial |
| `DELETE` | `/restaurants/{slug}/branches/{branch_id}/phones/{phone_id}` | — | Remover telefone da filial |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/employees` | — | Listar registros de RH de colaboradores da filial (`employee`) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/employees` | `{ name, cpf, position }` | Cadastrar colaborador no RH (registro prévio sem criação de conta) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/employees/{employee_id}/invite` | `{ account_id }` | Vincular conta global existente ao colaborador via convite |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/terminal/session` | `{ pin, device_id }` | Abrir sessão de terminal de salão via PIN de 6 dígitos (expira em 8h, bloqueio por 5 falhas) |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/terminal/orders` | — | Visualização restrita de pedidos para terminal de salão (`X-Device-Id` + Token) |
| **Catálogo, Cardápio & Itens** | | | |
| `GET` | `/restaurants/{slug}/menu` | — | Obter layout e identidade visual do cardápio (`menu`) |
| `PUT` | `/restaurants/{slug}/menu` | `{ name, template_style, accent_color, typography }` | Atualizar tema/layout do cardápio |
| `GET` | `/restaurants/{slug}/categories` | — | Listar categorias do cardápio (`category`) |
| `POST` | `/restaurants/{slug}/categories` | `{ name, display_order }` | Criar categoria no cardápio |
| `PUT` | `/restaurants/{slug}/categories/{id}` | `{ name, display_order }` | Atualizar categoria |
| `DELETE` | `/restaurants/{slug}/categories/{id}` | — | Remover categoria |
| `GET` | `/restaurants/{slug}/items` | — | Listar itens vendáveis do catálogo (`menu_item`) |
| `GET` | `/restaurants/{slug}/items/{id}` | — | Obter item por ID com ingredientes e opcionais |
| `POST` | `/restaurants/{slug}/items` | `{ category_id, name, description, base_price, photo_url, is_combo, is_available }` | Criar item no catálogo |
| `PUT` | `/restaurants/{slug}/items/{id}` | `{ category_id, name, description, base_price, photo_url, is_combo, is_available }` | Atualizar item do catálogo |
| `DELETE` | `/restaurants/{slug}/items/{id}` | — | Remover item do catálogo |
| `GET` | `/restaurants/{slug}/ingredients` | — | Listar ingredientes cadastrados (`Ingredient_item`) |
| `POST` | `/restaurants/{slug}/ingredients` | `{ name }` | Cadastrar novo ingrediente base |
| `DELETE` | `/restaurants/{slug}/ingredients/{id}` | — | Remover ingrediente base |
| `GET` | `/restaurants/{slug}/items/{item_id}/ingredients` | — | Listar ingredientes do item (`item_ingredient`) |
| `POST` | `/restaurants/{slug}/items/{item_id}/ingredients` | `{ ingredient_id, proportion, is_removable }` | Associar ingrediente com regras de remoção |
| `DELETE` | `/restaurants/{slug}/items/{item_id}/ingredients/{ingredient_id}` | — | Desvincular ingrediente do item |
| `GET` | `/restaurants/{slug}/items/{item_id}/sides` | — | Listar grupos de opcionais do item (`item_side`) |
| `POST` | `/restaurants/{slug}/items/{item_id}/sides` | `{ title, min_options, max_options }` | Criar grupo de adicionais (ex.: "Ponto da carne") |
| `PUT` | `/restaurants/{slug}/items/{item_id}/sides/{side_id}` | `{ title, min_options, max_options }` | Atualizar grupo de adicionais |
| `DELETE` | `/restaurants/{slug}/items/{item_id}/sides/{side_id}` | — | Remover grupo de adicionais |
| `POST` | `/restaurants/{slug}/items/{item_id}/sides/{side_id}/options` | `{ name, additional_price }` | Adicionar opção ao grupo (`side_option`) |
| `PUT` | `/restaurants/{slug}/sides/options/{option_id}` | `{ name, additional_price }` | Atualizar opção de adicional |
| `DELETE` | `/restaurants/{slug}/sides/options/{option_id}` | — | Remover opção de adicional |
| `GET` | `/restaurants/{slug}/items/{item_id}/combo-items` | — | Listar composição do combo (`combo_item`) |
| `POST` | `/restaurants/{slug}/items/{item_id}/combo-items` | `{ sub_item_id, quantity }` | Adicionar produto unitário à receita do combo |
| `DELETE` | `/restaurants/{slug}/items/{item_id}/combo-items/{sub_item_id}` | — | Remover item da composição do combo |
| `GET` | `/restaurants/{slug}/items/{item_id}/sales` | — | Listar promoções ativas do item (`item_sale`) |
| `POST` | `/restaurants/{slug}/items/{item_id}/sales` | `{ discount_type, discount_value, starts_at, ends_at }` | Aplicar promoção/desconto temporário ao item |
| `DELETE` | `/restaurants/{slug}/items/{item_id}/sales/{sale_id}` | — | Encerrar promoção do item |
| **Fidelidade & Cupons** | | | |
| `GET` | `/restaurants/{slug}/users/{cpf}/wallet` | — | Consultar saldo de pontos/coins do cliente (`user_restaurant`) |
| `POST` | `/restaurants/{slug}/users/{cpf}/wallet/adjust` | `{ points_delta, reason }` | Creditar ou debitar saldo de fidelidade |
| `GET` | `/restaurants/{slug}/coupons` | — | Listar cupons de desconto (`coupon`) |
| `GET` | `/restaurants/{slug}/coupons/{code}` | — | Validar cupom por código |
| `POST` | `/restaurants/{slug}/coupons` | `{ code, discount_type, discount_value, min_order_value, expires_at, max_uses }` | Criar cupom promocional |
| `DELETE` | `/restaurants/{slug}/coupons/{id}` | — | Desativar cupom promocional |
| **Vendas & Pedidos (Salão & Delivery)** | | | |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/orders` | — | Listar comandas e pedidos da filial (`order`) |
| `GET` | `/orders/{order_id}` | — | Obter detalhes completos do pedido |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/orders/local` | `{ user_cpf, table_number, items: [{ item_id, quantity, notes, sides }] }` | Abrir pedido presencial no salão (`local_order` + `order_item`) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/orders/remote` | `{ user_cpf, delivery_address, phone, freight_fee, items: [...] }` | Criar pedido delivery (`remote_order` + `order_item`) |
| `PATCH` | `/orders/{order_id}/status` | `{ status }` | Atualizar status do pedido (recebido, preparo, pronto, entregue) |
| `POST` | `/orders/{order_id}/items` | `{ item_id, quantity, notes, sides, parent_order_item_id }` | Adicionar item a um pedido aberto |
| `DELETE` | `/orders/{order_id}/items/{order_item_id}` | — | Cancelar/remover item do pedido |
| **Pagamentos & Faturamento** | | | |
| `POST` | `/orders/{order_id}/transactions` | `{ payment_method, amount }` | Iniciar pagamento do pedido (`transaction`) |
| `GET` | `/orders/{order_id}/transactions` | — | Consultar transações de pagamento do pedido |
| `POST` | `/transactions/{transaction_id}/pix` | — | Gerar cobrança PIX (copia-e-cola e QR code) (`transaction_pix`) |
| `GET` | `/transactions/{transaction_id}/pix` | — | Checar status e expiração do PIX |
| `GET` | `/orders/{order_id}/receipt` | — | Obter recibo/comprovante digital do pedido (`receipt`) |
| **Operação de Turnos & Equipe** | | | |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/staff-sessions` | — | Listar turnos operacionais (`restaurant_staff_session`) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/staff-sessions/open` | `{ shift_code }` | Abrir novo turno de trabalho e gerar QR token |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/staff-sessions/{session_id}/close` | — | Encerrar turno de trabalho |
| `POST` | `/staff-sessions/check-in` | `{ qr_token, account_id }` | Bater ponto de entrada no turno ativo (`staff_session`) |
| `POST` | `/staff-sessions/{id}/check-out` | — | Registrar saída do turno (`staff_session`) |
| `GET` | `/staff-sessions/active` | — | Listar equipe ativa em turno no momento |

---

## Arquitetura & Modelo de Segurança

### 1. Arquitetura em 3 Camadas

O backend é estruturado em três camadas com responsabilidades estritas:

1. **Camada de Interface (`handlers.go`, `routes.go`)**:
   - Parse e decodificação do payload HTTP JSON e query/path parameters.
   - Validação estrutural de entrada (campos obrigatórios, formatos de email e senha).
   - Autenticação de tokens JWT (`Authorization: Bearer <token>`) e injeção do contexto do usuário.
   - Tradução dos retornos de negócio e erros para códigos de status HTTP semânticos (200, 201, 204, 400, 401, 403, 404, 409, 500).

2. **Camada de Negócio / Regras de Negócio (`service.go`)**:
   - Controle de Acesso Baseado em Papéis (RBAC): validação de autoridade `OWNER`, `MANAGER` ou `sysadmin`.
   - Invariante de orfandade de entidade: impede a remoção do último proprietário de um restaurante.
   - Verificação de CNPJs: validação cruzada entre o CNPJ do proprietário registrado e a criação da filial física.
   - Segurança de credenciais com `bcrypt` (custo mínimo 12) e validação de tokens JWT HMAC-SHA256.
   - Máquina de estados para sessões de terminais de salão e política de bloqueio por força bruta.

3. **Camada de Persistência (`store.go`, `schema.sql`)**:
   - Execução de queries preparadas e transações ACID via PostgreSQL (`pgx`).
   - Implementação de repositórios com interfaces desacopladas (`AccountStore`, `RestaurantStore`, `DishStore`).

---

### 2. Contas Globais e Membresias da Entidade

- **Tipos de Contas Globais**:
  - `account` (`users`): Contas globais criadas em `/auth/register` com hash seguro de senha em `bcrypt`. Podem atuar como consumidores ou receber permissões em restaurantes.
  - `sysadmin`: Superadministradores da plataforma Pratello com privilégios de governança em todo o sistema.

- **Papéis no Restaurante Lógico (`restaurant`)**:
  - `OWNER` (Proprietário): Detém autoridade no nível de entidade. Pode nomear ou revogar outros proprietários (`POST`/`DELETE /restaurants/{slug}/owners`), alterar dados fiscais/CNPJ, gerenciar faturamento e excluir a marca. Múltiplos proprietários são suportados. É proibida a remoção do último proprietário (prevenção de orfandade).
  - `MANAGER` (Gerente Operacional): Detém autoridade operacional plena. Pode gerenciar cardápios, categorias, itens, pratos, opcionais, promoções, gerenciar colaboradores de RH e criar filiais físicas. Não possui autoridade fiscal ou de destruição da entidade.
  - Uma mesma conta de usuário pode acumular simultaneamente ambos os papéis (`OWNER` e `MANAGER`).

---

### 3. Gestão de Colaboradores de RH (Desacoplamento de Acesso)

- Colaboradores são registrados primeiramente como fichas cadastrais de RH (`employee`), vinculadas a uma filial física (`name`, `cpf`, `position`).
- **Nenhum login ou conta de usuário é gerado automaticamente** na contratação.
- Acesso administrativo ao sistema é concedido estritamente via fluxo explícito de convite (`POST /restaurants/{slug}/branches/{branch_id}/employees/{employee_id}/invite`), associando um `account_id` previamente existente.

---

### 4. Terminais de Salão & PIN de Turno (Sessões Device-Bound)

- **PIN Compartilhado**: Cada filial física possui um PIN numérico compartilhado de 6 dígitos, utilizado em dispositivos de chão de fábrica/salão.
- **Validade do Turno**: A sessão aberta via PIN expira automaticamente em 8 horas e é restrita ao turno operacional vigente.
- **Vinculação de Dispositivo**: O token de sessão é estritamente atrelado ao identificador físico do terminal (`device_id`). O envio do cabeçalho `X-Device-Id` é obrigatório nas requisições.
- **Proteção contra Força Bruta**: Após 5 tentativas incorretas consecutivas de PIN, o terminal é bloqueado temporariamente por 15 minutos (`lockout`).
- **Raio de Alcance Mínimo (Blast Radius)**:
  - Sessões de terminal não são nominalmente atribuíveis a pessoas físicas.
  - Acesso estritamente restrito à visualização operacional de pedidos da respectiva filial (`GET .../terminal/orders`).
  - Sem permissões de gestão, relatórios financeiros, dados sensíveis (PII de clientes) ou acesso entre filiais distintas.
