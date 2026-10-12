# Plano e Especificação Detalhada de Testes — Pratello

> **Classificação:** Suíte de Testes Profunda
> **Total de Casos de Teste:** 200 
> **Objetivo:** Fornecer uma suíte exaustiva cobrindo Caminho Feliz (*Happy Path*), Condições de Limite (*Boundary Testing*), Entradas Maliciosas e Inválidas (*Negative Testing*), Integridade Referencial (*ACID/FK*), Idempotência e Concorrência para o ecossistema Pratello.

---

## Resumo Executivo e Metodologia de Teste

A arquitetura do Pratello foi dividida em **6 Módulos Coesos de Alta Profundidade (Deep Modules)**, eliminando a dispersão anterior e garantindo que cada endpoint e entidade seja testada com suas variações completas de entrada, restrições SQL e respostas HTTP:

1. **Módulo 1: Identidade, Autenticação e Perfis (IDP)** — Casos `IDP-001` a `IDP-035` (35 testes): Contas globais, hash seguro, validações de formato (CPF, E-mail, Telefone), permissões RBAC (`sysadmin`, `manager`, `users`), sessões e conflitos de unicidade.
2. **Módulo 2: Estabelecimentos, Filiais e Infraestrutura (TEN)** — Casos `TEN-001` a `TEN-030` (30 testes): Gestão de tenants/marcas (`restaurant`), filiais (`physical_restaurant`), padronização postal (`Address`), canais de contato e isolamento multi-tenant estrito.
3. **Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT)** — Casos `CAT-001` a `CAT-045` (45 testes): Layout visual (`menu`), categorias ordenadas, itens vendáveis, ficha técnica de ingredientes (`item_ingredient`), grupos de adicionais/opcionais (`item_side`, `side_option`), receitas de combos (`combo_item`) e promoções temporárias (`item_sale`).
4. **Módulo 4: Pedidos, Especialização (Salão vs. Delivery) e Imutabilidade (ORD)** — Casos `ORD-001` a `ORD-040` (40 testes): Máquina de estados de pedidos, congelamento obrigatório de preço (`unit_price_frozen`), especialização exclusiva (`local_order` vs `remote_order`), cálculo de frete, cancelamentos e consistência de itens hierárquicos.
5. **Módulo 5: Pagamentos, Transações, PIX e Fidelidade (FIN)** — Casos `FIN-001` a `FIN-030` (30 testes): Métodos de pagamento, transações parciais/divididas, idempotência de geração de PIX (`transaction_pix`), cupons de desconto (`coupon`), carteira de cashback/coins (`user_restaurant`) e recibos fiscais imutáveis (`receipt`).
6. **Módulo 6: Turnos Operacionais, Concorrência e Resiliência (OPS)** — Casos `OPS-001` a `OPS-020` (20 testes): Abertura e fechamento de turnos (`restaurant_staff_session`), validação de token QR de presença (`staff_session`), concorrência em pedidos simultâneos, lock pessimista/otimista e recuperação de falhas.

---

## Módulo 1: Identidade, Autenticação e Perfis (IDP-001 a IDP-035)

### IDP-001: Cadastro de Conta com E-mail e Senha Válidos
- **Pré-condição:** E-mail `admin@pratello.com` inexistente no banco.
- **Passos:** Enviar requisição POST para `/auth/register`. Consultar banco `SELECT id, email, password_hash, created_at FROM account WHERE email = ?`.
- **Entradas:**
  - Válida padrão: `{"email": "admin@pratello.com", "password": "Password123!", "phone": "11987654321"}`
  - Válida limite mínimo de senha (8 chars): `{"email": "admin2@pratello.com", "password": "Abc1234!", "phone": "11987654322"}`
  - Válida e-mail com subdomínio: `{"email": "dev.test+qa@corp.pratello.com.br", "password": "Password123!", "phone": "11987654323"}`
- **Esperado:** HTTP 201 Created. Senha nunca gravada em texto plano (`password_hash` com bcrypt/argon2). `created_at` preenchido.
- **Critério de Aceite:** O hash decodificado confere com a senha via algoritmo de verificação.

### IDP-002: Rejeição de E-mail Duplicado (Unique Constraint)
- **Pré-condição:** Conta com e-mail `owner@bela.com` já persistida.
- **Passos:** Tentar cadastrar nova conta com variações de caixa alta/baixa do mesmo e-mail.
- **Entradas:**
  - Exato: `{"email": "owner@bela.com", "password": "DiffPassword99#"}`
  - Case insensitive: `{"email": "OWNER@BELA.COM", "password": "DiffPassword99#"}`
  - Com espaços nas pontas: `{"email": "  owner@bela.com  ", "password": "DiffPassword99#"}`
- **Esperado:** HTTP 409 Conflict. Erro `email already registered`.
- **Critério de Aceite:** Nenhuma linha adicional inserida na tabela `account`.

### IDP-003: Validação de Formato de E-mail Inválido
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/auth/register` com e-mails malformados.
- **Entradas:**
  - Sem `@`: `"email": "usuario-sem-arroba.com"`
  - Sem domínio: `"email": "usuario@"`
  - Sem TLD: `"email": "usuario@dominio"`
  - Caracteres proibidos: `"email": "user name@dominio.com"`
  - Vazio: `"email": ""`
- **Esperado:** HTTP 400 Bad Request. Campo `email` apontado no payload de erro.

### IDP-004: Políticas de Complexidade e Tamanho de Senha
- **Pré-condição:** Nenhuma.
- **Passos:** Testar senhas fora da política de segurança.
- **Entradas:**
  - Menor que 8 chars: `{"password": "Ab1!"}`
  - Sem números: `{"password": "Abcdefgh!"}`
  - Sem letras maiúsculas: `{"password": "abcdefgh1!"}`
  - Sem símbolos: `{"password": "Abcdefgh1"}`
  - Gigante (> 128 chars): `{"password": "<string com 256 caracteres>"}`
- **Esperado:** HTTP 400 Bad Request com mensagem detalhando a regra violada.

### IDP-005: Autenticação / Login com Credenciais Válidas
- **Pré-condição:** Conta `user@pratello.com` com senha `SecretPass@1`.
- **Passos:** POST `/auth/login`.
- **Entradas:**
  - `{"email": "user@pratello.com", "password": "SecretPass@1"}`
  - Case-insensitive e-mail: `{"email": "USER@PRATELLO.COM", "password": "SecretPass@1"}`
- **Esperado:** HTTP 200 OK. Resposta contendo token JWT válido com claims (`sub: account_id`, `exp`) e objeto `account` (sem campo de hash).

### IDP-006: Rejeição de Login com Credenciais Inválidas (Anti-Enumeração)
- **Pré-condição:** Conta `user@pratello.com` existente.
- **Passos:** POST `/auth/login` com senha incorreta ou usuário inexistente.
- **Entradas:**
  - Senha errada: `{"email": "user@pratello.com", "password": "WrongPassword#9"}`
  - E-mail inexistente: `{"email": "inexistente@pratello.com", "password": "SecretPass@1"}`
  - Payload vazio: `{}`
- **Esperado:** HTTP 401 Unauthorized com mensagem genérica: `invalid email or password` (sem revelar se o e-mail existe).

### IDP-007: Promoção e Verificação de Superadministrador (sysadmin)
- **Pré-condição:** Conta `acc_01` criada sem privilégios.
- **Passos:** Superadministrador executa POST `/sysadmins` vinculando `acc_01`. Consultar tabela `sysadmin`.
- **Entradas:** `{"account_id": 1}`.
- **Esperado:** HTTP 201 Created. Registro inserido com `account_id = 1` e `created_at`.
- **Critério de Aceite:** `acc_01` agora obtém escopo de permissão global sobre todos os restaurantes.

### IDP-008: Unicidade de Registro de Superadministrador
- **Pré-condição:** `acc_01` já registrado na tabela `sysadmin`.
- **Passos:** Tentar promover novamente a mesma conta.
- **Entradas:** `{"account_id": 1}`.
- **Esperado:** HTTP 409 Conflict. Chave única violada em `sysadmin.account_id`.

### IDP-009: Associação de Gerente a Restaurante (manager)
- **Pré-condição:** `account_id = 2` existente e restaurante `slug = 'pizzaria-bela'` existente.
- **Passos:** POST `/restaurants/pizzaria-bela/managers`.
- **Entradas:** `{"account_id": 2}`.
- **Esperado:** HTTP 201 Created. Registro criado na tabela `manager` associando as duas entidades.

### IDP-010: Tentativa de Associar Gerente a Restaurante Inexistente
- **Pré-condição:** `account_id = 2` existente, slug `'restaurante-fantasma'` não existe.
- **Passos:** POST `/restaurants/restaurante-fantasma/managers` com `{"account_id": 2}`.
- **Esperado:** HTTP 404 Not Found (ou violação de Foreign Key no banco).

### IDP-011: Unicidade do Vínculo Gerente × Restaurante
- **Pré-condição:** `account_id = 2` já é gerente de `'pizzaria-bela'`.
- **Passos:** Submeter novamente POST com os mesmos parâmetros.
- **Esperado:** HTTP 409 Conflict. Violação da restrição `UNIQUE(restaurant_slug, account_id)`.

### IDP-012: Suporte a Multi-Gerência (Mesma Conta em Múltiplos Restaurantes)
- **Pré-condição:** `account_id = 2` é gerente de `'pizzaria-bela'`. Restaurante `'hamburgueria-bela'` cadastrado.
- **Passos:** POST `/restaurants/hamburgueria-bela/managers` com `{"account_id": 2}`.
- **Esperado:** HTTP 201 Created. A mesma conta pode gerenciar múltiplos estabelecimentos legítimos.

### IDP-013: Desvinculação de Gerente de Restaurante
- **Pré-condição:** Vínculo existente entre `account_id = 2` e `'pizzaria-bela'`.
- **Passos:** DELETE `/restaurants/pizzaria-bela/managers/2`.
- **Esperado:** HTTP 204 No Content. Registro removido da tabela `manager`. A conta `account` continua existindo no sistema.

### IDP-014: Cadastro de Consumidor Final (users) com CPF Válido
- **Pré-condição:** CPF sem registro prévio.
- **Passos:** POST `/users`.
- **Entradas:**
  - CPF sem formatação: `{"cpf": "52998224725", "name": "Carlos Silva", "phone": "81999991111"}`
  - CPF com pontuação: `{"cpf": "529.982.247-25", "name": "Carlos Silva", "phone": "81999991111"}`
- **Esperado:** HTTP 201 Created. Banco normaliza e grava rigorosamente 11 dígitos numéricos (`52998224725`).

### IDP-015: Rejeição de CPF com Dígito Verificador Inválido (Algoritmo Módulo 11)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter CPFs matematicamente inválidos para POST `/users`.
- **Entradas:**
  - Dígito final errado: `{"cpf": "52998224720", "name": "Teste", "phone": "81999990000"}`
  - Todos dígitos iguais: `{"cpf": "11111111111", "name": "Teste", "phone": "81999990000"}`
  - Menos de 11 dígitos: `{"cpf": "12345678", "name": "Teste", "phone": "81999990000"}`
  - Com letras: `{"cpf": "5299822472A", "name": "Teste", "phone": "81999990000"}`
- **Esperado:** HTTP 400 Bad Request. `invalid CPF verification digits`.

### IDP-016: Rejeição de CPF Duplicado (Chave Primária de users)
- **Pré-condição:** CPF `52998224725` cadastrado para "Carlos Silva".
- **Passos:** Submeter novo cadastro com o mesmo CPF.
- **Entradas:** `{"cpf": "52998224725", "name": "Outro Nome", "phone": "11988887777"}`.
- **Esperado:** HTTP 409 Conflict. Dados originais de Carlos Silva intactos.

### IDP-017: Consulta de Cliente por CPF
- **Pré-condição:** Cliente com CPF `52998224725` existente.
- **Passos:** GET `/users/52998224725`.
- **Esperado:** HTTP 200 OK retornando objeto com `cpf`, `name`, `phone` e `created_at`.

### IDP-018: Consulta de Cliente por CPF Inexistente
- **Pré-condição:** CPF `99999999999` não cadastrado.
- **Passos:** GET `/users/99999999999`.
- **Esperado:** HTTP 404 Not Found com payload `{"error": "user not found"}`.

### IDP-019: Atualização de Dados Cadastrais do Cliente (Nome e Telefone)
- **Pré-condição:** Cliente `52998224725` ("Carlos Silva").
- **Passos:** PUT `/users/52998224725` com novos dados.
- **Entradas:** `{"name": "Carlos Silva Santos", "phone": "81988880000"}`.
- **Esperado:** HTTP 200 OK. Registro alterado. O CPF permanece imutável.

### IDP-020: Sanitização de Nome de Cliente (Valores Limite)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter PUT/POST com variações de nome.
- **Entradas:**
  - Limite mínimo (1 caractere): `{"name": "A"}` -> HTTP 400 (mínimo 2 caracteres exigidos).
  - Limite máximo (150 caracteres): `{"name": "<150 caracteres>"}` -> HTTP 200.
  - Acima do limite (151 caracteres): `{"name": "<151 caracteres>"}` -> HTTP 400.
  - Apenas espaços: `{"name": "    "}` -> HTTP 400.
- **Esperado:** Validação estrita em conformidade com as regras de schema.

### IDP-021: Formatação e Validação de Número de Telefone
- **Pré-condição:** Nenhuma.
- **Passos:** Testar formatos de telefone no cadastro de usuários e contas.
- **Entradas:**
  - Padrão DDD + 9 dígitos: `"81999998888"` -> Válido (HTTP 200/201).
  - Padrão com código de país: `"+5581999998888"` -> Normalizado para 5581999998888.
  - Fixo DDD + 8 dígitos: `"8133334444"` -> Válido.
  - Telefone curto: `"12345"` -> HTTP 400.
  - Telefone com letras: `"8199999ABCD"` -> HTTP 400.
- **Esperado:** Apenas telefones normalizados são persistidos.

### IDP-022: Obtenção de Perfil Autenticado (`/auth/me`)
- **Pré-condição:** Token JWT válido emitido para `account_id = 5`.
- **Passos:** GET `/auth/me` com header `Authorization: Bearer <token>`.
- **Esperado:** HTTP 200 OK contendo dados da conta e lista de restaurantes gerenciados.

### IDP-023: Rejeição de Token JWT Expirado
- **Pré-condição:** Token assinado com timestamp de expiração no passado (`exp < now()`).
- **Passos:** GET `/auth/me` com token expirado.
- **Esperado:** HTTP 401 Unauthorized com `token has expired`.

### IDP-024: Rejeição de Token JWT com Assinatura Inválida
- **Pré-condição:** Token alterado manualmente no payload (violação do HMAC).
- **Passos:** Requisição com token corrompido.
- **Esperado:** HTTP 401 Unauthorized com `invalid token signature`.

### IDP-025: Controle de Acesso: Bloqueio de Usuário Comum em Rota de Sysadmin
- **Pré-condição:** Usuário autenticado com perfil apenas de `manager`.
- **Passos:** Tentar POST `/sysadmins` para promover outra conta.
- **Esperado:** HTTP 403 Forbidden.

### IDP-026: Controle de Acesso: Gerente Acessando Restaurante de Terceiro
- **Pré-condição:** Gerente `acc_02` vinculado exclusivamente ao restaurante `slug = 'pizzaria-a'`.
- **Passos:** `acc_02` tenta atualizar cardápio de `slug = 'pizzaria-b'` via PUT `/restaurants/pizzaria-b/menu`.
- **Esperado:** HTTP 403 Forbidden com mensagem de tenant não autorizado.

### IDP-027: Logout e Invalidação de Sessão / Token Blacklist
- **Pré-condição:** Token válido.
- **Passos:** Executar POST `/auth/logout`. Tentar reutilizar o mesmo token imediatamente.
- **Esperado:** Primeira chamada HTTP 200 OK. Segunda chamada HTTP 401 Unauthorized.

### IDP-028: Alteração de Senha com Verificação de Senha Atual
- **Pré-condição:** Conta autenticada com senha `OldPass@123`.
- **Passos:** POST `/auth/change-password`.
- **Entradas:**
  - Incorreta atual: `{"current_password": "Wrong", "new_password": "NewPass@123"}` -> HTTP 400.
  - Correta atual: `{"current_password": "OldPass@123", "new_password": "NewPass@123"}` -> HTTP 200.
- **Esperado:** Hash atualizado no banco. Antiga senha não autentica mais.

### IDP-029: Prevenção de SQL Injection no Login e Consultas de Usuário
- **Pré-condição:** Tabela `account` populada.
- **Passos:** POST `/auth/login` com injeções SQL clássicas nos campos de entrada.
- **Entradas:**
  - `{"email": "' OR '1'='1", "password": "' OR '1'='1"}`
  - `{"email": "admin@pratello.com'; DROP TABLE account; --", "password": "123"}`
- **Esperado:** HTTP 401 Unauthorized. Nenhuma instrução SQL maliciosa executada (garantido por queries parametrizadas).

### IDP-030: Tratamento de Caracteres Especiais e UTF-8 em Nomes
- **Pré-condição:** Nenhuma.
- **Passos:** Cadastrar usuário com acentos, cedilha, emojis e alfabetos especiais.
- **Entradas:** `{"name": "João Peña da Silva & Filho 🍕", "cpf": "52998224725", "phone": "11999990000"}`.
- **Esperado:** HTTP 201 Created. Armazenado e recuperado com encoding UTF-8 preservado.

### IDP-031: Proteção contra Cross-Site Scripting (XSS) em Campos Textuais
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar payloads com scripts em campos de texto de usuário.
- **Entradas:** `{"name": "<script>alert('xss')</script>Carlos", "cpf": "52998224725", "phone": "11999990000"}`.
- **Esperado:** Strings sanitizadas ou codificadas sem execução no retorno JSON.

### IDP-032: Limite de Requisições de Login (Rate Limiting contra Brute-Force)
- **Pré-condição:** Conta existente.
- **Passos:** Disparar 15 requisições de login falhas consecutivas em menos de 10 segundos para o mesmo IP/e-mail.
- **Esperado:** As primeiras 5 retornam HTTP 401; a partir da 6ª retorna HTTP 429 Too Many Requests com header `Retry-After`.

### IDP-033: Deleção Lógica / Anonimização de Consumidor (LGPD / GDPR)
- **Pré-condição:** Cliente com CPF cadastrado e histórico de pedidos anteriores.
- **Passos:** Solicitar exclusão do perfil do usuário.
- **Esperado:** Dados pessoais (`name`, `phone`) anonimizados para `"Anonimizado"`. CPF desvinculado ou mantido apenas hash irreversível para auditoria fiscal sem expor o titular.

### IDP-034: Tentativa de Criar Conta com Payload JSON Corrompido
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/auth/register` com corpo truncado ou sem fechar chaves (`{"email": "a@b.com"`).
- **Esperado:** HTTP 400 Bad Request com `invalid JSON syntax`.

### IDP-035: Recuperação de Senha com Token de Vida Curta
- **Pré-condição:** E-mail cadastrado.
- **Passos:** POST `/auth/forgot-password`. Receber token de redefinição com validade de 15 minutos. Usar após 16 minutos.
- **Esperado:** HTTP 400 Bad Request `password reset token expired`.

---

## Módulo 2: Estabelecimentos, Filiais e Infraestrutura (TEN-001 a TEN-030)

### TEN-001: Criação de Restaurante (Tenant) com Slug Automático
- **Pré-condição:** Slug derivado inexistente.
- **Passos:** POST `/restaurants` omitindo o campo `slug`.
- **Entradas:** `{"name": "Pizzaria Bela Napoli", "description": "Autêntica pizza italiana"}`.
- **Esperado:** HTTP 201 Created. O slug é gerado automaticamente em minúsculas e kebab-case: `"pizzaria-bela-napoli"`.

### TEN-002: Criação de Restaurante com Slug Manual Válido
- **Pré-condição:** Slug `bela-napoli` disponível.
- **Passos:** POST `/restaurants`.
- **Entradas:** `{"name": "Bela Napoli", "slug": "bela-napoli", "description": "Massas e pizzas"}`.
- **Esperado:** HTTP 201 Created. Chave primária gravada exatamente como `bela-napoli`.

### TEN-003: Validação de Formato de Slug (Regex e Caracteres Proibidos)
- **Pré-condição:** Nenhuma.
- **Passos:** Tentar criar restaurantes com slugs que violem regras de URL.
- **Entradas:**
  - Com espaços: `"slug": "pizzaria bela"` -> HTTP 400.
  - Com caracteres especiais: `"slug": "pizzaria@bela!"` -> HTTP 400.
  - Com letras maiúsculas: `"slug": "PizzariaBela"` -> Normalizado para `pizzariabela` ou HTTP 400.
  - Menor que 3 caracteres: `"slug": "ab"` -> HTTP 400.
- **Esperado:** Apenas slugs compatíveis com URLs amigáveis (`^[a-z0-9-]+$`) são aceitos.

### TEN-004: Conflito de Slug Existente (Unique PK)
- **Pré-condição:** Restaurante com slug `pratello-burger` já criado.
- **Passos:** Submeter POST `/restaurants` com mesmo slug.
- **Entradas:** `{"name": "Outro Pratello Burger", "slug": "pratello-burger"}`.
- **Esperado:** HTTP 409 Conflict. Mensagem: `restaurant with this slug already exists`.

### TEN-005: Consulta de Restaurante por Slug
- **Pré-condição:** Restaurante `pratello-burger` cadastrado.
- **Passos:** GET `/restaurants/pratello-burger`.
- **Esperado:** HTTP 200 OK contendo `slug`, `name`, `description`, `created_at`.

### TEN-006: Consulta de Restaurante Inexistente
- **Pré-condição:** Slug `nao-existe` não cadastrado.
- **Passos:** GET `/restaurants/nao-existe`.
- **Esperado:** HTTP 404 Not Found com payload padronizado `{"error": "restaurant not found"}`.

### TEN-007: Atualização Completa de Restaurante
- **Pré-condição:** Restaurante `bela-napoli` existente.
- **Passos:** PUT `/restaurants/bela-napoli`.
- **Entradas:** `{"name": "Bela Napoli Trattoria", "description": "Nova descrição refinada"}`.
- **Esperado:** HTTP 200 OK. Novos dados persistidos. O slug continua inalterado.

### TEN-008: Tentativa de Modificar Slug via PUT
- **Pré-condição:** Restaurante `bela-napoli` existente.
- **Passos:** Enviar body contendo `"slug": "novo-slug"` no PUT `/restaurants/bela-napoli`.
- **Esperado:** O endpoint ignora a alteração do slug ou retorna erro; a chave primária permanece estritamente `bela-napoli`.

### TEN-009: Deleção de Restaurante e Cascade Referencial
- **Pré-condição:** Restaurante com cardápios, filiais e categorias associadas.
- **Passos:** DELETE `/restaurants/slug-antigo`.
- **Esperado:** HTTP 204 No Content. Ou desativação lógica (soft delete) ou exclusão em cascata controlada, sem deixar registros órfãos no banco de dados.

### TEN-010: Criação de Filial Física (`physical_restaurant`) com Endereço
- **Pré-condição:** Restaurante `pizzaria-bela` existente.
- **Passos:** POST `/restaurants/pizzaria-bela/branches`.
- **Entradas:**
  ```json
  {
    "name": "Unidade Boa Viagem",
    "address": {
      "cep": "51020010",
      "street": "Av. Boa Viagem",
      "number": "500",
      "apt": "Loja 02",
      "complement": "Em frente ao mar"
    }
  }
  ```
- **Esperado:** HTTP 201 Created. Registro inserido em `Address` e chave estrangeira vinculada em `physical_restaurant.address_id`.

### TEN-011: Validação de Formato e Obrigatoriedade de CEP no Endereço
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter endereços com CEPs inválidos.
- **Entradas:**
  - Com 7 dígitos: `"cep": "5102001"` -> HTTP 400.
  - Com 9 dígitos: `"cep": "510200100"` -> HTTP 400.
  - Com letras: `"cep": "5102001A"` -> HTTP 400.
  - Com máscara: `"cep": "51020-010"` -> Normalizado para `51020010`.
- **Esperado:** Apenas 8 caracteres numéricos são persistidos na coluna `Address.cep`.

### TEN-012: Listagem de Filiais por Restaurante
- **Pré-condição:** Restaurante com 3 filiais cadastradas.
- **Passos:** GET `/restaurants/pizzaria-bela/branches`.
- **Esperado:** HTTP 200 OK com array de 3 objetos, cada qual contendo seu endereço completo aninhado.

### TEN-013: Consulta de Filial por ID
- **Pré-condição:** Filial ID 10 cadastrada.
- **Passos:** GET `/restaurants/pizzaria-bela/branches/10`.
- **Esperado:** HTTP 200 OK com detalhes da filial.

### TEN-014: Consulta de Filial Pertencente a Outro Restaurante (Tenant Mismatch)
- **Pré-condição:** Filial ID 10 pertence a `pizzaria-bela`.
- **Passos:** GET `/restaurants/hamburgueria-x/branches/10`.
- **Esperado:** HTTP 404 Not Found (ou 403 Forbidden). A rota não deve vazar dados de outro tenant.

### TEN-015: Atualização de Endereço de Filial Existente
- **Pré-condição:** Filial com endereço cadastrado.
- **Passos:** PUT `/restaurants/pizzaria-bela/branches/10/address`.
- **Entradas:** `{"cep": "52050000", "street": "Rua Nova", "number": "123", "apt": null, "complement": null}`.
- **Esperado:** HTTP 200 OK. Tabela `Address` atualizada com os novos valores.

### TEN-016: Cadastro de Telefone de Filial (`p_restaurant_phone`)
- **Pré-condição:** Filial ID 10 existente.
- **Passos:** POST `/restaurants/pizzaria-bela/branches/10/phones`.
- **Entradas:**
  - Válido WhatsApp: `{"phone_number": "81999990001", "phone_type": "whatsapp"}`
  - Válido Fixo: `{"phone_number": "8134567890", "phone_type": "landline"}`
- **Esperado:** HTTP 201 Created. Registros inseridos com chave estrangeira `branch_id = 10`.

### TEN-017: Validação de Tipo de Telefone (`phone_type`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter tipos de telefone não permitidos.
- **Entradas:** `{"phone_number": "81999990001", "phone_type": "invalido_tipo"}`.
- **Esperado:** HTTP 400 Bad Request informando tipos aceitos (`whatsapp`, `landline`, `delivery`).

### TEN-018: Remoção de Telefone de Filial
- **Pré-condição:** Telefone ID 5 vinculado à filial ID 10.
- **Passos:** DELETE `/restaurants/pizzaria-bela/branches/10/phones/5`.
- **Esperado:** HTTP 204 No Content. Registro excluído de `p_restaurant_phone`.

### TEN-019: Listagem de Todos os Telefones de uma Filial
- **Pré-condição:** Filial com 2 telefones.
- **Passos:** GET `/restaurants/pizzaria-bela/branches/10/phones`.
- **Esperado:** HTTP 200 OK retornando array com os dois números cadastrados.

### TEN-020: Desativação de Filial Física
- **Pré-condição:** Filial ID 10 ativa.
- **Passos:** DELETE `/restaurants/pizzaria-bela/branches/10`.
- **Esperado:** HTTP 204 No Content. Filial marcada como inativa ou removida sem afetar os dados da matriz do restaurante.

### TEN-021: Múltiplas Filiais com Nomes Iguais em Cidades Diferentes
- **Pré-condição:** Filial "Centro" cadastrada no Recife.
- **Passos:** Cadastrar outra filial "Centro" em Olinda para o mesmo restaurante.
- **Esperado:** HTTP 201 Created. Permitido, pois o identificador único é o `id` da filial e os endereços são distintos.

### TEN-022: Atualização de Dados Comerciais da Filial
- **Pré-condição:** Filial ID 10 existente.
- **Passos:** PUT `/restaurants/pizzaria-bela/branches/10`.
- **Entradas:** `{"name": "Unidade Boa Viagem - Reformada"}`.
- **Esperado:** HTTP 200 OK com o novo nome refletido na consulta.

### TEN-023: Prevenção de Criação de Filial com Dados de Endereço Faltantes
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/restaurants/pizzaria-bela/branches` omitindo campos obrigatórios de endereço (`street` ou `number`).
- **Entradas:** `{"name": "Filial Incompleta", "address": {"cep": "51020010"}}`.
- **Esperado:** HTTP 400 Bad Request detalhando `number is required` e `street is required`.

### TEN-024: Listagem Paginada de Restaurantes da Plataforma
- **Pré-condição:** 50 restaurantes cadastrados na plataforma.
- **Passos:** GET `/restaurants?page=1&limit=20` e GET `/restaurants?page=2&limit=20`.
- **Esperado:** Cada requisição devolve exatamente 20 itens ordenados, com metadados de paginação (`total: 50`, `page: 1`, `pages: 3`).

### TEN-025: Busca de Restaurante por Termo / Nome (Full-text Search)
- **Pré-condição:** Restaurantes "Pizzaria Bela Napoli" e "Hamburgueria Bela Vista".
- **Passos:** GET `/restaurants?search=Napoli`.
- **Esperado:** HTTP 200 OK retornando apenas "Pizzaria Bela Napoli".

### TEN-026: Tentativa de Deletar Filial com Pedidos Ativos
- **Pré-condição:** Filial ID 10 possui comanda em status `preparing`.
- **Passos:** DELETE `/restaurants/pizzaria-bela/branches/10`.
- **Esperado:** HTTP 409 Conflict. Ação bloqueada enquanto houver pedidos pendentes de conclusão.

### TEN-027: Associação de Coordenadas Geográficas (Latitude e Longitude no Endereço)
- **Pré-condição:** Tabela `Address` estendida com suporte geoespacial.
- **Passos:** Salvar filial com coordenadas: `{"latitude": -8.123456, "longitude": -34.890123}`.
- **Esperado:** Coordenadas armazenadas para cálculo de raio de entrega de delivery.

### TEN-028: Isolamento de Tenants na Listagem de Filiais
- **Pré-condição:** Tenant A tem 2 filiais, Tenant B tem 3 filiais.
- **Passos:** Consultar filiais de Tenant A.
- **Esperado:** Retorno contém estritamente as 2 filiais de Tenant A, sem vazamento de dados de Tenant B.

### TEN-029: Inclusão de Telefone com Número Vazio ou Nulo
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/branches/1/phones` com `{"phone_number": ""}`.
- **Esperado:** HTTP 400 Bad Request `phone_number cannot be blank`.

### TEN-030: Remoção de Endereço em Cascata com Exclusão da Filial
- **Pré-condição:** Filial 15 associada ao `Address` 80.
- **Passos:** Excluir filial 15. Consultar `SELECT * FROM Address WHERE id = 80`.
- **Esperado:** Endereço da filial removido ou mantido histórico sem quebra de integridade referencial.

---

## Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT-001 a CAT-045)

### CAT-001: Obtenção de Configurações Visuais do Cardápio (`menu`)
- **Pré-condição:** Restaurante `pizzaria-bela` com menu padrão inicializado.
- **Passos:** GET `/restaurants/pizzaria-bela/menu`.
- **Esperado:** HTTP 200 OK com `name`, `template_style`, `accent_color` (hex) e `typography`.

### CAT-002: Atualização de Identidade Visual do Cardápio
- **Pré-condição:** Menu existente.
- **Passos:** PUT `/restaurants/pizzaria-bela/menu`.
- **Entradas:** `{"name": "Cardápio Principal", "template_style": "modern", "accent_color": "#FF5722", "typography": "Poppins"}`.
- **Esperado:** HTTP 200 OK. Novos dados visuais salvos.

### CAT-003: Validação de Cor de Destaque em Hexadecimal (`accent_color`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter cores inválidas no PUT `/menu`.
- **Entradas:**
  - Sem tralha: `"accent_color": "FF5722"` -> HTTP 400.
  - Com 5 caracteres: `"accent_color": "#FFF55"` -> HTTP 400.
  - Com letras inválidas: `"accent_color": "#ZZZZZZ"` -> HTTP 400.
  - Válido 6 dígitos: `"accent_color": "#00AAFF"` -> HTTP 200.
- **Esperado:** Apenas padrões regex `^#[0-9A-Fa-f]{6}$` aceitos.

### CAT-004: Criação de Categoria com Ordem de Exibição
- **Pré-condição:** Nenhuma categoria cadastrada.
- **Passos:** POST `/restaurants/pizzaria-bela/categories`.
- **Entradas:**
  - Categoria 1: `{"name": "Pizzas Salgadas", "display_order": 1}`
  - Categoria 2: `{"name": "Bebidas", "display_order": 2}`
- **Esperado:** HTTP 201 Created para ambas. Ordenação preservada.

### CAT-005: Listagem Ordenada de Categorias
- **Pré-condição:** Categorias cadastradas com `display_order` 3, 1 e 2.
- **Passos:** GET `/restaurants/pizzaria-bela/categories`.
- **Esperado:** HTTP 200 OK com array ordenado estritamente por `display_order ASC`.

### CAT-006: Atualização de Ordem de Exibição de Categoria
- **Pré-condição:** Categoria "Sobremesas" com ordem 10.
- **Passos:** PUT `/restaurants/pizzaria-bela/categories/5` com `{"name": "Sobremesas", "display_order": 1}`.
- **Esperado:** HTTP 200 OK. Categoria movida para o topo.

### CAT-007: Exclusão de Categoria Vazia
- **Pré-condição:** Categoria ID 8 sem nenhum produto associado.
- **Passos:** DELETE `/restaurants/pizzaria-bela/categories/8`.
- **Esperado:** HTTP 204 No Content. Categoria removida.

### CAT-008: Bloqueio de Exclusão de Categoria com Produtos Vinculados
- **Pré-condição:** Categoria ID 1 ("Pizzas") possui 10 itens ativos.
- **Passos:** DELETE `/restaurants/pizzaria-bela/categories/1`.
- **Esperado:** HTTP 409 Conflict. Mensagem: `cannot delete category containing active items`.

### CAT-009: Criação de Item do Cardápio (`menu_item`)
- **Pré-condição:** Categoria ID 1 existente.
- **Passos:** POST `/restaurants/pizzaria-bela/items`.
- **Entradas:**
  ```json
  {
    "category_id": 1,
    "name": "Pizza Margherita",
    "description": "Molho de tomate artesanal, mussarela de búfala e manjericão",
    "base_price": 45.00,
    "photo_url": "https://img.pratello.com/margherita.jpg",
    "is_combo": false,
    "is_available": true
  }
  ```
- **Esperado:** HTTP 201 Created. Registro criado com ID numérico auto-incremental.

### CAT-010: Validação de Preço Base Negativo ou Zero
- **Pré-condição:** Nenhuma.
- **Passos:** Tentar cadastrar itens com preços anômalos.
- **Entradas:**
  - Negativo: `{"base_price": -10.00}` -> HTTP 400 Bad Request.
  - Zero: `{"base_price": 0.00}` -> Permitido se for cortesia ou brinde, ou HTTP 400 se a regra de negócio exigir preço estritamente positivo (> 0.00).
- **Esperado:** Erro de validação retornado quando `base_price < 0`.

### CAT-011: Validação de Precisão Decimal de Preço (Máximo 2 Casas Decimais)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter item com `{"base_price": 19.999}`.
- **Esperado:** HTTP 400 Bad Request ou arredondamento formal para `20.00` conforme contrato da API.

### CAT-012: Atualização de Disponibilidade de Item (`is_available = false`)
- **Pré-condição:** Item "Pizza Margherita" ativo.
- **Passos:** PUT `/restaurants/pizzaria-bela/items/1` com `{"is_available": false}`.
- **Esperado:** HTTP 200 OK. Item não deve aparecer no cardápio público do cliente, mas continua visível no painel administrativo.

### CAT-013: Cadastro de Ingrediente Base no Catálogo (`Ingredient_item`)
- **Pré-condição:** Restaurante `pizzaria-bela`.
- **Passos:** POST `/restaurants/pizzaria-bela/ingredients`.
- **Entradas:**
  - `{"name": "Mussarela de Búfala"}`
  - `{"name": "Cebola Roxa"}`
  - `{"name": "Bacon Crocante"}`
- **Esperado:** HTTP 201 Created para cada um. Ingredientes cadastrados na tabela `Ingredient_item`.

### CAT-014: Unicidade de Nome de Ingrediente por Restaurante
- **Pré-condição:** Ingrediente "Bacon Crocante" já existe.
- **Passos:** Submeter novamente `{"name": "Bacon Crocante"}`.
- **Esperado:** HTTP 409 Conflict informando ingrediente duplicado para este tenant.

### CAT-015: Associação de Ingrediente a Item com Regra de Remoção (`item_ingredient`)
- **Pré-condição:** Item "Pizza Margherita" e Ingrediente "Cebola Roxa" existentes.
- **Passos:** POST `/restaurants/pizzaria-bela/items/1/ingredients`.
- **Entradas:** `{"ingredient_id": 2, "proportion": "50g", "is_removable": true}`.
- **Esperado:** HTTP 201 Created. Registro inserido com chave composta `(item_id, ingredient_id)`.

### CAT-016: Ingrediente Não Removível pelo Cliente (`is_removable = false`)
- **Pré-condição:** Item "Molho Secreto" configurado com `is_removable = false`.
- **Passos:** Tentar fazer pedido solicitando remoção do ingrediente.
- **Esperado:** Pedido rejeitado com erro: `ingredient is not removable`.

### CAT-017: Desvinculação de Ingrediente de Item
- **Pré-condição:** Ingrediente vinculado.
- **Passos:** DELETE `/restaurants/pizzaria-bela/items/1/ingredients/2`.
- **Esperado:** HTTP 204 No Content. Ingrediente desvinculado da ficha técnica do produto.

### CAT-018: Criação de Grupo de Opcionais / Adicionais (`item_side`)
- **Pré-condição:** Item "X-Burger Artesanal" existente.
- **Passos:** POST `/restaurants/pizzaria-bela/items/1/sides`.
- **Entradas:**
  - Escolha obrigatória única: `{"title": "Ponto da Carne", "min_options": 1, "max_options": 1}`
  - Escolha opcional múltipla: `{"title": "Adicionais Pagos", "min_options": 0, "max_options": 4}`
- **Esperado:** HTTP 201 Created para ambos.

### CAT-019: Validação de Invariante de Quantidade em `item_side` (`min <= max`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter grupo com `{"min_options": 3, "max_options": 1}`.
- **Esperado:** HTTP 400 Bad Request `min_options cannot be greater than max_options`.

### CAT-020: Cadastro de Opções Dentro de Grupo de Adicionais (`side_option`)
- **Pré-condição:** Grupo "Ponto da Carne" (ID 1) e "Adicionais Pagos" (ID 2).
- **Passos:** POST nas opções de cada grupo.
- **Entradas:**
  - Opção gratuita: `{"name": "Ao Ponto", "additional_price": 0.00}` no grupo 1.
  - Opção gratuita: `{"name": "Bem Passado", "additional_price": 0.00}` no grupo 1.
  - Opção paga: `{"name": "Bacon Extra", "additional_price": 5.00}` no grupo 2.
- **Esperado:** HTTP 201 Created com preços associados a cada opção.

### CAT-021: Prevenção de Preço Negativo em `side_option`
- **Pré-condição:** Grupo de opcionais existente.
- **Passos:** POST com `{"name": "Desconto", "additional_price": -2.00}`.
- **Esperado:** HTTP 400 Bad Request `additional_price must be >= 0`.

### CAT-022: Remoção de Opção de Grupo de Adicionais
- **Pré-condição:** Opção "Bacon Extra" (ID 5) existente.
- **Passos:** DELETE `/restaurants/pizzaria-bela/sides/options/5`.
- **Esperado:** HTTP 204 No Content. Opção removida.

### CAT-023: Criação de Item Classificado como Combo (`is_combo = true`)
- **Pré-condição:** Categoria "Combos" existente.
- **Passos:** POST `/restaurants/pizzaria-bela/items` com `{"name": "Combo Família", "base_price": 89.90, "is_combo": true}`.
- **Esperado:** HTTP 201 Created com flag `is_combo = true`.

### CAT-024: Composição de Receita do Combo (`combo_item`)
- **Pré-condição:** Combo ID 10 ("Combo Família"), Pizza ID 1 ("Margherita") e Refrigerante ID 2 ("Guaraná 2L").
- **Passos:** Associar sub-itens via POST `/items/10/combo-items`.
- **Entradas:**
  - Sub-item 1: `{"sub_item_id": 1, "quantity": 2}`
  - Sub-item 2: `{"sub_item_id": 2, "quantity": 1}`
- **Esperado:** HTTP 201 Created. Registros inseridos em `combo_item` definindo o blueprint do pacote.

### CAT-025: Prevenção de Auto-Referência Circular em Combos
- **Pré-condição:** Item ID 10 é um combo.
- **Passos:** Tentar adicionar o item ID 10 como sub-item dele mesmo (`sub_item_id = 10`).
- **Esperado:** HTTP 400 Bad Request `a combo cannot contain itself`.

### CAT-026: Prevenção de Quantidade Zero ou Negativa em `combo_item`
- **Pré-condição:** Combo existente.
- **Passos:** POST com `{"sub_item_id": 1, "quantity": 0}` e `{"quantity": -1}`.
- **Esperado:** HTTP 400 Bad Request `quantity must be >= 1`.

### CAT-027: Cadastro de Promoção Temporária de Item (`item_sale`) em Porcentagem
- **Pré-condição:** Item com preço base de 50.00.
- **Passos:** POST `/items/1/sales`.
- **Entradas:**
  ```json
  {
    "discount_type": "percent",
    "discount_value": 20.00,
    "starts_at": "2026-10-12T10:00:00Z",
    "ends_at": "2026-10-12T14:00:00Z"
  }
  ```
- **Esperado:** HTTP 201 Created. Desconto de 20% (R$ 10,00) registrado.

### CAT-028: Cadastro de Promoção com Valor Fixo de Desconto
- **Pré-condição:** Item com preço base de 50.00.
- **Passos:** POST `/items/1/sales` com `{"discount_type": "fixed", "discount_value": 15.00}`.
- **Esperado:** HTTP 201 Created. Preço promocional resultante será 35.00.

### CAT-029: Validação de Janela de Datas em `item_sale` (`starts_at < ends_at`)
- **Pré-condição:** Nenhuma.
- **Passos:** POST com data de fim anterior ao início: `{"starts_at": "2026-10-12T15:00:00Z", "ends_at": "2026-10-12T10:00:00Z"}`.
- **Esperado:** HTTP 400 Bad Request `ends_at must be after starts_at`.

### CAT-030: Validação de Desconto Superior a 100% ou Superior ao Preço Base
- **Pré-condição:** Item de 50.00.
- **Passos:** Submeter desconto percentual de 105% ou desconto fixo de 60.00.
- **Entradas:**
  - Percentual: `{"discount_type": "percent", "discount_value": 105.00}` -> HTTP 400.
  - Fixo: `{"discount_type": "fixed", "discount_value": 60.00}` -> HTTP 400.
- **Esperado:** Preço resultante nunca pode ser negativo.

### CAT-031: Obtenção de Detalhes Completos do Item (`MenuItemDetail`)
- **Pré-condição:** Item com ingredientes, grupos de opcionais e promoções cadastradas.
- **Passos:** GET `/restaurants/pizzaria-bela/items/1`.
- **Esperado:** HTTP 200 OK com árvore JSON completa contendo dados base, array de `ingredients` e array de `sides` com suas respectivas `options`.

### CAT-032: Exclusão de Item do Catálogo sem Histórico de Pedidos
- **Pré-condição:** Item recém-criado que nunca foi comprado.
- **Passos:** DELETE `/restaurants/pizzaria-bela/items/99`.
- **Esperado:** HTTP 204 No Content. Item removido de `menu_item`.

### CAT-033: Bloqueio de Exclusão Física de Item com Pedidos Anteriores (Soft Delete)
- **Pré-condição:** Item ID 1 já foi comprado em pedidos históricos.
- **Passos:** DELETE `/restaurants/pizzaria-bela/items/1`.
- **Esperado:** O sistema não executa `DELETE` físico para não quebrar a chave estrangeira em `order_item`. O item tem sua flag `is_available` definida como `false` ou é arquivado logicamente.

### CAT-034: Busca de Itens por Filtro de Categoria
- **Pré-condição:** Categoria 1 tem 5 itens; Categoria 2 tem 3 itens.
- **Passos:** GET `/restaurants/pizzaria-bela/items?category_id=1`.
- **Esperado:** HTTP 200 OK retornando exatamente os 5 itens da Categoria 1.

### CAT-035: Comportamento de Item Disponível sem Foto (`photo_url` Nula)
- **Pré-condição:** Nenhuma.
- **Passos:** POST com `{"photo_url": null}`.
- **Esperado:** HTTP 201 Created. Frontend exibe imagem placeholder padrão.

### CAT-036: Modificação de Preço Base de Item Não Afeta Pedidos Já Realizados
- **Pré-condição:** Pedido anterior continha Item 1 comprado por 30.00.
- **Passos:** Atualizar preço do Item 1 para 45.00 via PUT `/items/1`. Consultar pedido anterior.
- **Esperado:** Pedido anterior permanece exibindo 30.00 (`unit_price_frozen`).

### CAT-037: Reordenação em Lote de Categorias (Batch Reorder)
- **Pré-condição:** Categorias 1, 2, 3 existentes.
- **Passos:** PATCH `/categories/reorder` com payload `[{"id": 3, "order": 1}, {"id": 1, "order": 2}, {"id": 2, "order": 3}]`.
- **Esperado:** HTTP 200 OK com todas as ordens atualizadas em transação atômica.

### CAT-038: Cadastro de Item com Descrição Longa (Valor Limite de Texto)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter item com descrição de 1.000 caracteres.
- **Esperado:** HTTP 201 Created. Campo `TEXT` armazena sem truncamento.

### CAT-039: Concorrência na Edição do Mesmo Produto por Dois Gerentes
- **Pré-condição:** Produto com preço 20.00.
- **Passos:** Gerente A envia preço 25.00 e Gerente B envia preço 30.00 simultaneamente.
- **Esperado:** O último commit prevalece de forma consistente, sem corrupção do registro.

### CAT-040: Ativação Simultânea de Múltiplas Promoções para o Mesmo Item
- **Pré-condição:** Item 1 com promoção A (10% de desconto) ativa.
- **Passos:** Tentar cadastrar promoção B (15% de desconto) para o mesmo período de datas.
- **Esperado:** HTTP 409 Conflict. Apenas uma promoção pode vigorar para o mesmo item na mesma janela de tempo.

### CAT-041: Consulta de Cardápio com Cache (Headers ETag e 304 Not Modified)
- **Pré-condição:** Cardápio sem alterações nos últimos 10 minutos.
- **Passos:** GET `/restaurants/pizzaria-bela/menu` com header `If-None-Match: <hash>`.
- **Esperado:** HTTP 304 Not Modified economizando tráfego de rede.

### CAT-042: Cadastro de Opcional com Nome Extremamente Longo
- **Pré-condição:** Grupo de opcionais existente.
- **Passos:** POST opção com nome de 300 caracteres.
- **Esperado:** HTTP 400 Bad Request se o limite do campo for 100 caracteres.

### CAT-043: Isolamento de Catálogo: Tentativa de Vincular Categoria de Outro Restaurante
- **Pré-condição:** Categoria 50 pertence a `hamburgueria-x`.
- **Passos:** Tentar criar produto em `pizzaria-bela` apontando para `category_id = 50`.
- **Esperado:** HTTP 400 Bad Request `category does not belong to this restaurant`.

### CAT-044: Exclusão de Grupo de Opcionais com Opções Aninhadas
- **Pré-condição:** Grupo ID 3 possui 4 opções cadastradas em `side_option`.
- **Passos:** DELETE `/items/1/sides/3`.
- **Esperado:** HTTP 204 No Content. Todas as 4 opções filhas são deletadas em cascata sem erro de chave estrangeira.

### CAT-045: Remoção de Ingrediente Base Referenciado em Fichas Técnicas
- **Pré-condição:** Ingrediente "Tomate" utilizado em 15 pizzas diferentes.
- **Passos:** DELETE `/ingredients/1`.
- **Esperado:** HTTP 409 Conflict `ingredient is in use by existing products`.

---

## Módulo 4: Pedidos, Especialização (Salão vs. Delivery) e Imutabilidade (ORD-001 a ORD-040)

### ORD-001: Criação de Pedido de Salão Completo (`local_order`)
- **Pré-condição:** Filial 1 ativa; Mesa 5 disponível; Item 1 (preço 25.00).
- **Passos:** POST `/restaurants/pizzaria-bela/branches/1/orders/local`.
- **Entradas:**
  ```json
  {
    "user_cpf": "52998224725",
    "table_number": 5,
    "items": [
      {"item_id": 1, "quantity": 2, "notes": "Sem cebola"}
    ]
  }
  ```
- **Esperado:** HTTP 201 Created. Registro inserido em `order` (`status = 'pending'`, `order_type = 'local'`, `subtotal = 50.00`, `total = 50.00`), registro em `local_order` (`table_number = 5`), e 2 itens em `order_item` com `unit_price_frozen = 25.00`. Nenhuma linha em `remote_order`.

### ORD-002: Criação de Pedido para Entrega Completo (`remote_order`)
- **Pré-condição:** Filial 1; Endereço ID 20 cadastrado; Taxa de frete 8.50; Item 1 (40.00).
- **Passos:** POST `/restaurants/pizzaria-bela/branches/1/orders/remote`.
- **Entradas:**
  ```json
  {
    "user_cpf": "52998224725",
    "delivery_address": {"cep": "51020010", "street": "Rua A", "number": "100"},
    "phone": "81999990000",
    "freight_fee": 8.50,
    "items": [{"item_id": 1, "quantity": 1}]
  }
  ```
- **Esperado:** HTTP 201 Created. `order.total = 48.50` (`40.00 + 8.50`). Registro em `remote_order`. Nenhuma linha em `local_order`.

### ORD-003: Subtipo Exclusivo: Bloqueio de Pedido Híbrido (Local + Remoto)
- **Pré-condição:** Pedido com `order_type = 'local'`.
- **Passos:** Tentar inserir registro em `remote_order` para o mesmo `order_id`.
- **Esperado:** Violação de integridade ou HTTP 422 Unprocessable Entity. Um pedido não pode ser simultaneamente local e remoto.

### ORD-004: Pedido Anônimo (Sem CPF do Consumidor)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido local com `user_cpf = null`.
- **Esperado:** HTTP 201 Created. Pedido processado normalmente. Nenhum ponto de fidelidade creditado ao cliente.

### ORD-005: Congelamento Obrigatório de Preço (`unit_price_frozen`)
- **Pré-condição:** Item 1 cadastrado no catálogo a 35.00.
- **Passos:** Criar pedido contendo o Item 1. Consultar `order_item.unit_price_frozen`.
- **Esperado:** Valor `35.00` gravado diretamente na linha do item. Alterações futuras no catálogo não afetam o pedido.

### ORD-006: Aplicação de Desconto Vigente de Promoção no `unit_price_frozen`
- **Pré-condição:** Item 1 com preço base de 50.00 e promoção ativa de 20% (preço promocional 40.00).
- **Passos:** Criar pedido dentro do horário da promoção.
- **Esperado:** `unit_price_frozen` gravado como `40.00`, e não 50.00.

### ORD-007: Compra Fora da Janela da Promoção Não Aplica Desconto
- **Pré-condição:** Promoção das 11:00 às 14:00.
- **Passos:** Criar pedido às 14:05.
- **Esperado:** `unit_price_frozen` gravado com o preço integral de 50.00.

### ORD-008: Validação de Quantidade Mínima de Item no Pedido (`quantity >= 1`)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido com `{"quantity": 0}` ou `{"quantity": -2}`.
- **Esperado:** HTTP 400 Bad Request `quantity must be greater than zero`.

### ORD-009: Pedido com Múltiplos Itens Diferentes
- **Pré-condição:** Item A (10.00), Item B (15.00), Item C (20.00).
- **Passos:** Criar pedido com 2x Item A, 1x Item B e 3x Item C.
- **Esperado:** Subtotal calculado com exatidão: `(2*10) + (1*15) + (3*20) = 20 + 15 + 60 = 95.00`.

### ORD-010: Pedido com Produto Indisponível (`is_available = false`)
- **Pré-condição:** Item com `is_available = false`.
- **Passos:** Tentar submeter pedido contendo este item.
- **Esperado:** HTTP 400 Bad Request `item is currently unavailable`.

### ORD-011: Validação de Itens Obrigatórios de Adicionais (`min_options` Violado)
- **Pré-condição:** Grupo "Ponto da Carne" exige `min_options = 1`.
- **Passos:** Fazer pedido do hambúrguer sem selecionar nenhum ponto.
- **Esperado:** HTTP 400 Bad Request `missing required options for group: Ponto da Carne`.

### ORD-012: Validação de Limite Máximo de Adicionais (`max_options` Violado)
- **Pré-condição:** Grupo de opcionais permite no máximo 2 molhos (`max_options = 2`).
- **Passos:** Fazer pedido selecionando 3 molhos.
- **Esperado:** HTTP 400 Bad Request `exceeded maximum options allowed (2)`.

### ORD-013: Adição do Preço do Opcional ao Valor do Item do Pedido
- **Pré-condição:** Hambúrguer base (25.00) + Bacon Extra (5.00).
- **Passos:** Criar pedido com essa seleção.
- **Esperado:** O valor computado para a linha do item reflete `30.00`.

### ORD-014: Decomposição de Combo em `order_item` (Itens Pais e Filhos)
- **Pré-condição:** "Combo Família" é composto por 1 Pizza e 1 Refrigerante.
- **Passos:** Criar pedido com o combo.
- **Esperado:**
  - Linha pai: Item do Combo com valor total.
  - Linhas filhas: Pizza e Refrigerante com `parent_order_item_id` apontando para o ID da linha pai e `unit_price_frozen = 0.00`.

### ORD-015: Adição de Novo Item a Comanda Aberta (Mesa Ativa)
- **Pré-condição:** Pedido de salão em status `pending` na Mesa 5.
- **Passos:** POST `/orders/101/items` com `{"item_id": 2, "quantity": 1}`.
- **Esperado:** HTTP 201 Created. Item inserido e `order.subtotal` e `order.total` recalculados e atualizados no banco.

### ORD-016: Bloqueio de Adição de Item a Pedido Já Concluído ou Cancelado
- **Pré-condição:** Pedido em status `delivered` ou `cancelled`.
- **Passos:** POST `/orders/101/items`.
- **Esperado:** HTTP 409 Conflict `cannot modify a completed or cancelled order`.

### ORD-017: Cancelamento de Item Específico de Comanda Aberta
- **Pré-condição:** Comanda com 3 itens (Total 75.00).
- **Passos:** DELETE `/orders/101/items/5` (item de 25.00).
- **Esperado:** HTTP 204 No Content. Linha removida e total do pedido reajustado para 50.00.

### ORD-018: Transição de Status: `pending` -> `preparing`
- **Pré-condição:** Pedido em `pending`.
- **Passos:** PATCH `/orders/101/status` com `{"status": "preparing"}`.
- **Esperado:** HTTP 200 OK. Status alterado. Notificação enviada para a cozinha.

### ORD-019: Transição de Status: `preparing` -> `ready`
- **Pré-condição:** Pedido em `preparing`.
- **Passos:** PATCH com `{"status": "ready"}`.
- **Esperado:** HTTP 200 OK. Comanda pronta para entrega ao garçom ou motoboy.

### ORD-020: Transição de Status: `ready` -> `delivered`
- **Pré-condição:** Pedido em `ready`.
- **Passos:** PATCH com `{"status": "delivered"}`.
- **Esperado:** HTTP 200 OK. Pedido finalizado.

### ORD-021: Rejeição de Transição Inválida de Status (Pulo Proibido)
- **Pré-condição:** Pedido em `pending`.
- **Passos:** Tentar mover direto para `delivered` sem passar por preparo.
- **Esperado:** HTTP 400 Bad Request `invalid status transition from pending to delivered`.

### ORD-022: Cancelamento de Pedido com Motivo (`cancelled`)
- **Pré-condição:** Pedido em `pending`.
- **Passos:** PATCH com `{"status": "cancelled", "reason": "Cliente desistiu"}`.
- **Esperado:** HTTP 200 OK. Pedido cancelado.

### ORD-023: Bloqueio de Cancelamento de Pedido Já Entregue
- **Pré-condição:** Pedido em status `delivered`.
- **Passos:** Tentar mudar para `cancelled`.
- **Esperado:** HTTP 400 Bad Request `cannot cancel an order that is already delivered`.

### ORD-024: Validação de Número de Mesa Válido em Pedido Local
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido local com `table_number = 0` ou `table_number = -5`.
- **Esperado:** HTTP 400 Bad Request `table_number must be a positive integer`.

### ORD-025: Múltiplas Comandas na Mesma Mesa (Mesas Compartilhadas)
- **Pré-condição:** Mesa 4 já possui um pedido aberto.
- **Passos:** Abrir segundo pedido para outro cliente na Mesa 4.
- **Esperado:** HTTP 201 Created com IDs de pedido distintos, identificados pelo cliente.

### ORD-026: Cálculo de Frete com Valor Negativo Rejeitado
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido de delivery com `{"freight_fee": -10.00}`.
- **Esperado:** HTTP 400 Bad Request `freight_fee must be >= 0`.

### ORD-027: Validação de Campo Observações (`notes` no item)
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar observação de 500 caracteres (ex: detalhes de alergia alimentar).
- **Esperado:** HTTP 201 Created com observação gravada e enviada para o ticket da cozinha.

### ORD-028: Inclusão de Cupom de Desconto Válido no Pedido
- **Pré-condição:** Cupom `DESC10` de 10% válido; Pedido com subtotal 100.00.
- **Passos:** Aplicar cupom no pedido.
- **Esperado:** `discount = 10.00`, `total = 90.00`.

### ORD-029: Rejeição de Cupom com Pedido Abaixo do Valor Mínimo
- **Pré-condição:** Cupom exige pedido mínimo de 50.00; Pedido com subtotal 30.00.
- **Passos:** Tentar aplicar o cupom.
- **Esperado:** HTTP 400 Bad Request `minimum order value for this coupon is 50.00`.

### ORD-030: Rejeição de Cupom Expirado
- **Pré-condição:** Cupom com `expires_at` anterior à data atual.
- **Passos:** Aplicar cupom no checkout.
- **Esperado:** HTTP 400 Bad Request `coupon has expired`.

### ORD-031: Rejeição de Cupom Esgotado (`max_uses` Atingido)
- **Pré-condição:** Cupom configurado para 5 usos máximos, já utilizado 5 vezes.
- **Passos:** Sexto usuário tenta aplicar o cupom.
- **Esperado:** HTTP 400 Bad Request `coupon usage limit reached`.

### ORD-032: Consulta de Pedido com Todos os Dados Detalhados (`OrderDetail`)
- **Pré-condição:** Pedido com itens, adicionais, dados locais e pagamentos.
- **Passos:** GET `/orders/101`.
- **Esperado:** HTTP 200 OK com payload estruturado completo contendo itens e histórico.

### ORD-033: Consulta de Pedido por ID Inexistente
- **Pré-condição:** ID 999999 inexistente.
- **Passos:** GET `/orders/999999`.
- **Esperado:** HTTP 404 Not Found `order not found`.

### ORD-034: Listagem de Pedidos Filtrados por Status
- **Pré-condição:** 10 pedidos na filial (4 `pending`, 3 `preparing`, 3 `ready`).
- **Passos:** GET `/branches/1/orders?status=preparing`.
- **Esperado:** Retorno com exatamente os 3 pedidos em preparo.

### ORD-035: Concorrência: Dois Garçons Adicionando Itens na Mesma Comanda Simultaneamente
- **Pré-condição:** Comanda aberta na Mesa 2.
- **Passos:** Dois garçons submetem novos itens no mesmo milissegundo.
- **Esperado:** Ambos os itens inseridos corretamente e o total recalculado sem perda de registros (transação atômica serializável).

### ORD-036: Verificação de Pedido Vazio (Sem Itens)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/orders/local` com array `"items": []`.
- **Esperado:** HTTP 400 Bad Request `an order must contain at least one item`.

### ORD-037: Validação de Telefone de Contato no Pedido Remoto
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter pedido de delivery sem telefone de contato.
- **Esperado:** HTTP 400 Bad Request `contact phone is required for delivery orders`.

### ORD-038: Bloqueio de Pedido para Filial Fechada / Fora de Horário
- **Pré-condição:** Filial física sem turno operacional aberto.
- **Passos:** Tentar abrir pedido na filial.
- **Esperado:** HTTP 409 Conflict `branch is currently closed`.

### ORD-039: Idempotência na Criação de Pedido com Header `Idempotency-Key`
- **Pré-condição:** Cliente com oscilação de rede envia a mesma requisição duas vezes com a mesma chave.
- **Passos:** Disparar dois POST idênticos com `Idempotency-Key: abc-123`.
- **Esperado:** Apenas um pedido é criado no banco; a segunda requisição devolve a resposta cacheada do primeiro.

### ORD-040: Deleção de Pedido em Cascata Rejeitada para Pedidos Pagos
- **Pré-condição:** Pedido com pagamento aprovado.
- **Passos:** Tentar executar DELETE direto no banco sem estorno.
- **Esperado:** Bloqueado por integridade referencial com a tabela `transaction`.

---

## Módulo 5: Pagamentos, Transações, PIX e Fidelidade (FIN-001 a FIN-030)

### FIN-001: Pagamento Integral de Pedido em Dinheiro (`cash`)
- **Pré-condição:** Pedido de 50.00 com status `pending`.
- **Passos:** POST `/orders/101/transactions`.
- **Entradas:** `{"payment_method": "cash", "amount": 50.00}`.
- **Esperado:** HTTP 201 Created. Registro criado em `transaction` com `status = 'paid'`.

### FIN-002: Pagamento Parcial / Dividido da Comanda (Split Payment)
- **Pré-condição:** Comanda de 100.00 na Mesa 10 dividida entre 2 clientes.
- **Passos:** Submeter duas transações de 50.00.
- **Entradas:**
  - Cliente 1: `{"payment_method": "credit_card", "amount": 50.00}`
  - Cliente 2: `{"payment_method": "pix", "amount": 50.00}`
- **Esperado:** Ambas transações aprovadas. Soma das transações pagas atinge o valor total do pedido.

### FIN-003: Bloqueio de Pagamento com Valor Superior ao Saldo Devedor
- **Pré-condição:** Pedido com saldo devedor restante de 20.00.
- **Passos:** Submeter transação de 50.00 via cartão de crédito.
- **Esperado:** HTTP 400 Bad Request `payment amount exceeds remaining order balance`.

### FIN-004: Geração de Cobrança Instantânea PIX (`transaction_pix`)
- **Pré-condição:** Transação ID 500 criada com `payment_method = 'pix'`.
- **Passos:** POST `/transactions/500/pix`.
- **Esperado:** HTTP 201 Created. Registro em `transaction_pix` contendo código copia-e-cola (`copy_paste_code`), imagem base64 do QR code (`qr_code_base64`) e data de expiração (`expires_at` em 15 minutos).

### FIN-005: Consulta de Status do Pagamento PIX
- **Pré-condição:** Cobrança PIX gerada.
- **Passos:** GET `/transactions/500/pix`.
- **Esperado:** Retorno do status atual (`pending` ou `paid`).

### FIN-006: Webhook / Confirmação de Pagamento PIX pelo Banco
- **Pré-condição:** Transação PIX em `pending`.
- **Passos:** Simular notificação do PSP bancário via webhook POST `/webhooks/pix`.
- **Esperado:** HTTP 200 OK. `transaction.status` atualizado para `'paid'`. Comanda liberada automaticamente.

### FIN-007: Expiração de Cobrança PIX após Tempo Limite
- **Pré-condição:** Cobrança gerada com expiração de 15 minutos.
- **Passos:** Consultar status após 16 minutos sem pagamento.
- **Esperado:** `transaction.status` atualizado para `'failed'` com motivo `expired`.

### FIN-008: Validação de Métodos de Pagamento Permitidos
- **Pré-condição:** Nenhuma.
- **Passos:** Tentar criar transação com método não suportado.
- **Entradas:** `{"payment_method": "cheque_sem_fundo", "amount": 50.00}`.
- **Esperado:** HTTP 400 Bad Request. Aceitos apenas `credit_card`, `debit_card`, `pix` e `cash`.

### FIN-009: Emissão de Recibo Digital Imutável (`receipt`)
- **Pré-condição:** Pedido 101 totalmente pago.
- **Passos:** GET `/orders/101/receipt`.
- **Esperado:** HTTP 200 OK. Snapshot estruturado em JSON com hash único `receipt_code`, lista de itens, impostos, método de pagamento e carimbo temporal.

### FIN-010: Imutabilidade do Recibo Emitido
- **Pré-condição:** Recibo gerado para o Pedido 101.
- **Passos:** Tentar alterar dados do recibo via API ou banco.
- **Esperado:** Ação rejeitada. Recibos são registros fiscais estritamente append-only.

### FIN-011: Cálculo da Taxa de Take-Rate da Plataforma Pratello
- **Pré-condição:** Restaurante com comissão de 5% acordada. Pedido total de 200.00.
- **Passos:** Concluir transação financeira do pedido.
- **Esperado:** Comissão de R$ 10,00 calculada e segregada para o faturamento da plataforma Pratello.

### FIN-012: Crédito de Pontos de Fidelidade / Cashback (`user_restaurant`)
- **Pré-condição:** Cliente identificado pelo CPF realizou pedido de 100.00. Regra de fidelidade: 1 ponto a cada R$ 10 gastos.
- **Passos:** Concluir pagamento do pedido.
- **Esperado:** Saldo em `user_restaurant.points_balance` incrementado em 10 pontos.

### FIN-013: Consulta de Saldo de Carteira de Fidelidade por Cliente
- **Pré-condição:** Cliente com 50 pontos acumulados na "Pizzaria Bela".
- **Passos:** GET `/restaurants/pizzaria-bela/users/52998224725/wallet`.
- **Esperado:** HTTP 200 OK com `points_balance = 50`.

### FIN-014: Isolamento de Carteira de Fidelidade Entre Restaurantes
- **Pré-condição:** Cliente tem 50 pontos na "Pizzaria Bela" e 0 pontos na "Hamburgueria Bela".
- **Passos:** Consultar carteira na "Hamburgueria Bela".
- **Esperado:** Saldo de 0 pontos. Pontos não são transferíveis entre tenants.

### FIN-015: Resgate / Débito de Pontos de Fidelidade como Desconto no Pedido
- **Pré-condição:** Cliente com 100 pontos (cada 10 pontos = R$ 1,00 de desconto). Pedido de 40.00.
- **Passos:** Resgatar 50 pontos no pedido.
- **Esperado:** Desconto de R$ 5,00 aplicado no pedido. Saldo da carteira atualizado para 50 pontos.

### FIN-016: Bloqueio de Resgate de Pontos Superior ao Saldo Disponível
- **Pré-condição:** Cliente possui 20 pontos.
- **Passos:** Tentar resgatar 50 pontos.
- **Esperado:** HTTP 400 Bad Request `insufficient loyalty points balance`.

### FIN-017: Prevenção de Saldo Negativo na Carteira de Pontos
- **Pré-condição:** Saldo de 0 pontos.
- **Passos:** Tentativa de débito manual.
- **Esperado:** Violação da restrição de integridade `CHECK (points_balance >= 0)`.

### FIN-018: Criação de Cupom de Desconto com Código Único por Restaurante
- **Pré-condição:** Cupom `BEMVINDO` disponível.
- **Passos:** POST `/restaurants/pizzaria-bela/coupons`.
- **Entradas:**
  ```json
  {
    "code": "BEMVINDO",
    "discount_type": "percent",
    "discount_value": 15.00,
    "min_order_value": 30.00,
    "expires_at": "2026-12-31T23:59:59Z",
    "max_uses": 100
  }
  ```
- **Esperado:** HTTP 201 Created. Cupom cadastrado na tabela `coupon`.

### FIN-019: Mesmo Código de Cupom Permitido em Restaurantes Diferentes
- **Pré-condição:** Cupom `PROMO10` existe em `pizzaria-bela`.
- **Passos:** Criar cupom com o mesmo código `PROMO10` em `hamburgueria-x`.
- **Esperado:** HTTP 201 Created. Permitido devido ao isolamento multi-tenant.

### FIN-020: Conflito de Cupom Duplicado no Mesmo Restaurante
- **Pré-condição:** Cupom `PROMO10` já existe em `pizzaria-bela`.
- **Passos:** Tentar recriar `PROMO10` para `pizzaria-bela`.
- **Esperado:** HTTP 409 Conflict `coupon code already exists for this restaurant`.

### FIN-021: Desativação Manual de Cupom Promocional
- **Pré-condição:** Cupom ativo.
- **Passos:** DELETE `/restaurants/pizzaria-bela/coupons/1`.
- **Esperado:** HTTP 204 No Content. Cupom não pode mais ser aplicado em novos pedidos.

### FIN-022: Estorno / Reembolso de Transação (`refunded`)
- **Pré-condição:** Pedido cancelado com transação de cartão paga.
- **Passos:** Executar estorno via POST `/transactions/500/refund`.
- **Esperado:** `transaction.status` atualizado para `'refunded'`.

### FIN-023: Pagamento de Pedido com Falha no Cartão (Transação Recusada)
- **Pré-condição:** Pedido aberto.
- **Passos:** Gateway recusa o cartão por saldo insuficiente.
- **Esperado:** `transaction.status` registrado como `'failed'`. Pedido continua aberto aguardando outro método de pagamento.

### FIN-024: Múltiplas Tentativas de Pagamento para a Mesma Comanda
- **Pré-condição:** Tentativa 1 com cartão falhou.
- **Passos:** Cliente efetua Tentativa 2 com PIX e é aprovado.
- **Esperado:** Tabela `transaction` mantém o histórico de ambas as tentativas (1 failed, 1 paid).

### FIN-025: Pagamento em Dinheiro com Cálculo de Troco
- **Pré-condição:** Pedido de 35.00.
- **Passos:** Cliente entrega nota de 50.00 em dinheiro.
- **Esperado:** Sistema registra transação de 35.00 e calcula troco de R$ 15,00 no comprovante.

### FIN-026: Idempotência no Processamento de Pagamento com Cartão
- **Pré-condição:** Requisição de cobrança enviada.
- **Passos:** Disparo duplo da mesma transação de cartão com a mesma chave de idempotência.
- **Esperado:** O cartão do cliente é debitado estritamente uma única vez.

### FIN-027: Estorno de Pontos de Fidelidade em Pedido Cancelado
- **Pré-condição:** Pedido gerou 10 pontos para o cliente, mas foi cancelado posteriormente.
- **Passos:** Executar cancelamento e estorno do pedido.
- **Esperado:** 10 pontos debitados da carteira do cliente, restaurando o saldo original.

### FIN-028: Relatório Financeiro: Fechamento de Caixa por Filial e Turno
- **Pré-condição:** Turno encerrado com 20 transações (R$ 500 Dinheiro, R$ 800 Cartão, R$ 700 PIX).
- **Passos:** GET `/reports/cashier-closure?session_id=12`.
- **Esperado:** Total consolidado de R$ 2.000 com discriminação precisa por método de pagamento.

### FIN-029: Auditoria de Transações Financeiras (Log Imutável)
- **Pré-condição:** Qualquer alteração de status financeiro.
- **Passos:** Validar persistência de logs de data, hora e operador responsável pela cobrança.
- **Esperado:** Rastreabilidade garantida para compliance e conciliação bancária.

### FIN-030: Tentativa de Pagar Pedido com Valor Zero (Sem Itens)
- **Pré-condição:** Pedido sem itens cadastrados.
- **Passos:** Tentar gerar pagamento.
- **Esperado:** HTTP 400 Bad Request `cannot process payment for zero or empty order`.

---

## Módulo 6: Turnos Operacionais, Concorrência e Resiliência (OPS-001 a OPS-020)

### OPS-001: Abertura de Turno Operacional de Trabalho (`restaurant_staff_session`)
- **Pré-condição:** Filial física ativa.
- **Passos:** POST `/restaurants/pizzaria-bela/branches/1/staff-sessions/open`.
- **Entradas:** `{"shift_code": "ALMOCO-2026-10-12"}`.
- **Esperado:** HTTP 201 Created. Registro criado com `opened_at = now()`, `closed_at = null` e `qr_token` criptograficamente seguro gerado para ponto da equipe.

### OPS-002: Bloqueio de Abertura de Turno Simultâneo na Mesma Filial
- **Pré-condição:** Turno de almoço já aberto e ativo na Filial 1.
- **Passos:** Tentar abrir outro turno na mesma filial antes de encerrar o anterior.
- **Esperado:** HTTP 409 Conflict `an active staff session is already running for this branch`.

### OPS-003: Check-in de Funcionário via Token QR do Turno (`staff_session`)
- **Pré-condição:** Turno aberto com `qr_token = 'token_xyz_123'`. Funcionário com conta `account_id = 8`.
- **Passos:** POST `/staff-sessions/check-in` com `{"qr_token": "token_xyz_123", "account_id": 8}`.
- **Esperado:** HTTP 201 Created. Registro inserido em `staff_session` com `checked_in_at = now()` e `checked_out_at = null`.

### OPS-004: Rejeição de Check-in com Token QR Inválido ou Forjado
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter POST `/staff-sessions/check-in` com token inexistente.
- **Esperado:** HTTP 401 Unauthorized `invalid or expired shift QR token`.

### OPS-005: Bloqueio de Duplo Check-in do Mesmo Funcionário sem Check-out
- **Pré-condição:** Funcionário ID 8 já efetuou check-in no turno ativo.
- **Passos:** Tentar efetuar check-in novamente.
- **Esperado:** HTTP 409 Conflict `employee is already clocked in`.

### OPS-006: Check-out de Funcionário do Turno
- **Pré-condição:** Ponto de entrada registrado.
- **Passos:** POST `/staff-sessions/55/check-out`.
- **Esperado:** HTTP 200 OK. `checked_out_at` preenchido com carimbo de data/hora atual.

### OPS-007: Fechamento do Turno Operacional da Filial
- **Pré-condição:** Turno aberto.
- **Passos:** POST `/restaurants/pizzaria-bela/branches/1/staff-sessions/12/close`.
- **Esperado:** HTTP 200 OK. `closed_at = now()`. Token QR invalidado para novos check-ins. Check-out automático ou advertência para funcionários que não registraram saída.

### OPS-008: Listagem de Funcionários Ativos no Turno Atual
- **Pré-condição:** 3 funcionários com ponto registrado e sem check-out.
- **Passos:** GET `/staff-sessions/active?branch_id=1`.
- **Esperado:** Retorno de lista com os 3 funcionários em plantão ativo.

### OPS-009: Teste de Concorrência: Venda do Último Item em Estoque
- **Pré-condição:** Item limitado a 1 unidade disponível.
- **Passos:** Dois clientes disparam pedido para a mesma unidade no exato mesmo milissegundo.
- **Esperado:** Uma requisição obtém sucesso (HTTP 201); a outra recebe erro de estoque esgotado (HTTP 409) sem estoque negativo (lock pessimista via `SELECT ... FOR UPDATE`).

### OPS-010: Concorrência: Atualização Simultânea de Saldo de Fidelidade
- **Pré-condição:** Cliente com saldo de 50 pontos.
- **Passos:** Duas requisições paralelas de débito de 30 pontos.
- **Esperado:** Apenas uma transação de débito é concluída com sucesso; a segunda é rejeitada por saldo insuficiente, prevenindo condição de corrida (Race Condition).

### OPS-011: Resiliência contra Queda Repentina de Conexão com o Banco de Dados
- **Pré-condição:** Aplicação em execução.
- **Passos:** Forçar interrupção do PostgreSQL durante a inicialização do roteador.
- **Esperado:** A aplicação não sofre pânico (crash incontrolado); retorna HTTP 503 Service Unavailable no healthcheck e tenta reconexão automática com backoff exponencial.

### OPS-012: Teste de Carga e Vazamento de Conexões (Connection Pool Exhaustion)
- **Pré-condição:** Pool configurado para 25 conexões máximas.
- **Passos:** Disparar 100 requisições simultâneas de leitura de cardápio.
- **Esperado:** Todas as requisições liberam suas conexões (`rows.Close()`) e concluem sem travar o pool de conexões do Go.

### OPS-013: Atomicidade de Pedido (Rollback em Caso de Erro no Meio da Transação)
- **Pré-condição:** Pedido com 3 itens, onde o 3º item possui um ID de produto inexistente.
- **Passos:** Submeter requisição de criação do pedido.
- **Esperado:** Falha no 3º item dispara `tx.Rollback()`. Nenhuma linha é gravada em `order`, `local_order` ou `order_item`. O banco permanece limpo.

### OPS-014: Validação de Healthcheck da Aplicação (`/health` e `/api/health`)
- **Pré-condição:** Banco de dados online.
- **Passos:** GET `/health`.
- **Esperado:** HTTP 200 OK com payload `{"status": "ok", "service": "pratello"}` em menos de 10ms.

### OPS-015: Validação de Healthcheck com Banco de Dados Fora do Ar
- **Pré-condição:** Banco de dados indisponível.
- **Passos:** GET `/health`.
- **Esperado:** HTTP 500 ou 503 com `{"status": "degraded", "database": "unreachable"}`.

### OPS-016: Proteção contra Payload JSON Gigante (Denial of Service por Memória)
- **Pré-condição:** Nenhuma.
- **Passos:** Enviar requisição POST com corpo de 50MB contendo caracteres em branco ou lixo.
- **Esperado:** HTTP 413 Payload Too Large (middleware limita tamanho do body a 2MB).

### OPS-017: Sanitização e Ocultação de Stack Traces em Produção
- **Pré-condição:** Erro interno proposital disparado (ex.: query SQL com tabela com permissão negada).
- **Passos:** Executar chamada que dispara o erro.
- **Esperado:** Retorno HTTP 500 contendo apenas mensagem amigável `{"error": "internal server error"}`. Nenhuma linha de código, path do servidor ou query SQL é vazada na resposta HTTP.

### OPS-018: Teste de Deadlock em Transações Cruzadas de Pedido e Pagamento
- **Pré-condição:** Duas threads tentando atualizar `order` e `transaction` em ordem inversa.
- **Passos:** Executar requisições concorrentes entrelaçadas.
- **Esperado:** O banco detecta deadlock e a camada de repositório do Go executa retry automático sem falhar para o usuário final.

### OPS-019: Idempotência de Inicialização de Schema (`InitSchema`)
- **Pré-condição:** Banco de dados já populado com todas as tabelas.
- **Passos:** Chamar `app.InitSchema()` múltiplas vezes na inicialização.
- **Esperado:** Nenhuma tabela é recriada ou truncada; comandos utilizam `CREATE TABLE IF NOT EXISTS` e executam sem erros.

### OPS-020: Encerramento Gracioso do Servidor (Graceful Shutdown)
- **Pré-condição:** Servidor processando requisição longa de relatório (duração de 3 segundos).
- **Passos:** Enviar sinal de término `SIGTERM` / `SIGINT` para o processo da API Go.
- **Esperado:** O servidor para de aceitar novas requisições na porta, aguarda a finalização da requisição ativa e encerra o processo com exit code 0 sem corromper transações.