# Plano e Especificação Detalhada de Testes — Pratello (DBML Alinhado)

> **Classificação:** Suíte de Testes Profunda (Deep Testing Specification)  
> **Total de Casos de Teste:** 200 Casos Rigorosos  
> **Conformidade de Schema:** Totalmente alinhado ao DBML atual do Pratello.  
> **Novas Diretrizes de Domínio:**  
> - **Privacidade do Pedido:** `order` não possui CPF nem chave estrangeira para `account` ou `users`.  
> - **Ciclo de Vida das Moedas (`user_restaurant`):** O CPF acumula saldo a cada compra (cria o registro se não existir, atualiza se existir). **O resgate exige obrigatoriamente uma conta vinculada (`account_id IS NOT NULL`)**.  
> - **Enums Padronizados:** Status e tipos em caixa alta (`PENDING`, `COMPLETED`, `LOCAL`, `REMOTE`, `PIX`, etc.).

---

## Resumo Executivo e Metodologia de Teste

A suíte está estruturada em **6 Módulos Coesos de Alta Profundidade (Deep Modules)** cobrindo Caminho Feliz (*Happy Path*), Condições de Limite (*Boundary Testing*), Casos Negativos de Erro, Integridade Referencial (*ACID/FK*), Idempotência e Concorrência:

1. **Módulo 1: Identidade, Contas e Gestão (IDP)** — Casos `IDP-001` a `IDP-035` (35 testes): Contas (`account`), documento flexível (`document` para CPF/CNPJ), papéis e tipos em estabelecimentos (`account_restaurant`), endereço de cobrança (`Address`) e proteção contra injeções.
2. **Módulo 2: Estabelecimentos, Filiais e Infraestrutura (TEN)** — Casos `TEN-001` a `TEN-030` (30 testes): Gestão de tenants/marcas (`restaurant` com `name_slug` e `cnpj`), filiais físicas (`physical_restaurant`), telefones de unidade (`restaurant_phone`) e isolamento multi-tenant.
3. **Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT)** — Casos `CAT-001` a `CAT-045` (45 testes): Layout visual (`menu` com `CLASSICO` / `EXPERIENCIA`), categorias, itens (`menu_item`), ficha técnica com medidas (`item_ingredient`), adicionais (`item_side`, `side_option`), combos vinculados (`combo_item`) e promoções (`item_sale`).
4. **Módulo 4: Pedidos Anônimos, Especialização e Itens (ORD)** — Casos `ORD-001` a `ORD-040` (40 testes): Pedidos desacoplados de CPF com `customer_name` opcional, especialização estrita (`local_order` com `"table"` vs `remote_order` com `"delivery_fee"`), congelamento de valor (`unit_price`), combos com `parent_order_id` e máquina de estados (`PENDING` -> `COMPLETED`).
5. **Módulo 5: Pagamentos, PIX, Cupons e Fidelidade (FIN)** — Casos `FIN-001` a `FIN-030` (30 testes): Tipos de transação (`PIX`, `CREDIT`, `DEBIT`, `CASH`), geração de PIX com `pix_copy_paste`, recibos em JSONB (`receipt`), cupons com restrição composta `(restaurant_id, code)` e **ciclo de moedas (acúmulo anônimo por CPF vs exigência de `account` para resgate)**.
6. **Módulo 6: Turnos Operacionais, Concorrência e Resiliência (OPS)** — Casos `OPS-001` a `OPS-020` (20 testes): Sessões de turno (`staff_session` com `shift_code` de 6 dígitos e `qr_token`), concorrência em pedidos e estoque, exaustão de pool de conexões e graceful shutdown.

---

## Módulo 1: Identidade, Contas e Gestão (IDP-001 a IDP-035)

### IDP-001: Cadastro de Conta com E-mail e Senha Válidos
- **Pré-condição:** E-mail `admin@pratello.com` não cadastrado na tabela `account`.
- **Passos:** POST `/auth/register`. Inspecionar `SELECT id, email, password, document, address_id FROM account WHERE email = ?`.
- **Entradas:**
  - Válida com CPF: `{"email": "admin@pratello.com", "password": "Password123!", "name": "Admin Pratello", "document": "52998224725", "phone": "11987654321"}`
  - Válida com CNPJ: `{"email": "gestor@corp.com", "password": "Password123!", "name": "Gestor Corp", "document": "12345678000199", "phone": "11987654322"}`
- **Esperado:** HTTP 201 Created. Senha protegida por hash. Campo `document` persistido sem formatação.

### IDP-002: Rejeição de E-mail Duplicado (Unique Constraint em `account`)
- **Pré-condição:** Conta com e-mail `owner@bela.com` já gravada.
- **Passos:** POST `/auth/register` com variações de caixa alta/baixa do mesmo e-mail.
- **Entradas:** `{"email": "OWNER@BELA.COM", "password": "OtherPassword99#"}`.
- **Esperado:** HTTP 409 Conflict. Erro `email already registered`.

### IDP-003: Validação de Formato de E-mail Inválido
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/auth/register` com e-mails malformados.
- **Entradas:** `"email": "usuario-sem-arroba.com"`, `"email": "usuario@"`, `"email": ""`
- **Esperado:** HTTP 400 Bad Request detalhando erro no campo `email`.

### IDP-004: Políticas de Segurança e Complexidade de Senha
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter senhas com menos de 8 caracteres ou sem caracteres especiais.
- **Entradas:** `"password": "123456"`, `"password": "abcdefgh"`.
- **Esperado:** HTTP 400 Bad Request com violação de requisitos de senha.

### IDP-005: Autenticação de Conta (`/auth/login`)
- **Pré-condição:** Conta `user@pratello.com` existente.
- **Passos:** POST `/auth/login` com credenciais corretas.
- **Esperado:** HTTP 200 OK retornando token JWT e objeto `account` (sem o campo de senha).

### IDP-006: Rejeição de Credenciais Inválidas com Mensagem Anti-Enumeração
- **Pré-condição:** Conta existente.
- **Passos:** POST `/auth/login` com senha incorreta ou e-mail inexistente.
- **Esperado:** HTTP 401 Unauthorized com mensagem neutra: `invalid email or password`.

### IDP-007: Associação de Conta a Estabelecimento (`account_restaurant`)
- **Pré-condição:** `account_id = 1` e `restaurant_id = 5` existentes.
- **Passos:** POST `/restaurants/5/staff` associando papel de gerente.
- **Entradas:** `{"account_id": 1, "account_type": "MANAGER", "document": "52998224725"}`.
- **Esperado:** HTTP 201 Created. Registro gravado na tabela `account_restaurant`.

### IDP-008: Unicidade Composta da Associação (`account_restaurant`)
- **Pré-condição:** Vínculo de `account_id = 1` com `restaurant_id = 5` já existente.
- **Passos:** Repetir a inserção para o mesmo restaurante.
- **Esperado:** HTTP 409 Conflict violando o índice de chave primária composta `(id, restaurant_id)`.

### IDP-009: Associação da Conta com Endereço de Cobrança (`Address`)
- **Pré-condição:** Endereço ID 10 cadastrado em `Address`.
- **Passos:** Atualizar conta associando `address_id = 10`.
- **Esperado:** HTTP 200 OK. FK referencial em `account.address_id` validada com sucesso.

### IDP-010: Tentativa de Vincular `address_id` Inexistente na Conta
- **Pré-condição:** Endereço ID 999999 não existe.
- **Passos:** POST `/auth/register` com `{"address_id": 999999}`.
- **Esperado:** HTTP 400 Bad Request (Foreign key constraint violation).

### IDP-011: Validação de Campo `document` (CPF de 11 dígitos ou CNPJ de 14 dígitos)
- **Pré-condição:** Nenhuma.
- **Passos:** Testar valores no campo `document` da conta.
- **Entradas:**
  - Válido CPF: `"document": "52998224725"` (11 dígitos) -> HTTP 201.
  - Válido CNPJ: `"document": "12345678000199"` (14 dígitos) -> HTTP 201.
  - Inválido (tamanho 9): `"document": "123456789"` -> HTTP 400.
  - Inválido (tamanho 12): `"document": "123456789012"` -> HTTP 400.
- **Esperado:** Validação estrita de documentos corporativos e pessoais.

### IDP-012: Desvinculação de Conta de um Restaurante
- **Pré-condição:** Registro em `account_restaurant` para `account_id = 2` e `restaurant_id = 5`.
- **Passos:** DELETE `/restaurants/5/staff/2`.
- **Esperado:** HTTP 204 No Content. Registro removido de `account_restaurant`. A tabela `account` permanece inalterada.

### IDP-013: Obtenção de Dados da Própria Conta (`/auth/me`)
- **Pré-condição:** Token de autenticação válido.
- **Passos:** GET `/auth/me` com header `Authorization: Bearer <token>`.
- **Esperado:** HTTP 200 OK com `id`, `name`, `email`, `document`, `phone` e lista de restaurantes associados via `account_restaurant`.

### IDP-014: Rejeição de Token Expirado
- **Pré-condição:** Token assinado com timestamp de expiração expirado.
- **Passos:** GET `/auth/me`.
- **Esperado:** HTTP 401 Unauthorized `token has expired`.

### IDP-015: Rejeição de Token com Assinatura Forjada
- **Pré-condição:** Token com payload alterado.
- **Passos:** Requisição com o token adulterado.
- **Esperado:** HTTP 401 Unauthorized `invalid signature`.

### IDP-016: Controle de Permissão por `account_type`
- **Pré-condição:** Conta vinculada com `account_type = 'WAITER'`.
- **Passos:** Tentar alterar configurações fiscais do restaurante.
- **Esperado:** HTTP 403 Forbidden. Operação permitida apenas para `account_type = 'MANAGER'` ou `'ADMIN'`.

### IDP-017: Prevenção de SQL Injection no Login
- **Pré-condição:** Tabela `account` com registros ativos.
- **Passos:** Submeter POST `/auth/login` com strings `' OR '1'='1`.
- **Esperado:** HTTP 401 Unauthorized. Nenhuma instrução arbitrária executada.

### IDP-018: Sanitização de Nome de Conta contra XSS
- **Pré-condição:** Nenhuma.
- **Passos:** Cadastrar conta com `{"name": "<script>alert(1)</script>Carlos"}`.
- **Esperado:** Caracteres escapados no banco e na resposta JSON.

### IDP-019: Rate Limiting em Tentativas Repetidas de Login
- **Pré-condição:** Conta existente.
- **Passos:** 10 tentativas com senha incorreta em 5 segundos.
- **Esperado:** HTTP 429 Too Many Requests a partir da 5ª tentativa.

### IDP-020: Atualização de Telefone e Nome da Conta
- **Pré-condição:** Conta cadastrada.
- **Passos:** PUT `/auth/me` com `{"name": "Novo Nome", "phone": "11988887777"}`.
- **Esperado:** HTTP 200 OK com os dados atualizados.

### IDP-021: Alteração de Senha com Verificação de Senha Anterior
- **Pré-condição:** Senha atual `OldPass@123`.
- **Passos:** Submeter nova senha informando a anterior.
- **Esperado:** HTTP 200 OK. Hash atualizado; senha antiga invalidada.

### IDP-022: Atualização de Endereço Residencial do Usuário
- **Pré-condição:** Conta vinculada a um `Address`.
- **Passos:** PUT no endereço da conta com novo CEP e número.
- **Esperado:** HTTP 200 OK. Tabela `Address` atualizada.

### IDP-023: Vínculo da Mesma Conta em Diferentes Estabelecimentos
- **Pré-condição:** `account_id = 3` vinculada a `restaurant_id = 1`.
- **Passos:** Vincular `account_id = 3` a `restaurant_id = 2` com outro papel.
- **Esperado:** HTTP 201 Created. A mesma conta pode atuar em múltiplos restaurantes.

### IDP-024: Consulta de Todas as Contas Vinculadas a um Restaurante
- **Pré-condição:** Restaurante ID 1 com 3 colaboradores.
- **Passos:** GET `/restaurants/1/staff`.
- **Esperado:** HTTP 200 OK com lista de 3 colaboradores e seus respectivos `account_type`.

### IDP-025: Tentativa de Vincular `account_id` Inexistente a um Restaurante
- **Pré-condição:** `account_id = 88888` não cadastrado.
- **Passos:** POST `/restaurants/1/staff` com ID inexistente.
- **Esperado:** HTTP 404 Not Found (ou violação de chave estrangeira).

### IDP-026: Tentativa de Vincular Colaborador a `restaurant_id` Inexistente
- **Pré-condição:** `restaurant_id = 9999` não cadastrado.
- **Passos:** POST `/restaurants/9999/staff`.
- **Esperado:** HTTP 404 Not Found.

### IDP-027: Logout com Invalidação de Sessão
- **Pré-condição:** Token válido.
- **Passos:** POST `/auth/logout`. Reutilizar o token na sequência.
- **Esperado:** Segunda chamada rejeitada com HTTP 401 Unauthorized.

### IDP-028: Preservação de Caracteres UTF-8 em `account.name`
- **Pré-condição:** Nenhuma.
- **Passos:** Salvar conta com acentuação e caracteres especiais (`"João Peña de Alcântara"`).
- **Esperado:** HTTP 201 Created com codificação UTF-8 preservada.

### IDP-029: Validação de Payload JSON Vazio ou Truncado
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar body `{"email":` sem fechar.
- **Esperado:** HTTP 400 Bad Request `invalid JSON syntax`.

### IDP-030: Rejeição de E-mail com Mais de 255 Caracteres
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/auth/register` com e-mail de 256 caracteres.
- **Esperado:** HTTP 400 Bad Request violando o limite da coluna `varchar(255)`.

### IDP-031: Consulta de Permissões de um Usuário por Restaurante
- **Pré-condição:** `account_id = 1` é `MANAGER` no restaurante 10.
- **Passos:** GET `/restaurants/10/staff/1/role`.
- **Esperado:** HTTP 200 OK retornando `{"account_type": "MANAGER"}`.

### IDP-032: Exclusão de Conta sem Dependências Fiscais
- **Pré-condição:** Conta recém-criada sem vínculos operacionais.
- **Passos:** DELETE `/auth/me`.
- **Esperado:** HTTP 204 No Content. Registro excluído.

### IDP-033: Bloqueio de Exclusão de Conta Vinculada a Restaurante Ativo
- **Pré-condição:** Conta vinculada como `MANAGER` em restaurante ativo.
- **Passos:** Tentar excluir conta.
- **Esperado:** HTTP 409 Conflict. Exige desvincular ou transferir a gerência previamente.

### IDP-034: Validação de Formato do Campo `phone` na Conta
- **Pré-condição:** Nenhuma.
- **Passos:** Cadastrar conta com telefone contendo letras.
- **Esperado:** HTTP 400 Bad Request `phone must contain only digits`.

### IDP-035: Recuperação de Senha via Token Temporário
- **Pré-condição:** Conta existente.
- **Passos:** POST `/auth/forgot-password`. Utilizar token gerado para redefinir senha via POST `/auth/reset-password`.
- **Esperado:** HTTP 200 OK. Senha redefinida.

---

## Módulo 2: Estabelecimentos, Filiais e Infraestrutura (TEN-001 a TEN-030)

### TEN-001: Criação de Restaurante (`restaurant`) com `name_slug` e `cnpj`
- **Pré-condição:** Slug `pizzaria-bela` disponível.
- **Passos:** POST `/restaurants`.
- **Entradas:** `{"name": "Pizzaria Bela", "name_slug": "pizzaria-bela", "brand": "Bela Alimentos", "cnpj": "12345678000199"}`.
- **Esperado:** HTTP 201 Created. Registro inserido na tabela `restaurant` com ID autoincremental e `created_at = now()`.

### TEN-002: Validação de Formato de `name_slug`
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/restaurants` com slugs inválidos.
- **Entradas:** `"name_slug": "pizza com espaço"`, `"name_slug": "PIZZA_MAIUSCULA"`.
- **Esperado:** HTTP 400 Bad Request. Aceito somente `^[a-z0-9-]+$`.

### TEN-003: Unicidade de `name_slug` entre Restaurantes
- **Pré-condição:** Restaurante com `name_slug = 'bela-pizza'` já cadastrado.
- **Passos:** Submeter outro restaurante com o mesmo `name_slug`.
- **Esperado:** HTTP 409 Conflict `restaurant slug already exists`.

### TEN-004: Validação de CNPJ do Estabelecimento (Algoritmo Módulo 11)
- **Pré-condição:** Nenhuma.
- **Passos:** Cadastrar restaurante com CNPJ de dígitos verificadores errados.
- **Entradas:** `"cnpj": "12345678000100"`.
- **Esperado:** HTTP 400 Bad Request `invalid CNPJ verification digits`.

### TEN-005: Consulta de Restaurante por ID e por `name_slug`
- **Pré-condição:** Restaurante ID 1 com slug `pizzaria-bela`.
- **Passos:** GET `/restaurants/1` e GET `/restaurants/by-slug/pizzaria-bela`.
- **Esperado:** Ambas as requisições devolvem HTTP 200 OK com o mesmo payload.

### TEN-006: Atualização de Dados Cadastrais do Restaurante
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** PUT `/restaurants/1` com novo `name` e `brand`.
- **Esperado:** HTTP 200 OK. Dados alterados.

### TEN-007: Criação de Unidade Física (`physical_restaurant`) com Endereço
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** POST `/restaurants/1/branches`.
- **Entradas:**
  ```json
  {
    "cnpj": "12345678000270",
    "address": {
      "CEP": "51020010",
      "number": 500,
      "apt": "Loja 1",
      "complementary": "Em frente à praia"
    }
  }
  ```
- **Esperado:** HTTP 201 Created. Linha inserida em `Address`, e chave primária composta `(id, restaurant_id, address_id)` criada em `physical_restaurant`.

### TEN-008: Validação de Número Inteiro Obrigatório em `Address.number`
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar endereço com número ausente ou como texto alfabético (`"number": "S/N"`).
- **Esperado:** HTTP 400 Bad Request `number must be an integer`. Se não houver número, deve ser convencionado valor `0` com complemento `"S/N"`.

### TEN-009: Validação do Campo `Address.CEP`
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter CEPs inválidos.
- **Entradas:** `"CEP": "123"` (3 dígitos), `"CEP": "510200101"` (9 dígitos).
- **Esperado:** HTTP 400 Bad Request.

### TEN-010: Cadastro de Telefone de Unidade Física (`restaurant_phone`)
- **Pré-condição:** Unidade física ID 10 existente.
- **Passos:** POST `/branches/10/phones`.
- **Entradas:** `{"phone": "81999990000", "label": "Delivery WhatsApp"}`.
- **Esperado:** HTTP 201 Created. Registro inserido em `restaurant_phone` com `physical_restaurant_id = 10`.

### TEN-011: Telefone com Label Padrão (`label = 'Principal'`)
- **Pré-condição:** Unidade física existente.
- **Passos:** POST `/branches/10/phones` omitindo `label`.
- **Esperado:** HTTP 201 Created. Coluna `label` persistida com valor default `'Principal'`.

### TEN-012: Listagem de Unidades Físicas por Restaurante
- **Pré-condição:** Restaurante ID 1 possui 2 filiais cadastradas.
- **Passos:** GET `/restaurants/1/branches`.
- **Esperado:** HTTP 200 OK com array contendo as 2 unidades e seus respectivos endereços.

### TEN-013: Remoção de Unidade Física sem Histórico de Pedidos
- **Pré-condição:** Unidade física recém-criada sem pedidos.
- **Passos:** DELETE `/branches/10`.
- **Esperado:** HTTP 204 No Content. Unidade e telefones associados removidos.

### TEN-014: Bloqueio de Exclusão de Unidade Física com Pedidos Associados
- **Pré-condição:** Unidade física possui registros em `order`.
- **Passos:** DELETE `/branches/10`.
- **Esperado:** HTTP 409 Conflict violando integridade referencial com a tabela `order`.

### TEN-015: Atualização de Dados da Unidade Física (CNPJ da Filial)
- **Pré-condição:** Unidade física existente.
- **Passos:** PUT `/branches/10` atualizando o CNPJ da filial.
- **Esperado:** HTTP 200 OK. CNPJ alterado após validação.

### TEN-016: Consulta de Unidade Física por ID Inexistente
- **Pré-condição:** ID 99999 inexistente.
- **Passos:** GET `/branches/99999`.
- **Esperado:** HTTP 404 Not Found `branch not found`.

### TEN-017: Múltiplos Telefones para a Mesma Unidade Física
- **Pré-condição:** Unidade física cadastrada.
- **Passos:** Inserir telefone principal e telefone de suporte.
- **Esperado:** Registros aceitos e listados corretamente em GET `/branches/10/phones`.

### TEN-018: Atualização de Endereço de Unidade Física
- **Pré-condição:** Unidade física vinculada a `address_id = 5`.
- **Passos:** PUT `/branches/10/address` com novos dados de logradouro.
- **Esperado:** HTTP 200 OK. Registro em `Address` atualizado.

### TEN-019: Isolamento Multi-Tenant em Unidades Físicas
- **Pré-condição:** Restaurante 1 e Restaurante 2 possuem filiais próprias.
- **Passos:** Tentar associar uma filial do Restaurante 2 à rota do Restaurante 1.
- **Esperado:** HTTP 403 Forbidden ou HTTP 404 Not Found.

### TEN-020: Criação de Restaurante com CNPJ em Branco (Opcional no DBML)
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/restaurants` omitindo campo `cnpj`.
- **Esperado:** HTTP 201 Created com coluna `cnpj` nula (`NULL`).

### TEN-021: Busca de Restaurantes por Filtro de Nome
- **Pré-condição:** Restaurantes cadastrados.
- **Passos:** GET `/restaurants?name=Napoli`.
- **Esperado:** Retorno filtrado apenas com restaurantes cujo nome contém "Napoli".

### TEN-022: Paginação da Listagem de Restaurantes
- **Pré-condição:** 30 restaurantes cadastrados.
- **Passos:** GET `/restaurants?limit=10&page=2`.
- **Esperado:** Retorno exato dos registros 11 a 20.

### TEN-023: Exclusão de Telefone de Unidade Física
- **Pré-condição:** Telefone cadastrado para a unidade 10.
- **Passos:** DELETE `/branches/10/phones`.
- **Esperado:** HTTP 204 No Content.

### TEN-024: Verificação de CEP Inexistente via Integração de Correios
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter CEP `00000000`.
- **Esperado:** HTTP 400 Bad Request `unrecognized or invalid CEP`.

### TEN-025: Sanitização de Nome de Restaurante
- **Pré-condição:** Nenhuma.
- **Passos:** Criar restaurante com tags HTML no nome.
- **Esperado:** Tags removidas ou escapadas no retorno.

### TEN-026: Verificação de Chave Composta de `physical_restaurant`
- **Pré-condição:** Unidade física com `(id=1, restaurant_id=1, address_id=1)`.
- **Passos:** Tentar inserir linha com chave primária idêntica.
- **Esperado:** HTTP 409 Conflict.

### TEN-027: Cadastro de Filial com CNPJ da Matriz
- **Pré-condição:** Restaurante matriz com CNPJ `12345678000199`.
- **Passos:** Cadastrar filial com CNPJ `12345678000270` (mesma raiz).
- **Esperado:** HTTP 201 Created.

### TEN-028: Remoção de Endereço sem Referência Ativa
- **Pré-condição:** Endereço criado que teve filial excluída.
- **Passos:** Executar limpeza de endereços órfãos.
- **Esperado:** Registro em `Address` removido sem impacto em outras entidades.

### TEN-029: Inclusão de `brand` com Tamanho Limite de 500 Caracteres
- **Pré-condição:** Nenhuma.
- **Passos:** POST restaurante com `brand` de 500 caracteres.
- **Esperado:** HTTP 201 Created.

### TEN-030: Rejeição de `brand` Excedendo 500 Caracteres
- **Pré-condição:** Nenhuma.
- **Passos:** POST com `brand` de 501 caracteres.
- **Esperado:** HTTP 400 Bad Request.

---

## Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT-001 a CAT-045)

### CAT-001: Criação de Menu com Estilo Válido (`menu_style_enum`)
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** POST `/restaurants/1/menu`.
- **Entradas:**
  - Estilo Clássico: `{"name": "Menu Principal", "style": "CLASSICO", "accent_color": "#FF5722"}`
  - Estilo Experiência: `{"name": "Menu Degustação", "style": "EXPERIENCIA", "accent_color": "#4CAF50"}`
- **Esperado:** HTTP 201 Created para ambos. Chave composta `(id, restaurant_id)` criada na tabela `menu`.

### CAT-002: Rejeição de Menu com Estilo Fora do Enum
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/restaurants/1/menu` com `{"style": "MODERNO"}`.
- **Esperado:** HTTP 400 Bad Request informando que apenas `CLASSICO` ou `EXPERIENCIA` são aceitos.

### CAT-003: Atualização de Cor de Destaque (`accent_color`) do Menu
- **Pré-condição:** Menu cadastrado.
- **Passos:** PUT `/restaurants/1/menu/1` com `{"accent_color": "#1A2B3C"}`.
- **Esperado:** HTTP 200 OK. Cor atualizada.

### CAT-004: Criação de Categoria (`category`) Vinculada a Restaurante
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** POST `/restaurants/1/categories` com `{"name": "Pizzas Tradicionais"}`.
- **Esperado:** HTTP 201 Created. Registro em `category` associado ao `restaurant_id = 1`.

### CAT-005: Listagem de Categorias do Restaurante
- **Pré-condição:** 3 categorias no Restaurante 1 e 2 categorias no Restaurante 2.
- **Passos:** GET `/restaurants/1/categories`.
- **Esperado:** HTTP 200 OK retornando estritamente as 3 categorias do Restaurante 1.

### CAT-006: Exclusão de Categoria Vazia
- **Pré-condição:** Categoria sem produtos vinculados.
- **Passos:** DELETE `/restaurants/1/categories/10`.
- **Esperado:** HTTP 204 No Content.

### CAT-007: Bloqueio de Exclusão de Categoria com Itens
- **Pré-condição:** Categoria possui itens cadastrados em `menu_item`.
- **Passos:** DELETE `/restaurants/1/categories/1`.
- **Esperado:** HTTP 409 Conflict violando FK `menu_item.category_id`.

### CAT-008: Criação de Item do Cardápio (`menu_item`)
- **Pré-condição:** Restaurante 1 e Categoria 1 existentes.
- **Passos:** POST `/restaurants/1/items`.
- **Entradas:**
  ```json
  {
    "category_id": 1,
    "name": "Pizza Calabresa",
    "description": "Molho de tomate, calabresa fatiada e cebola",
    "price": 42.50,
    "image_url": "https://cdn.pratello.com/calabresa.png",
    "is_available": true,
    "is_combo": false
  }
  ```
- **Esperado:** HTTP 201 Created. Registro inserido em `menu_item`.

### CAT-009: Validação de Preço de Item (`price` Numérico Positivo)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter item com `{"price": -5.00}`.
- **Esperado:** HTTP 400 Bad Request `price must be greater than zero`.

### CAT-010: Alternância de Disponibilidade de Item (`is_available`)
- **Pré-condição:** Item disponível.
- **Passos:** PATCH `/restaurants/1/items/5` com `{"is_available": false}`.
- **Esperado:** HTTP 200 OK. Item não deve mais constar na listagem pública do cardápio.

### CAT-011: Cadastro de Ingrediente Base no Catálogo Global (`Ingredient_item`)
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/ingredients` com `{"name": "Orégano Desidratado"}`.
- **Esperado:** HTTP 201 Created. Registro em `Ingredient_item`.

### CAT-012: Composição de Ficha Técnica (`item_ingredient`) com Medida e Quantidade
- **Pré-condição:** Item ID 5 e Ingrediente ID 2 existentes.
- **Passos:** POST `/items/5/ingredients`.
- **Entradas:** `{"ingredient_id": 2, "is_removable": true, "quantity": 100, "measure": "gramas"}`.
- **Esperado:** HTTP 201 Created. Chave composta `(item_id, ingredient_id)` gravada com sucesso.

### CAT-013: Ingrediente Não Removível na Ficha Técnica (`is_removable = false`)
- **Pré-condição:** Ficha técnica configurada com `is_removable = false`.
- **Passos:** Consultar ficha técnica do produto.
- **Esperado:** Flag retornada como `false` para que a UI do cliente desabilite a opção de remover.

### CAT-014: Criação de Grupo de Opcionais (`item_side`)
- **Pré-condição:** Item ID 5 existente.
- **Passos:** POST `/items/5/sides`.
- **Entradas:** `{"name": "Borda Recheada", "amount": 1, "price": 8.00}`.
- **Esperado:** HTTP 201 Created na tabela `item_side`.

### CAT-015: Associação de Opcional Escolhido ao Pedido (`side_option`)
- **Pré-condição:** `order_item_id = 20` e `item_side_id = 3` existentes.
- **Passos:** Inserir seleção do cliente em `side_option`.
- **Entradas:** `{"order_item_id": 20, "side_id": 3, "additional_price": 8.00}`.
- **Esperado:** Chave composta `(id, order_item_id, side_id)` gravada.

### CAT-016: Criação de Item do Tipo Combo (`is_combo = true`)
- **Pré-condição:** Categoria "Combos" existente.
- **Passos:** POST `/restaurants/1/items` com `{"name": "Combo Casal", "price": 79.90, "is_combo": true}`.
- **Esperado:** HTTP 201 Created com flag `is_combo = true`.

### CAT-017: Composição de Receita do Pacote Combo (`combo_item`)
- **Pré-condição:** Combo ID 10 ("Combo Casal") e Itens unitários ID 1 ("Pizza") e ID 2 ("Refrigerante").
- **Passos:** POST `/items/10/combo-items`.
- **Entradas:**
  - `{"included_item_id": 1, "quantity": 1}`
  - `{"included_item_id": 2, "quantity": 2}`
- **Esperado:** HTTP 201 Created. Chave composta `(combo_id, included_item_id)` inserida em `combo_item`.

### CAT-018: Prevenção de Inclusão do Próprio Combo nele Mesmo
- **Pré-condição:** Combo ID 10.
- **Passos:** POST com `{"included_item_id": 10}`.
- **Esperado:** HTTP 400 Bad Request `a combo cannot contain itself`.

### CAT-019: Cadastro de Promoção com Desconto Fixo (`item_sale` com `FIXED_AMOUNT`)
- **Pré-condição:** Item com preço de 50.00.
- **Passos:** POST `/items/1/sale`.
- **Entradas:**
  ```json
  {
    "discount_type": "FIXED_AMOUNT",
    "value": 10.00,
    "is_active": true,
    "starts_at": "2026-10-17T10:00:00Z",
    "ends_at": "2026-10-17T23:59:59Z"
  }
  ```
- **Esperado:** HTTP 201 Created. Registro em `item_sale` 1-para-1 com `item_id`.

### CAT-020: Cadastro de Promoção com Desconto Percentual (`item_sale` com `PERCENTAGE`)
- **Pré-condição:** Item com preço de 50.00.
- **Passos:** POST `/items/1/sale` com `{"discount_type": "PERCENTAGE", "value": 20.00}`.
- **Esperado:** HTTP 201 Created (representando 20% de desconto).

### CAT-021: Desativação Manual de Promoção (`is_active = false`)
- **Pré-condição:** Promoção ativa cadastrada.
- **Passos:** PATCH na promoção com `{"is_active": false}`.
- **Esperado:** HTTP 200 OK. Promoção não surte mais efeito no cálculo do preço.

### CAT-022: Validação de Janela Temporal de Promoção (`starts_at` e `ends_at`)
- **Pré-condição:** Promoção expirada no passado.
- **Passos:** Consultar preço efetivo do produto.
- **Esperado:** Preço integral do produto retornado, pois a data atual está fora do intervalo.

### CAT-023: Consulta Completa de Detalhes de Item com Ficha e Opcionais
- **Pré-condição:** Item cadastrado com ingredientes e adicionais.
- **Passos:** GET `/restaurants/1/items/1`.
- **Esperado:** HTTP 200 OK com árvore de dados aninhada.

### CAT-024: Atualização de Descrição e Foto do Item
- **Pré-condição:** Item existente.
- **Passos:** PUT `/restaurants/1/items/1` com nova URL de imagem.
- **Esperado:** HTTP 200 OK.

### CAT-025: Exclusão de Item sem Vendas Anteriores
- **Pré-condição:** Item sem pedidos em `order_item`.
- **Passos:** DELETE `/restaurants/1/items/99`.
- **Esperado:** HTTP 204 No Content.

### CAT-026: Bloqueio de Exclusão Física de Item com Pedidos Realizados
- **Pré-condição:** Item presente em compras na tabela `order_item`.
- **Passos:** DELETE `/restaurants/1/items/1`.
- **Esperado:** HTTP 409 Conflict violando FK com `order_item.item_id`.

### CAT-027: Desvinculação de Ingrediente da Ficha Técnica
- **Pré-condição:** Ingrediente vinculado ao item.
- **Passos:** DELETE `/items/1/ingredients/2`.
- **Esperado:** HTTP 204 No Content. Linha em `item_ingredient` excluída.

### CAT-028: Atualização de Quantidade de Sub-Itens no Combo
- **Pré-condição:** Combo com 1 refrigerante na receita.
- **Passos:** PUT `/items/10/combo-items/2` com `{"quantity": 3}`.
- **Esperado:** HTTP 200 OK. Quantidade alterada para 3.

### CAT-029: Remoção de Sub-Item da Receita do Combo
- **Pré-condição:** Combo cadastrado.
- **Passos:** DELETE `/items/10/combo-items/2`.
- **Esperado:** HTTP 204 No Content.

### CAT-030: Consulta Pública do Cardápio Filtrado por Disponibilidade
- **Pré-condição:** 10 itens no restaurante, 2 com `is_available = false`.
- **Passos:** GET `/restaurants/1/public-menu`.
- **Esperado:** Retorno com exatamente os 8 itens disponíveis.

### CAT-031: Validação de Desconto Superior a 100% no Tipo `PERCENTAGE`
- **Pré-condição:** Nenhuma.
- **Passos:** POST `item_sale` com `value = 110.00`.
- **Esperado:** HTTP 400 Bad Request `percentage discount cannot exceed 100%`.

### CAT-032: Validação de Desconto Fixo Superior ao Preço do Item
- **Pré-condição:** Item com preço de 30.00.
- **Passos:** POST `item_sale` com `FIXED_AMOUNT` e `value = 40.00`.
- **Esperado:** HTTP 400 Bad Request `discount cannot exceed item price`.

### CAT-033: Tentativa de Criar Item Vinculado a Categoria de Outro Restaurante
- **Pré-condição:** Categoria pertence ao Restaurante 2.
- **Passos:** Criar item no Restaurante 1 apontando para essa categoria.
- **Esperado:** HTTP 400 Bad Request informando inconsistência de tenant.

### CAT-034: Validação do Campo `measure` na Ficha Técnica (Ex: "g", "ml", "unidades")
- **Pré-condição:** Nenhuma.
- **Passos:** Salvar ingrediente com medida `"gramas"`.
- **Esperado:** HTTP 201 Created.

### CAT-035: Remoção de Grupo de Opcionais em Cascata
- **Pré-condição:** Grupo de opcionais cadastrado.
- **Passos:** DELETE `/items/1/sides/5`.
- **Esperado:** HTTP 204 No Content.

### CAT-036: Validação de Preço de Opcional em `item_side`
- **Pré-condição:** Nenhuma.
- **Passos:** POST com `{"price": -3.00}`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-037: Concorrência na Atualização de Estoque/Disponibilidade
- **Pré-condição:** Item disponível.
- **Passos:** Dois administradores alteram a disponibilidade no mesmo instante.
- **Esperado:** Ambos os commits são serializados sem conflito de integridade.

### CAT-038: Consulta de Cardápio com Estilo `EXPERIENCIA`
- **Pré-condição:** Menu configurado com `style = 'EXPERIENCIA'`.
- **Passos:** GET no menu público.
- **Esperado:** Metadados retornam o estilo configurado para direcionar a renderização da UI.

### CAT-039: Limite de Caracteres no Nome do Item (`varchar(255)`)
- **Pré-condição:** Nenhuma.
- **Passos:** POST item com nome de 256 caracteres.
- **Esperado:** HTTP 400 Bad Request.

### CAT-040: Descrição Longa de Item no Campo `text`
- **Pré-condição:** Nenhuma.
- **Passos:** Cadastrar item com descrição detalhada de 1.500 caracteres.
- **Esperado:** HTTP 201 Created. Campo `text` persiste integralmente.

### CAT-041: Consulta de Todos os Ingredientes de um Restaurante
- **Pré-condição:** Ingredientes cadastrados no catálogo.
- **Passos:** GET `/ingredients`.
- **Esperado:** HTTP 200 OK com array de ingredientes.

### CAT-042: Exclusão de Ingrediente sem Uso
- **Pré-condição:** Ingrediente nunca associado a nenhum item.
- **Passos:** DELETE `/ingredients/50`.
- **Esperado:** HTTP 204 No Content.

### CAT-043: Bloqueio de Exclusão de Ingrediente Usado em Ficha Técnica
- **Pré-condição:** Ingrediente presente em `item_ingredient`.
- **Passos:** DELETE `/ingredients/1`.
- **Esperado:** HTTP 409 Conflict.

### CAT-044: Validação de URL de Imagem Válida (`image_url`)
- **Pré-condição:** Nenhuma.
- **Passos:** Salvar URL inválida `"not-a-url"`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-045: Alteração de Preço de Item no Catálogo Não Altera Pedidos Anteriores
- **Pré-condição:** Item comprado anteriormente por 40.00.
- **Passos:** Atualizar preço do item para 55.00.
- **Esperado:** Pedido anterior permanece com `unit_price = 40.00` em `order_item`.

---

## Módulo 4: Pedidos Anônimos, Especialização e Itens (ORD-001 a ORD-040)

### ORD-001: Criação de Pedido de Salão (`LOCAL`) sem CPF
- **Pré-condição:** Unidade física ID 1 ativa; Item 1 (preço 30.00).
- **Passos:** POST `/branches/1/orders/local`.
- **Entradas:**
  ```json
  {
    "order_type": "LOCAL",
    "customer_name": "Mesa do João",
    "table": 4,
    "display_order": 1,
    "items": [
      {"item_id": 1, "quantity": 2}
    ]
  }
  ```
- **Esperado:** HTTP 201 Created. `order` gerado com `subtotal = 60.00`, `total = 60.00`, `status = 'PENDING'`. Registro em `local_order` (`table = 4`). **Nenhum campo de CPF solicitado ou armazenado**.

### ORD-002: Criação de Pedido para Entrega (`REMOTE`) sem Vínculo de Conta
- **Pré-condição:** Unidade física ID 1; Endereço cadastrado; Frete 10.00; Item 1 (50.00).
- **Passos:** POST `/branches/1/orders/remote`.
- **Entradas:**
  ```json
  {
    "order_type": "REMOTE",
    "customer_name": "Maria Delivery",
    "address_id": 20,
    "contact_phone": "81999990000",
    "delivery_fee": 10.00,
    "items": [
      {"item_id": 1, "quantity": 1}
    ]
  }
  ```
- **Esperado:** HTTP 201 Created. Linha em `remote_order` (`contact_phone = '81999990000'`, `delivery_fee = 10.00`). Total do pedido = 60.00 (`50.00 + 10.00`).

### ORD-003: Subtipo Exclusivo: Rejeição de Pedido com Tipos Conflitantes
- **Pré-condição:** Pedido criado com `order_type = 'LOCAL'`.
- **Passos:** Tentar inserir registro em `remote_order` para o mesmo `order.id`.
- **Esperado:** Erro de integridade HTTP 422 Unprocessable Entity. Um pedido possui apenas um subtipo exclusivo.

### ORD-004: Pedido com `customer_name` Opcional / Nulo
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido omitindo `customer_name`.
- **Esperado:** HTTP 201 Created com `customer_name = NULL`. Pedido processado normalmente.

### ORD-005: Congelamento Obrigatório de Preço em `order_item.unit_price`
- **Pré-condição:** Item 1 com `price = 45.00` no catálogo.
- **Passos:** Criar pedido contendo o Item 1. Consultar banco.
- **Esperado:** `order_item.unit_price` gravado estritamente como `45.00`.

### ORD-006: Captura do Preço Promocional de `item_sale` no `unit_price`
- **Pré-condição:** Item 1 (preço 50.00) com promoção ativa de R$ 10,00 de desconto (preço 40.00).
- **Passos:** Criar pedido durante o período da promoção.
- **Esperado:** `order_item.unit_price` gravado como `40.00`.

### ORD-007: Compra Fora da Promoção Registra Preço Integral
- **Pré-condição:** Promoção expirada ou inativa (`is_active = false`).
- **Passos:** Criar pedido.
- **Esperado:** `order_item.unit_price` gravado como `50.00`.

### ORD-008: Validação de Quantidade Positiva (`quantity >= 1`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido com `{"quantity": 0}` ou `{"quantity": -1}`.
- **Esperado:** HTTP 400 Bad Request `quantity must be at least 1`.

### ORD-009: Pedido com Múltiplas Linhas de Itens
- **Pré-condição:** Item A (20.00) e Item B (15.00).
- **Passos:** Criar pedido com 2x Item A e 1x Item B.
- **Esperado:** `subtotal = 55.00`, `total = 55.00`.

### ORD-010: Bloqueio de Pedido Contendo Produto Indisponível
- **Pré-condição:** Item com `is_available = false`.
- **Passos:** Submeter pedido com o item.
- **Esperado:** HTTP 400 Bad Request `item is currently unavailable`.

### ORD-011: Registro de Avaliação no Item do Pedido (`order_item.review`)
- **Pré-condição:** Pedido em status `COMPLETED`.
- **Passos:** POST `/order-items/15/review` com nota 5 (escala de 1 a 5).
- **Esperado:** HTTP 200 OK. Coluna `review` preenchida com `5`.

### ORD-012: Validação de Nota de Avaliação Fora da Escala (Ex: 0 ou 6)
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar avaliação com nota 7.
- **Esperado:** HTTP 400 Bad Request `review must be between 1 and 5`.

### ORD-013: Decomposição de Combo com `parent_order_id`
- **Pré-condição:** Item 10 é um combo composto por Itens 1 e 2.
- **Passos:** Comprar o Combo 10.
- **Esperado:**
  - Linha pai: Item 10 com preço do combo.
  - Linhas filhas: Itens 1 e 2 com `parent_order_id = <id_da_linha_pai>` e `unit_price = 0.00`.

### ORD-014: Aplicação de Cupom de Desconto Válido (`coupon_id` em `order`)
- **Pré-condição:** Cupom `DESC10` de 10% ativo; Pedido de subtotal 100.00.
- **Passos:** Submeter pedido com o cupom.
- **Esperado:** `order.coupon_id` preenchido, `discount_amount = 10.00`, `total = 90.00`.

### ORD-015: Rejeição de Cupom de Outro Estabelecimento
- **Pré-condição:** Cupom pertence ao Restaurante 2; Pedido feito na filial do Restaurante 1.
- **Passos:** Aplicar cupom.
- **Esperado:** HTTP 400 Bad Request `coupon is not valid for this restaurant`.

### ORD-016: Rejeição de Cupom Abaixo do Valor Mínimo (`min_order_value`)
- **Pré-condição:** Cupom exige pedido mínimo de 80.00; Pedido com subtotal 50.00.
- **Passos:** Aplicar cupom.
- **Esperado:** HTTP 400 Bad Request `order subtotal does not reach coupon minimum value`.

### ORD-017: Transição de Status: `PENDING` -> `PREPARING`
- **Pré-condição:** Pedido criado com status padrão `PENDING`.
- **Passos:** PATCH `/orders/100/status` com `{"status": "PREPARING"}`.
- **Esperado:** HTTP 200 OK. `status` atualizado para `PREPARING`.

### ORD-018: Transição de Status: `PREPARING` -> `READY`
- **Pré-condição:** Pedido em `PREPARING`.
- **Passos:** PATCH com `{"status": "READY"}`.
- **Esperado:** HTTP 200 OK. Notificação de pedido pronto emitida.

### ORD-019: Transição de Status: `READY` -> `COMPLETED`
- **Pré-condição:** Pedido em `READY`.
- **Passos:** PATCH com `{"status": "COMPLETED"}`.
- **Esperado:** HTTP 200 OK. Pedido concluído com sucesso.

### ORD-020: Cancelamento de Pedido (`CANCELLED`)
- **Pré-condição:** Pedido em `PENDING`.
- **Passos:** PATCH com `{"status": "CANCELLED"}`.
- **Esperado:** HTTP 200 OK. Pedido cancelado.

### ORD-021: Rejeição de Cancelamento de Pedido Já Concluído
- **Pré-condição:** Pedido com status `COMPLETED`.
- **Passos:** Tentar alterar status para `CANCELLED`.
- **Esperado:** HTTP 400 Bad Request `cannot cancel an already completed order`.

### ORD-022: Validação de Mesa em Pedido de Salão (`local_order.table > 0`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido com `"table": 0` ou `"table": -1`.
- **Esperado:** HTTP 400 Bad Request `table must be a positive integer`.

### ORD-023: Campo de Notas da Mesa (`local_order.notes`)
- **Pré-condição:** Nenhuma.
- **Passos:** Salvar pedido local com `"notes": "Próximo à janela, levar pratos extras"`.
- **Esperado:** HTTP 201 Created. Notas registradas.

### ORD-024: Validação de Telefone de Contato no Pedido Remoto
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido de delivery sem `contact_phone`.
- **Esperado:** HTTP 400 Bad Request `contact_phone is required for remote orders`.

### ORD-025: Campo de Instruções de Entrega (`remote_order.delivery_notes`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter delivery com `"delivery_notes": "Tocar interfone 202, deixar na portaria"`.
- **Esperado:** HTTP 201 Created. Observações de entrega persistidas.

### ORD-026: Tempo Estimado de Entrega (`remote_order.estimated_minutes`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter com `"estimated_minutes": 45`.
- **Esperado:** HTTP 201 Created. Estimativa exibida no rastreamento do pedido.

### ORD-027: Validação de Taxa de Frete Positiva (`delivery_fee >= 0`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter frete com valor `-5.00`.
- **Esperado:** HTTP 400 Bad Request `delivery_fee cannot be negative`.

### ORD-028: Adição de Itens a Pedido Aberto na Mesa
- **Pré-condição:** Pedido local em status `PENDING`.
- **Passos:** POST `/orders/100/items` com mais 1 item de 20.00.
- **Esperado:** HTTP 201 Created. `subtotal` e `total` recalculados com precisão.

### ORD-029: Bloqueio de Adição de Itens a Pedido Finalizado
- **Pré-condição:** Pedido em status `COMPLETED`.
- **Passos:** Tentar adicionar item.
- **Esperado:** HTTP 409 Conflict `cannot modify a completed order`.

### ORD-030: Remoção de Linha de Item do Pedido
- **Pré-condição:** Pedido aberto com 2 itens.
- **Passos:** DELETE `/orders/100/items/5`.
- **Esperado:** HTTP 204 No Content. Total recalculado.

### ORD-031: Consulta de Pedido por ID com Linhas de Itens
- **Pré-condição:** Pedido ID 100 cadastrado.
- **Passos:** GET `/orders/100`.
- **Esperado:** HTTP 200 OK com dados do pedido, subtipo e array de `order_item`.

### ORD-032: Consulta de Pedido por ID Inexistente
- **Pré-condição:** ID 888888 inexistente.
- **Passos:** GET `/orders/888888`.
- **Esperado:** HTTP 404 Not Found `order not found`.

### ORD-033: Listagem de Pedidos Filtrados por Status na Filial
- **Pré-condição:** 5 pedidos na filial (2 `PENDING`, 3 `PREPARING`).
- **Passos:** GET `/branches/1/orders?status=PREPARING`.
- **Esperado:** Retorno exato dos 3 pedidos em preparo.

### ORD-034: Rejeição de Pedido Vazio (Array de Itens Vazio)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido com `"items": []`.
- **Esperado:** HTTP 400 Bad Request `order must have at least one item`.

### ORD-035: Concorrência: Inserções Paralelas na Mesma Comanda
- **Pré-condição:** Comanda aberta na mesa 1.
- **Passos:** Dois garçons adicionam produtos simultaneamente.
- **Esperado:** Transação atômica serializada; nenhum item é perdido e o total bate com a soma real.

### ORD-036: Validação de FK: Pedido Apontando para Filial Inexistente
- **Pré-condição:** `physical_restaurant_id = 99999` não existe.
- **Passos:** POST pedido com essa filial.
- **Esperado:** HTTP 400 Bad Request violando chave estrangeira.

### ORD-037: Rejeição de Transição para Enum de Status Inexistente
- **Pré-condição:** Pedido ativo.
- **Passos:** PATCH com `{"status": "AGUARDANDO_PAGAMENTO"}`.
- **Esperado:** HTTP 400 Bad Request informando valores aceitos pelo `order_status_enum`.

### ORD-038: Pedido sem Associação de CPF ou Usuário em `order`
- **Pré-condição:** Nenhuma.
- **Passos:** Criar pedido informando apenas itens e mesa. Consultar colunas da tabela `order`.
- **Esperado:** Nenhuma coluna ou chave estrangeira para CPF/usuário existe na tabela `order`; privacidade do consumidor preservada por design.

### ORD-039: Idempotência de Criação de Pedido com Header `Idempotency-Key`
- **Pré-condição:** Cliente com instabilidade de rede envia duas requisições idênticas.
- **Passos:** Enviar dois POSTs com o mesmo cabeçalho `Idempotency-Key: req-xyz-789`.
- **Esperado:** Apenas um pedido é persistido; a segunda chamada retorna a resposta idempotente sem duplicar linhas em `order`.

### ORD-040: Deleção de Pedido com Transações Financeiras Bloqueada
- **Pré-condição:** Pedido com pagamentos registrados na tabela `transaction`.
- **Passos:** Executar DELETE no pedido.
- **Esperado:** Bloqueado por integridade referencial com a tabela `transaction`.

---

## Módulo 5: Pagamentos, PIX, Cupons e Fidelidade (FIN-001 a FIN-030)

### FIN-001: Pagamento Integral em Dinheiro (`transaction_type = 'CASH'`)
- **Pré-condição:** Pedido 100 de 50.00 em status `PENDING`.
- **Passos:** POST `/orders/100/transactions`.
- **Entradas:** `{"transaction_type": "CASH", "amount": 50.00, "notes": "Pago no caixa"}`.
- **Esperado:** HTTP 201 Created. Linha em `transaction` com `status = 'COMPLETED'`.

### FIN-002: Pagamento com Cartão de Crédito (`CREDIT`) e Débito (`DEBIT`)
- **Pré-condição:** Pedido aberto.
- **Passos:** Criar transação com tipo `CREDIT` ou `DEBIT`.
- **Esperado:** HTTP 201 Created. Registro inserido em conformidade com o `transaction_type_enum`.

### FIN-003: Pagamento Dividido da Comanda (Split Payment)
- **Pré-condição:** Pedido de 100.00.
- **Passos:** Transação 1 de 60.00 via `CREDIT` e Transação 2 de 40.00 via `PIX`.
- **Esperado:** Ambas transações aprovadas, totalizando os 100.00 devidos.

### FIN-004: Geração de Cobrança Instantânea PIX (`transaction_pix`)
- **Pré-condição:** Transação ID 50 criada com `transaction_type = 'PIX'`.
- **Passos:** POST `/transactions/50/pix`.
- **Esperado:** HTTP 201 Created na tabela `transaction_pix` contendo:
  - `pix_copy_paste` (string copia-e-cola).
  - `qr_code_url` (URL ou base64 do QR code).
  - `expires_at` (timestamp futuro).

### FIN-005: Webhook de Confirmação de Pagamento PIX
- **Pré-condição:** Transação PIX em status `PENDING`.
- **Passos:** Webhook bancário notifica pagamento aprovado.
- **Esperado:** `transaction.status` atualizado para `'COMPLETED'`.

### FIN-006: Expiração de PIX Não Pago
- **Pré-condição:** Transação PIX que ultrapassou o `expires_at`.
- **Passos:** Rotina de conciliação executa verificação.
- **Esperado:** `transaction.status` marcado como `'EXPIRED'`.

### FIN-007: Emissão de Recibo Estruturado em JSON (`receipt`)
- **Pré-condição:** Transação ID 50 aprovada (`COMPLETED`).
- **Passos:** Gerar recibo digital.
- **Esperado:** Linha inserida na tabela `receipt` com `transaction_id = 50`, `receipt_code` único e `receipt` preenchido como objeto `jsonb`.

### FIN-008: Unicidade do Código de Recibo (`receipt.receipt_code`)
- **Pré-condição:** Código `REC-123456` já emitido.
- **Passos:** Tentar gerar outro recibo com mesmo código.
- **Esperado:** HTTP 409 Conflict.

### FIN-009: Criação de Cupom com Restrição de Unicidade Composta `(restaurant_id, code)`
- **Pré-condição:** Restaurante ID 1.
- **Passos:** POST `/restaurants/1/coupons` com `code = 'PRATELLO10'`.
- **Esperado:** HTTP 201 Created. Registro inserido na tabela `coupon`.

### FIN-010: Mesmo Código de Cupom Permitido em Restaurantes Distintos
- **Pré-condição:** Cupom `PRATELLO10` existe no Restaurante ID 1.
- **Passos:** Cadastrar cupom `PRATELLO10` no Restaurante ID 2.
- **Esperado:** HTTP 201 Created devido ao índice composto `idx_coupons_restaurant_id_code`.

### FIN-011: Bloqueio de Cupom Duplicado no Mesmo Restaurante
- **Pré-condição:** Cupom `PRATELLO10` já existente no Restaurante 1.
- **Passos:** Tentar recriar no Restaurante 1.
- **Esperado:** HTTP 409 Conflict violando o índice único `(restaurant_id, code)`.

### FIN-012: Validação de Desconto no Cupom (`PERCENTAGE` vs `FIXED_AMOUNT`)
- **Pré-condição:** Nenhuma.
- **Passos:** Testar os dois tipos de desconto em cupons.
- **Esperado:** Ambos persistidos e validados corretamente.

### FIN-013: Fidelidade: Criação Automática de `user_restaurant` para Novo CPF na Compra
- **Pré-condição:** CPF `52998224725` nunca comprou no Restaurante ID 1 (`user_restaurant` não existe para esse par).
- **Passos:** Cliente informa seu CPF no checkout de compra de 100.00. Concluir pagamento (Regra: 1 moeda a cada R$ 10).
- **Esperado:** Sistema insere automaticamente nova linha em `user_restaurant` com `restaurant_id = 1`, `cpf = '52998224725'`, `coins_amount = 10` e `account_id = NULL`.

### FIN-014: Fidelidade: Incremento de Saldo para CPF Já Existente na Compra
- **Pré-condição:** Registro em `user_restaurant` já existente com `cpf = '52998224725'`, `restaurant_id = 1` e `coins_amount = 10`.
- **Passos:** Cliente compra mais 50.00 no Restaurante ID 1 informando o mesmo CPF.
- **Esperado:** `coins_amount` incrementado para `15` (`10 + 5`).

### FIN-015: Fidelidade: Bloqueio de Resgate de Moedas para Usuário sem Conta (`account_id IS NULL`)
- **Pré-condição:** Registro em `user_restaurant` possui `coins_amount = 50`, mas `account_id = NULL` (usuário apenas informou CPF nas compras, nunca criou conta).
- **Passos:** Tentar aplicar as moedas como desconto no pedido.
- **Esperado:** HTTP 403 Forbidden com mensagem: `an account is required to redeem coins. Please register or log in`.

### FIN-016: Fidelidade: Vínculo de Conta (`account`) à Carteira de CPF Existente
- **Pré-condição:** Carteira em `user_restaurant` existe com `cpf = '52998224725'` e `coins_amount = 50`.
- **Passos:** Usuário cria uma conta `account` com `document = '52998224725'`.
- **Esperado:** Sistema vincula automaticamente `account_id` da conta recém-criada ao registro de `user_restaurant`.

### FIN-017: Fidelidade: Resgate de Moedas com Sucesso para Usuário com Conta (`account_id IS NOT NULL`)
- **Pré-condição:** `user_restaurant` possui `account_id = 5` preenchido e `coins_amount = 50` (cada moeda = R$ 1,00 de desconto). Pedido de 40.00.
- **Passos:** Usuário autenticado solicita resgate de 20 moedas.
- **Esperado:** HTTP 200 OK. Desconto de R$ 20,00 aplicado no pedido. Saldo em `user_restaurant.coins_amount` atualizado para `30`.

### FIN-018: Fidelidade: Bloqueio de Resgate de Moedas Superior ao Saldo
- **Pré-condição:** Usuário com conta possui 10 moedas.
- **Passos:** Tentar resgatar 25 moedas.
- **Esperado:** HTTP 400 Bad Request `insufficient coins balance`.

### FIN-019: Fidelidade: Isolamento de Moedas Entre Restaurantes Distintos
- **Pré-condição:** Cliente tem 50 moedas no Restaurante 1 e 0 moedas no Restaurante 2.
- **Passos:** Tentar resgatar moedas no Restaurante 2.
- **Esperado:** Rejeitado. Moedas são intransferíveis entre diferentes `restaurant_id`.

### FIN-020: Cancelamento de Pagamento (`CANCELLED`)
- **Pré-condição:** Transação pendente.
- **Passos:** PATCH no status da transação para `CANCELLED`.
- **Esperado:** HTTP 200 OK.

### FIN-021: Reversão / Estorno de Moedas Concedidas em Pedido Cancelado
- **Pré-condição:** Compra gerou 10 moedas para o CPF. Pedido foi posteriormente estornado/cancelado.
- **Passos:** Executar rotina de cancelamento.
- **Esperado:** 10 moedas deduzidas da carteira `user_restaurant`, restaurando o saldo anterior.

### FIN-022: Falha em Transação de Cartão de Crédito (`FAILED`)
- **Pré-condição:** Pagamento recusado pela adquirente.
- **Passos:** Registrar retorno da adquirente.
- **Esperado:** `transaction.status` gravado como `FAILED`. Pedido continua aguardando pagamento.

### FIN-023: Validação de Chave Estrangeira: Transação com `order_id` Inexistente
- **Pré-condição:** `order_id = 999999` não cadastrado.
- **Passos:** POST transação.
- **Esperado:** HTTP 400 Bad Request violando chave estrangeira com a tabela `order`.

### FIN-024: Limite Máximo de Usos de Cupom (`coupon.max_uses`)
- **Pré-condição:** Cupom com `max_uses = 3`, já utilizado 3 vezes.
- **Passos:** Quarto pedido tenta usar o cupom.
- **Esperado:** HTTP 400 Bad Request `coupon usage limit reached`.

### FIN-025: Cupom Fora da Validade (`coupon.expires_at`)
- **Pré-condição:** Cupom com data de expiração no passado.
- **Passos:** Aplicar cupom no pedido.
- **Esperado:** HTTP 400 Bad Request `coupon has expired`.

### FIN-026: Desativação de Cupom (`is_active = false`)
- **Pré-condição:** Cupom desativado manualmente.
- **Passos:** Tentar utilizar o cupom.
- **Esperado:** HTTP 400 Bad Request `coupon is inactive`.

### FIN-027: Validação de Valor Mínimo de Pedido no Cupom (`min_order_value`)
- **Pré-condição:** Cupom exige mínimo de 60.00; Pedido com total de 40.00.
- **Passos:** Tentar aplicar.
- **Esperado:** HTTP 400 Bad Request `order amount does not satisfy coupon requirements`.

### FIN-028: Cálculo e Auditoria de Recibo Fiscal em JSONB
- **Pré-condição:** Recibo gravado para pedido finalizado.
- **Passos:** Consultar coluna `receipt.receipt`.
- **Esperado:** Objeto JSON íntegro contendo discriminação de itens, taxas e formas de pagamento.

### FIN-029: Prevenção de Transação com Valor Negativo ou Zero
- **Pré-condição:** Nenhuma.
- **Passos:** POST transação com `amount = -10.00`.
- **Esperado:** HTTP 400 Bad Request `amount must be greater than zero`.

### FIN-030: Concorrência no Resgate de Moedas
- **Pré-condição:** Usuário com conta possui 30 moedas.
- **Passos:** Dois pedidos simultâneos solicitam o resgate de 20 moedas cada.
- **Esperado:** Apenas um resgate conclui com sucesso; o segundo falha por saldo insuficiente (prevenção de double-spending).

---

## Módulo 6: Turnos Operacionais, Concorrência e Resiliência (OPS-001 a OPS-020)

### OPS-001: Abertura de Turno Operacional (`staff_session`) com `shift_code` de 6 Dígitos
- **Pré-condição:** Unidade física ID 1 ativa.
- **Passos:** POST `/branches/1/staff-sessions`.
- **Entradas:** `{"shift_code": "TURNO1", "expires_at": "2026-10-18T00:00:00Z"}`.
- **Esperado:** HTTP 201 Created. Registro inserido em `staff_session` com `qr_token` de 64 caracteres único e `is_active = true`.

### OPS-002: Validação de Tamanho Limite de `shift_code` (`varchar(6)`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter turno com `{"shift_code": "TURNO_LONGO_DEMAIS"}`.
- **Esperado:** HTTP 400 Bad Request `shift_code cannot exceed 6 characters`.

### OPS-003: Unicidade do Token QR de Presença (`staff_session.qr_token`)
- **Pré-condição:** Token `token_hash_64_chars` existente.
- **Passos:** Tentar gravar outro turno com token idêntico.
- **Esperado:** HTTP 409 Conflict violando o índice único em `qr_token`.

### OPS-004: Encerramento de Turno (`is_active = false`)
- **Pré-condição:** Turno ativo.
- **Passos:** PATCH `/staff-sessions/1` com `{"is_active": false}`.
- **Esperado:** HTTP 200 OK. Turno encerrado; novos pontos com este token passam a ser rejeitados.

### OPS-005: Verificação de Validade de Turno Expirado (`expires_at`)
- **Pré-condição:** Turno cuja data `expires_at` é anterior ao horário atual.
- **Passos:** Consultar status do turno.
- **Esperado:** Considerado inativo automaticamente pelo sistema.

### OPS-006: Concorrência na Venda de Última Unidade de Item em Estoque
- **Pré-condição:** Item limitado a 1 unidade.
- **Passos:** Dois clientes finalizam a compra no mesmo milissegundo.
- **Esperado:** Um pedido obtém sucesso (HTTP 201); o segundo é rejeitado por indisponibilidade, sem permitir estoque negativo.

### OPS-007: Concorrência no Acúmulo de Moedas pelo Mesmo CPF
- **Pré-condição:** Cliente compra presencialmente e faz pedido de delivery simultaneamente com o mesmo CPF.
- **Passos:** Disparo paralelo de conclusão de pagamento das duas ordens.
- **Esperado:** Ambas as bonificações são somadas atomicamente sem perda de saldo (`coins_amount`).

### OPS-008: Atomicidade de Pedido com Rollback em Falha de Inserção de Item
- **Pré-condição:** Pedido com 2 itens válidos e 1 item com `item_id` inexistente.
- **Passos:** Enviar criação de pedido.
- **Esperado:** Falha no item inexistente reverte a transação inteira (`tx.Rollback()`). Nenhuma linha é inserida em `order` ou `order_item`.

### OPS-009: Resiliência contra Queda Repentina de Conexão com o PostgreSQL
- **Pré-condição:** Aplicação em execução.
- **Passos:** Derrubar temporariamente o banco de dados.
- **Esperado:** API retorna HTTP 503 Service Unavailable sem encerrar abruptamente o processo Go; reconecta automaticamente após restauração do banco.

### OPS-010: Esgotamento e Recuperação de Conexões do Pool (Connection Pool)
- **Pré-condição:** Limite do pool configurado para 20 conexões.
- **Passos:** Disparar 100 requisições simultâneas de leitura de cardápio.
- **Esperado:** Todas as conexões são devidamente liberadas após execução (`rows.Close()`) sem travamento do pool.

### OPS-011: Validação de Endpoint de Healthcheck (`/health`)
- **Pré-condição:** Banco de dados e aplicação operacionais.
- **Passos:** GET `/health`.
- **Esperado:** HTTP 200 OK com `{"status": "ok", "service": "pratello"}`.

### OPS-012: Healthcheck Reportando Falha de Banco (`/health`)
- **Pré-condição:** Banco de dados indisponível.
- **Passos:** GET `/health`.
- **Esperado:** HTTP 500 ou 503 com status degradado.

### OPS-013: Proteção contra Payload JSON Gigante (DoS de Memória)
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar body com 30MB de lixo textual.
- **Esperado:** HTTP 413 Payload Too Large.

### OPS-014: Ocultação Estrita de Stack Traces e Dados Internos em Erros 500
- **Pré-condição:** Erro de banco forçado.
- **Passos:** Executar chamada afetada.
- **Esperado:** Retorno HTTP 500 contendo apenas mensagem genérica amigável `{"error": "internal server error"}`. Nenhuma query SQL ou caminho de arquivo é exposto.

### OPS-015: Idempotência de Inicialização de Tabelas (`InitSchema`)
- **Pré-condição:** Banco já populado com o schema DBML.
- **Passos:** Invocar rotina de inicialização de schema novamente.
- **Esperado:** Execução sem falhas; nenhuma tabela ou índice duplicado.

### OPS-016: Encerramento Gracioso do Servidor (Graceful Shutdown)
- **Pré-condição:** Servidor processando transação de pagamento.
- **Passos:** Enviar sinal `SIGTERM`.
- **Esperado:** O servidor conclui as requisições em trânsito antes de desligar, sem interromper ou corromper dados de transações.

### OPS-017: Validação de Deadlock em Transações Concorrentes
- **Pré-condição:** Operações concorrentes acessando `order` e `transaction` em ordem cruzada.
- **Passos:** Disparar execução em threads simultâneas.
- **Esperado:** O driver trata a contenção de locks sem travar permanentemente a aplicação.

### OPS-018: Consulta de Sessão de Turno por Token QR
- **Pré-condição:** Turno cadastrado com token de 64 caracteres.
- **Passos:** GET `/staff-sessions/by-token?token=<64_chars>`.
- **Esperado:** HTTP 200 OK com dados da sessão.

### OPS-019: Bloqueio de Registro de Presença em Turno Inativo
- **Pré-condição:** Turno com `is_active = false`.
- **Passos:** Tentar validar presença com o token desse turno.
- **Esperado:** HTTP 400 Bad Request `shift session is inactive`.

### OPS-020: Verificação de Consistência e Isolamento ACID Completo
- **Pré-condição:** Ciclo completo: Pedido local -> Pagamento PIX -> Acúmulo de moedas em `user_restaurant` por CPF.
- **Passos:** Submeter requisições do fluxo completo.
- **Esperado:** Todas as operações comitadas com sucesso e integridade total dos saldos e comprovantes gerados.