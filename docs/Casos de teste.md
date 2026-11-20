# Plano e Especificação Detalhada de Testes — Pratello (DBML Alinhado)

> **Versão:** 0.1.4  
> **Classificação:** Suíte de Testes Profunda (Deep Testing Specification)  
> **Total de Casos de Teste:** 241 Casos

## Resumo Executivo e Metodologia de Teste

A suíte está estruturada em **7 Módulos Coesos de Alta Profundidade (Deep Modules)** cobrindo Caminho Feliz (*Happy Path*), Condições de Limite (*Boundary Testing*), Casos Negativos de Erro, Integridade Referencial (*ACID/FK*), Idempotência, Concorrência e Usabilidade Frontend:

1. **Módulo 1: Identidade, Contas, Papéis e Autoridade (IDP)** — Casos `IDP-001` a `IDP-026` (26 casos): cadastro, contas sombra, CPF, métodos de acesso, sessões e autorização.
2. **Módulo 2: Estabelecimentos, Filiais, RH e Operação de Terminal (TEN)** — Casos `TEN-001` a `TEN-030` (30 testes): Restaurante lógico (`restaurant` com slugs únicos), filiais físicas (`physical_restaurant`) com verificação cruzada obrigatória de CNPJ (filial + matriz), quadro de colaboradores de RH (`employee` sem contas auto-geradas e fluxo de convite) e sessões de terminal operacional de chão via PIN compartilhado de 6 dígitos (8h de turno, rate-limiting/lockout de 5 tentativas, baixo privilégio e isolamento de blast radius).
3. **Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT)** — Casos `CAT-001` a `CAT-045` (45 testes): Layout visual (`menu` com `CLASSICO` / `EXPERIENCIA`), categorias, itens (`menu_item`), ingredientes (`ingredient`), ficha técnica com medidas (`item_ingredient`), adicionais diretos (`item_side`), combos vinculados (`combo_item`) e promoções (`item_sale`).
4. **Módulo 4: Pedidos Anônimos, Especialização e Itens (ORD)** — Casos `ORD-001` a `ORD-040` (40 testes): Pedidos desacoplados de CPF com `customer_name` opcional, avaliações (`order.review` e `order_item.review`), observações/ingredientes removidos (`order_item.notes`), especialização estrita (`local_order` com `"table"` vs `remote_order` com `"delivery_fee"`), congelamento de valor (`unit_price`), combos com `parent_order_id` e máquina de estados (`PENDING` -> `COMPLETED`).
5. **Módulo 5: Pagamentos, PIX, Cupons e Fidelidade (FIN)** — Casos `FIN-001` a `FIN-030` (30 testes): Tipos de transação (`PIX`, `CREDIT`, `DEBIT`, `CASH`), geração de PIX com `pix_copy_paste`, recibos em JSONB (`receipt`), cupons com restrição composta `(restaurant_id, code)` e **ciclo de moedas (acúmulo anônimo por CPF vs exigência de `account` para resgate)**.
6. **Módulo 6: Turnos Operacionais, Concorrência e Resiliência (OPS)** — Casos `OPS-001` a `OPS-020` (20 testes): Sessões de turno (`staff_session` com `shift_code` de 6 dígitos e `qr_token`), concorrência em pedidos e estoque, exaustão de pool de conexões e graceful shutdown.
7. **Módulo 7: Interface do Usuário, Navegação e Interações do Website (WEB)** — Casos `WEB-001` a `WEB-050` (50 testes): Carrinho flutuante e gaveta de pedidos (`OrderCart` / `OrderSection`), abertura/fechamento de gaveta, incremento/decremento e remoção de itens, navegação e links no navbar, seleção de endereços (`AddressSection`), fluxo de autenticação e inputs de login (`Auth`), filtros do cardápio, modal de itens com adicionais (`ItemModal`) e finalização de checkout.

---

## Módulo 1: Identidade, Contas, Papéis e Autoridade (IDP-001 a IDP-026)

Os casos de identidade ficam em `server/internal/auth/auth_casos_de_testes_test.go`; os casos de autorização de restaurantes ficam em `server/internal/restaurants/restaurants_casos_de_testes_test.go`. Execute `go test -race ./server/internal/auth ./server/internal/restaurants`; os testes PostgreSQL usam `AUTH_TEST_DATABASE_URL` e um esquema temporário. A notificação e a verificação de e-mail/WhatsApp são **simuladas**: não há prova real de posse. `XFAIL` executa o cenário e registra `SKIP: XFAIL` somente quando a falha conhecida se manifesta; outros resultados inesperados falham normalmente.

### IDP-001: Cadastro, tokens e hash de senha
- **Pré-condição:** E-mail livre, notificador simulado.
- **Passos:** Registrar e examinar a resposta e a conta no armazenamento.
- **Entradas:** `user@example.com`, senha `password123`.
- **Esperado:** HTTP 201, par de tokens Bearer e senha bcrypt verificável, sem senha na resposta.
- **Teste:** `TestIDP001RegistrationHashAndTokens`.

### IDP-002: E-mail duplicado sem distinção de caixa
- **Pré-condição:** Conta `dup@pratello.com` existente.
- **Passos:** Registrar o mesmo e-mail com letras maiúsculas.
- **Entradas:** `DUP@PRATELLO.COM`, `validPassword123`.
- **Esperado:** HTTP 409 e nenhuma segunda conta.
- **Teste:** `TestIDP002DuplicateEmailCaseInsensitive`.

### IDP-003: JSON e credenciais de cadastro inválidos
- **Pré-condição:** Nenhuma conta criada.
- **Passos:** Enviar JSON truncado, e-mail inválido e senha de sete caracteres.
- **Entradas:** `{"email":`, `invalid-email`, `1234567`.
- **Esperado:** Cada requisição retorna HTTP 400 e não cria conta.
- **Teste:** `TestIDP003InvalidRegistration`.

### IDP-004: Login com senha e exclusão de e-mail não verificado
- **Pré-condição:** Conta com senha bcrypt existente.
- **Passos:** Autenticar com senha correta; marcar e-mail não verificado e repetir.
- **Entradas:** `a@example.com`, `password123`.
- **Esperado:** Login correto funciona; e-mail não verificado é rejeitado.
- **Testes:** `TestService_Register_And_Login_Success`, `TestService_LoginRejectsUnverifiedEmail`.

### IDP-005: Login sem enumeração de conta
- **Pré-condição:** Uma conta conhecida.
- **Passos:** Comparar login de conta existente com senha errada e conta inexistente.
- **Entradas:** Senha `wrong`, e-mail `missing@example.com`.
- **Esperado:** Ambos retornam HTTP 401 com a mesma mensagem.
- **Teste:** `TestIDP005InvalidLoginDoesNotEnumerate`.

### IDP-006: Graduação de conta sombra
- **Pré-condição:** Conta sombra com CPF e sem métodos de acesso.
- **Passos:** Registrar e-mail com o CPF da sombra; tentar novo cadastro com o mesmo CPF.
- **Entradas:** `12345678901`, `a@example.com`, `password123`.
- **Esperado:** Mesmo ID da sombra é preservado, CPF é marcado como reivindicado e o segundo cadastro conflita.
- **Teste:** `TestShadowGraduationAndClaim`.

### IDP-007: Reivindicação e desvinculação preservam carteira
- **Pré-condição:** PostgreSQL com conta real e sombra com carteira de 50 pontos.
- **Passos:** Reivindicar o CPF da sombra e depois removê-lo.
- **Entradas:** CPF `12345678901`.
- **Esperado:** Carteira passa à conta real sem perda de pontos e continua vinculada após remover CPF.
- **Teste:** `TestPostgresCPFClaimPreservesWallet` (requer `AUTH_TEST_DATABASE_URL`).

### IDP-008: Corrida por CPF único
- **Pré-condição:** CPF livre em armazenamento de memória.
- **Passos:** Executar cinco cadastros concorrentes para o mesmo CPF.
- **Entradas:** CPF `12345678901` e cinco telefones distintos.
- **Esperado:** Exatamente um cadastro vence; os demais retornam conflito.
- **Teste:** `TestConcurrentCPFRegistration`.

### IDP-009: CPF malformado
- **Pré-condição:** Conta autenticada.
- **Passos:** Tentar reivindicar CPF de tamanho ou caracteres inválidos.
- **Entradas:** `123`, `1234567890x`.
- **Esperado:** HTTP 400 e CPF da conta permanece ausente.
- **Teste:** `TestIDP009InvalidCPF`.

### IDP-010: Telefone não verificado não autentica
- **Pré-condição:** Conta cadastrada com telefone de contato não verificado.
- **Passos:** Tentar login, adicionar o telefone via confirmação simulada e tentar novamente.
- **Entradas:** `+5511999999999`.
- **Esperado:** Primeiro login rejeitado; após confirmação simulada, login gera tokens; após remoção, é rejeitado.
- **Teste:** `TestPhoneLoginRequiresVerifiedPhone`.

### IDP-011: Cadastro por telefone simulado
- **Pré-condição:** Telefone livre; verificação simulada.
- **Passos:** Concluir cadastro e autenticar pelo telefone sem inserir PIN.
- **Entradas:** `+5511888888888`.
- **Esperado:** HTTP 201 e par de tokens; login posterior retorna HTTP 200.
- **Teste:** `TestPhoneRegistrationHTTPWithoutPIN`.

### IDP-012: Telefone duplicado e último método
- **Pré-condição:** Duas contas; uma possui telefone verificado.
- **Passos:** Outra conta reivindica o telefone; em seguida, tentar removê-lo quando for o único método verificado.
- **Entradas:** `+5511999999999`.
- **Esperado:** Ambas operações retornam conflito e os métodos permanecem intactos.
- **Teste:** `TestPhoneConflictsAndLastMethod`.

### IDP-013: E-mail como último método
- **Pré-condição:** Conta com e-mail verificado e sem telefone verificado.
- **Passos:** Remover o e-mail; adicionar telefone e remover e-mail; tentar remover telefone.
- **Entradas:** `a@example.com`, `+5511999999999`.
- **Esperado:** Primeira e última remoções conflitam; remoção intermediária funciona.
- **Teste:** `TestEmailChangeAndLastMethod`.

### IDP-014: Alteração de e-mail exige senha atual
- **Pré-condição:** Conta com e-mail e senha.
- **Passos:** Alterar e-mail com senha atual incorreta e depois correta; autenticar com novo par.
- **Entradas:** `b@example.com`, `newpassword`, `wrong` e `password123`.
- **Esperado:** Tentativa incorreta rejeitada; alteração correta permite login com nova senha.
- **Teste:** `TestEmailChangeAndLastMethod`.

### IDP-015: Rotação de refresh e rejeição de replay
- **Pré-condição:** Conta real com par de tokens.
- **Passos:** Trocar refresh; reutilizar o antigo e usar o novo como bearer.
- **Entradas:** Refresh original e refresh rotacionado.
- **Esperado:** Troca funciona; replay e uso de refresh como access são rejeitados.
- **Teste:** `TestSessionRotationAndRevocation`.

### IDP-016: Logout global invalida sessões
- **Pré-condição:** Conta com par de tokens válido.
- **Passos:** Chamar `POST /auth/logout-all` e tentar usar tokens antigos.
- **Entradas:** Access e refresh emitidos antes do logout.
- **Esperado:** HTTP 204; access e refresh anteriores são rejeitados.
- **Testes:** `TestHTTPRegistrationAndLogoutAll`, `TestSessionRotationAndRevocation`.

### IDP-017: Bearer ausente, expirado ou forjado
- **Pré-condição:** Conta real e segredo de teste.
- **Passos:** Consultar `/auth/me` com cada token inválido.
- **Entradas:** Sem bearer, JWT expirado e JWT assinado com outro segredo.
- **Esperado:** HTTP 401 em todos os casos.
- **Teste:** `TestIDP017InvalidBearer`.

### IDP-018: Logout não revoga sessão alheia
- **Pré-condição:** Duas contas com IDs distintos e refresh da primeira.
- **Passos:** Segunda conta tenta revogar o refresh da primeira; primeira tenta renová-lo.
- **Entradas:** ID da segunda conta e refresh da primeira.
- **Esperado:** Revogação não autorizada; refresh original ainda funciona.
- **Teste:** `TestLogoutCannotRevokeAnotherAccount`.

### IDP-019: Proteção do último proprietário
- **Pré-condição:** Restaurante com apenas um proprietário.
- **Passos:** Solicitar remoção pela API; depois adicionar outro proprietário e repetir com o token dele.
- **Entradas:** IDs de duas contas existentes.
- **Esperado:** Primeira remoção rejeitada; segunda permitida sem deixar restaurante órfão.
- **Teste:** `TestRestaurantPostgresLifecycle` em `server/internal/restaurants/restaurants_casos_de_testes_test.go` (requer `AUTH_TEST_DATABASE_URL`).

### IDP-020: Promoção de sysadmin sem autenticação
- **Pré-condição:** Conta alvo existente; requisição anônima.
- **Passos:** Chamar `POST /sysadmins` sem bearer.
- **Entradas:** `{"account_id":1}`.
- **Esperado:** HTTP 401 e nenhuma promoção.
- **Teste:** `TestIDP020AnonymousSysadminPromotion`.

### IDP-021: Limite de tamanho de cadastro (XFAIL)
- **Pré-condição:** Limite de corpo esperado de 1 MiB.
- **Passos:** Enviar nome maior que 1 MiB ao registrar.
- **Entradas:** JSON com `name` de 1 MiB, e-mail e senha válidos.
- **Esperado:** HTTP 413 e nenhuma conta criada. **Falha conhecida:** o endpoint aceita o corpo.
- **Teste:** `TestIDP021OversizedRegistration`.

### IDP-022: Falha do notificador não cria conta
- **Pré-condição:** Notificador simulado que retorna erro.
- **Passos:** Solicitar cadastro por e-mail.
- **Entradas:** `alice@example.com`, `password123`.
- **Esperado:** Serviço retorna erro e não grava conta.
- **Teste:** `TestService_Register_NotificationFailureDoesNotCreateAccount`.

### IDP-023: Gerente não promove proprietário
- **Pré-condição:** `cafe` tem um proprietário; outra conta é gerente de uma filial sem ser proprietária, e há uma conta alvo.
- **Passos:** Enviar `POST /restaurants/cafe/owners` com o bearer do gerente, antes de revogar seu vínculo de gerente.
- **Entradas:** `account_id` da segunda conta.
- **Esperado:** HTTP 403 e somente o proprietário original permanece.
- **Teste:** `TestRestaurantPostgresLifecycle` em `server/internal/restaurants/restaurants_casos_de_testes_test.go` (requer `AUTH_TEST_DATABASE_URL`).

### IDP-024: CPF de outra conta real não pode ser reivindicado
- **Pré-condição:** Conta A possui CPF; conta B não possui CPF.
- **Passos:** Conta B tenta reivindicar o CPF de A.
- **Entradas:** CPF `12345678901`.
- **Esperado:** Conflito, CPF permanece em A e B continua sem CPF.
- **Teste:** `TestIDP024ClaimAnotherAccountsCPF`.

### IDP-025: Conta sombra não autentica com JWT
- **Pré-condição:** Conta com apenas CPF e um JWT de acesso corretamente assinado para seu ID.
- **Passos:** Enviar esse bearer em `GET /auth/me`.
- **Entradas:** CPF `12345678901`; JWT assinado pelo servidor.
- **Esperado:** HTTP 401; uma assinatura válida não transforma uma sombra em conta real.
- **Teste:** `TestIDP025ShadowCannotAuthenticate`.

### IDP-026: Remoção concorrente dos métodos de acesso
- **Pré-condição:** Conta com e-mail e telefone verificados em armazenamento de memória.
- **Passos:** Remover e-mail e telefone simultaneamente.
- **Entradas:** `race@example.com`, `+5511999999999`.
- **Esperado:** Exatamente uma remoção funciona; a outra conflita e um método verificado permanece.
- **Teste:** `TestIDP026ConcurrentLastMethodRemoval`.

---

## Módulo 2: Estabelecimentos e Filiais (TEN-001 a TEN-013)

Os testes estão em `server/internal/restaurants/restaurants_casos_de_testes_test.go`. Com `AUTH_TEST_DATABASE_URL` apontando para um PostgreSQL que permita criar esquemas, execute `go test -count=1 -race ./server/internal/restaurants -run '^TestTEN'`. Cada caso cria e remove seu próprio esquema. Esta revisão substitui os casos do antigo terminal por operações da API atual.

### TEN-001: Criação autenticada e vínculo inicial de proprietário
- **Pré-condição:** Duas contas válidas, sem organizações.
- **Passos:** Criar a organização sem autenticação e depois com o token do criador. Consultar `/me/restaurants` com ambas as contas e conferir o vínculo no banco.
- **Entradas:** `{"name":"Café","slug":"cafe"}`.
- **Esperado:** HTTP 401 e 201, respectivamente; somente a conta criadora recebe o papel `OWNER` e vê `cafe` em seus vínculos.
- **Teste:** `TestTEN001OrganizationCreation`.

### TEN-002: Formato, unicidade e imutabilidade do slug
- **Pré-condição:** Duas contas autenticadas e slug disponível.
- **Passos:** Enviar slug inválido; cadastrar sem slug; tentar duplicá-lo com outra conta e alterá-lo com `PUT`.
- **Entradas:** `Invalid slug`, nome `Trattoria & Cia`, slug duplicado `trattoria-cia` e novo slug `new-slug`.
- **Esperado:** HTTP 400, 201, 409 e 400, respectivamente. Apenas `trattoria-cia` é consultável.
- **Teste:** `TestTEN002SlugValidationAndUniqueness`.

### TEN-003: Paginação pública exclui organizações arquivadas
- **Pré-condição:** O proprietário cadastrou `gamma`, `alpha` e `beta` nessa ordem.
- **Passos:** Arquivar `alpha`; consultar três páginas de um item e solicitar uma página acima do limite de tamanho.
- **Entradas:** `page_size=1`, páginas 1 a 3; `page_size=101`.
- **Esperado:** `beta`, `gamma` e `[]`, com total 2 em todas as páginas válidas; HTTP 400 para tamanho 101.
- **Teste:** `TestTEN003DiscoveryPagination`.

### TEN-004: Validação de filial e rollback de endereço órfão
- **Pré-condição:** Organização `cafe`, proprietário e conta sem vínculo.
- **Passos:** Tentar criar filial sem autorização e com dados inválidos; cadastrar os dados válidos e repetir o cadastro; contar filiais e endereços.
- **Entradas:** CNPJ `12345678000100`, CNPJ do proprietário `12345678000199` e CEP `12345678`. Variações inválidas: CNPJ com pontuação, CNPJ do proprietário vazio e CEP `123`.
- **Esperado:** HTTP 403 sem vínculo, 400 para dados inválidos, 201 com verificação pendente e 409 para duplicata. O banco mantém somente uma filial e um endereço.
- **Teste:** `TestTEN004BranchCreationValidationAndRollback`.

### TEN-005: Aprovação legal, visibilidade pública e auditoria
- **Pré-condição:** Organização com uma filial não verificada, proprietário e sysadmin.
- **Passos:** Consultar filial, horários, listagem e cardápio sem token; consultar a filial como proprietário. Tentar aprovar como proprietário e depois como sysadmin; repetir as consultas públicas da filial, da listagem e do cardápio.
- **Entradas:** ID da filial, CNPJ do proprietário `12345678000199` e tokens de proprietário e sysadmin.
- **Esperado:** Antes da aprovação, filial, horários e cardápio públicos retornam 404 e a listagem é vazia. O proprietário vê os dados administrativos, mas recebe 403 ao aprovar. A aprovação pelo sysadmin retorna 200, registra a ação com ator e filial e permite a publicação sem expor `owner_cnpj`.
- **Teste:** `TestTEN005ApprovalAndPublicVisibility`.

### TEN-006: Autoridade e revogação do gerente da organização
- **Pré-condição:** Organização com proprietário e uma conta elegível para gerência.
- **Passos:** Vincular o gerente, atualizar a organização com o token dele, tentar arquivá-la, revogar o vínculo e repetir a atualização com o mesmo token.
- **Entradas:** `account_id` do gerente; os nomes `Atualizado` e `Negado`.
- **Esperado:** Atualização autorizada retorna 200, arquivamento retorna 403 e revogação retorna 204. A nova tentativa com o token ainda válido recebe 403 e o nome permanece `Atualizado`.
- **Teste:** `TestTEN006OrganizationManagerRevocation`.

### TEN-007: Escopo de gerência restrito à filial
- **Pré-condição:** Duas filiais e uma conta sem papel na organização.
- **Passos:** Vincular a conta à primeira filial. Tentar atualizar ambas e arquivar a primeira. Revogar o vínculo e tentar atualizá-la novamente.
- **Entradas:** IDs das duas filiais, ID do gerente e alteração de nome.
- **Esperado:** Atualização somente na filial atribuída; a outra filial e o arquivamento retornam 403. Após revogação, o mesmo token deixa de autorizar a alteração na primeira filial.
- **Teste:** `TestTEN007BranchManagerIsolation`.

### TEN-008: Isolamento entre organizações e recursos filhos
- **Pré-condição:** Duas organizações de proprietários distintos, cada uma com uma filial; há uma mesa na filial alheia.
- **Passos:** Tentar alterar a filial alheia com o slug próprio e depois com o slug correto usando o token do outro proprietário. Tentar excluir a mesa alheia pela rota da filial própria e conferi-la na filial original.
- **Entradas:** ID da filial alheia e ID de sua mesa número 7.
- **Esperado:** Associação incorreta de slug e filial retorna 404; falta de autoridade sobre o slug correto retorna 403. Exclusão cruzada da mesa retorna 404 e a mesa permanece na listagem correta.
- **Teste:** `TestTEN008CrossTenantAndChildIDs`.

### TEN-009: Substituição de endereço preserva dados após rejeição
- **Pré-condição:** Filial com CEP `12345678` e número `12A`.
- **Passos:** Atualizar o endereço, enviar outra alteração com CEP inválido e consultar a filial.
- **Entradas:** CEP `87654321`, rua `Nova Rua`, número `42B`, complemento `Fundos`; depois CEP `999`.
- **Esperado:** HTTP 200 e 400, respectivamente. A consulta retorna todos os campos do endereço válido, inclusive o número alfanumérico.
- **Teste:** `TestTEN009BranchAddressReplacement`.

### TEN-010: Validação, visibilidade e remoção de telefones
- **Pré-condição:** Filial sem contatos.
- **Passos:** Cadastrar contatos inválidos e um contato válido sem tipo explícito. Listar com e sem token, remover o contato e listar novamente.
- **Entradas:** `+5511999999999`, tipo inválido `fax` e número inválido `abc`.
- **Esperado:** HTTP 400 para entradas inválidas; cadastro válido retorna 201 com tipo `mobile`. Listagem anônima retorna 401. Após HTTP 204 na remoção, a listagem autenticada está vazia.
- **Teste:** `TestTEN010BranchContactLifecycle`.

### TEN-011: Horário semanal íntegro, incluindo funcionamento noturno
- **Pré-condição:** Filial existente e token de proprietário.
- **Passos:** Publicar a semana completa, enviar quatro substituições inválidas e consultar o horário salvo.
- **Entradas:** Dia 0 aberto das `18:00` às `02:00`, demais dias fechados. Variações: lista vazia, dia repetido, hora `25:00` e dia fechado com horários preenchidos.
- **Esperado:** A semana válida recebe 200; as quatro variações recebem 400. A consulta mantém os sete dias, o intervalo noturno e o último dia fechado.
- **Teste:** `TestTEN011HoursReplacement`.

### TEN-012: Unicidade de mesa por filial e preservação do histórico
- **Pré-condição:** Duas filiais da mesma organização, sem mesas.
- **Passos:** Criar mesa, repetir o número na mesma filial e na outra, alterar capacidade, registrar pedido associado e arquivar a mesa.
- **Entradas:** Mesa número 2 com 4 lugares; atualização para 6 lugares e pedido local que guarda o número 2.
- **Esperado:** Duplicata na mesma filial retorna 409; na outra filial retorna 201. A atualização mantém o ID da mesa. Após arquivamento, a mesa some da listagem ativa e o pedido ainda referencia a mesa e seu número histórico.
- **Teste:** `TestTEN012TableUniquenessAndHistory`.

### TEN-013: Arquivamento atômico com auditoria e histórico
- **Pré-condição:** Organização com filial e pedido existente.
- **Passos:** Forçar falha na gravação da auditoria de arquivamento da organização. Tentar arquivar, verificar visibilidade, restabelecer a auditoria e arquivar a filial e a organização.
- **Entradas:** IDs da filial e do pedido; ação de auditoria `restaurant.archive`.
- **Esperado:** A falha retorna 500 e a organização permanece consultável. Com a auditoria disponível, ambos os arquivamentos retornam 204, os recursos deixam de ser consultáveis e o pedido mantém sua referência à filial.
- **Teste:** `TestTEN013ArchiveRollbackAndHistory`.

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

### CAT-011: Cadastro de Ingrediente Base no Catálogo Global (`ingredient`)
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/ingredients` com `{"name": "Orégano Desidratado"}`.
- **Esperado:** HTTP 201 Created. Registro na tabela `ingredient`.

### CAT-012: Composição de Ficha Técnica (`item_ingredient`) com Medida e Quantidade
- **Pré-condição:** Item ID 5 e Ingrediente ID 2 existentes.
- **Passos:** POST `/items/5/ingredients`.
- **Entradas:** `{"ingredient_id": 2, "is_removable": true, "quantity": 100, "measure": "gramas"}`.
- **Esperado:** HTTP 201 Created. Chave composta `(item_id, ingredient_id)` gravada com sucesso.

### CAT-013: Ingrediente Não Removível na Ficha Técnica (`is_removable = false`)
- **Pré-condição:** Ficha técnica configurada com `is_removable = false`.
- **Passos:** Consultar ficha técnica do produto.
- **Esperado:** Flag retornada como `false` para que a UI do cliente desabilite a opção de remover.

### CAT-014: Criação de Opcional/Adicional Direto (`item_side`)
- **Pré-condição:** Item ID 5 existente.
- **Passos:** POST `/items/5/sides`.
- **Entradas:** `{"name": "Borda Recheada", "amount": 1, "price": 8.00}`.
- **Esperado:** HTTP 201 Created com registro gravado na tabela `item_side`.

### CAT-015: Customização de Pedido com Opcional e Ingrediente Removido (`order_item.notes`)
- **Pré-condição:** Item ID 5 com adicional `item_side` ("Borda Recheada") e ingrediente removível ("Cebola").
- **Passos:** Incluir item no pedido enviando customização nas observações de produção do item.
- **Entradas:** `{"item_id": 5, "quantity": 1, "unit_price": 53.00, "notes": "Borda Recheada (+R$ 8.00); Sem cebola"}`.
- **Esperado:** HTTP 201 Created. Registro inserido em `order_item` contendo o campo `notes` com as instruções de preparo e opcionais sem necessidade de joins transitórios.

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

### CAT-035: Remoção de Adicional (`item_side`)
- **Pré-condição:** Adicional `item_side` ID 5 cadastrado para o item ID 1.
- **Passos:** DELETE `/items/1/sides/5`.
- **Esperado:** HTTP 204 No Content. Registro excluído da tabela `item_side`.

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

### ORD-009: Pedido com Múltiplas Linhas de Itens e Instruções de Cozinha (`order_item.notes`)
- **Pré-condição:** Item A (20.00) e Item B (15.00).
- **Passos:** Criar pedido com 2x Item A e 1x Item B com observações `"notes": "Sem cebola; ponto bem passado"`.
- **Esperado:** `subtotal = 55.00`, `total = 55.00`. O campo `notes` em `order_item` armazena a personalização de preparo e remoção de ingredientes.

### ORD-010: Bloqueio de Pedido Contendo Produto Indisponível
- **Pré-condição:** Item com `is_available = false`.
- **Passos:** Submeter pedido com o item.
- **Esperado:** HTTP 400 Bad Request `item is currently unavailable`.

### ORD-011: Avaliação Dupla: Avaliação de Item (`order_item.review`) e Avaliação Geral do Pedido (`order.review`)
- **Pré-condição:** Pedido ID 100 em status `COMPLETED` contendo o item ID 15.
- **Passos:**
  1. POST `/orders/100/review` com `{"review": 5}`.
  2. POST `/order-items/15/review` com `{"review": 4}`.
- **Esperado:** HTTP 200 OK em ambos. Coluna `order.review` atualizada para `5` e coluna `order_item.review` atualizada para `4`.

### ORD-012: Validação de Escala de Avaliação (1 a 5) em Pedido e Itens
- **Pré-condição:** Pedido finalizado.
- **Passos:** Submeter avaliação com nota 0 ou 7 para `/orders/100/review` ou `/order-items/15/review`.
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

---

## Módulo 7: Interface do Usuário, Navegação e Interações do Website (WEB-001 a WEB-050)

### WEB-001: Clique no Botão Flutuante do Carrinho para Abertura da Gaveta Lateral
- **Pré-condição:** Usuário na página inicial ou na página do restaurante com o botão de carrinho flutuante visível.
- **Passos:** Clicar no botão flutuante de carrinho (com ícone e badge de contagem).
- **Esperado:** A gaveta lateral do carrinho desliza a partir da borda direita da tela (`translate-x-0`), exibindo o cabeçalho "Seu Pedido" e a lista de itens/estado vazio.

### WEB-002: Fechamento do Carrinho pelo Botão 'X' no Cabeçalho da Gaveta
- **Pré-condição:** Gaveta lateral do carrinho aberta na tela.
- **Passos:** Clicar no botão 'X' (fechar) localizado no canto superior direito do cabeçalho da gaveta.
- **Esperado:** A gaveta lateral do carrinho recolhe suavemente para fora da tela (`translate-x-full`) e o foco retorna à página.

### WEB-003: Fechamento da Gaveta do Carrinho ao Clicar no Backdrop / Fundo Escurecido
- **Pré-condição:** Gaveta lateral do carrinho aberta com backdrop escurecido cobrindo a área da página.
- **Passos:** Clicar em qualquer ponto da área externa à gaveta (no elemento backdrop).
- **Esperado:** A gaveta fecha imediatamente e o backdrop desaparece com transição de opacidade.

### WEB-004: Fechamento da Gaveta do Carrinho pela Tecla 'Escape'
- **Pré-condição:** Gaveta lateral do carrinho aberta e com foco ativo.
- **Passos:** Pressionar a tecla `Escape` no teclado físico.
- **Esperado:** O evento de teclado é capturado e a gaveta do carrinho é fechada.

### WEB-005: Renderização de Estado Vazio do Carrinho ("Seu carrinho está vazio")
- **Pré-condição:** Nenhum item adicionado à store do carrinho (`cart.items.length === 0`).
- **Passos:** Abrir a gaveta do carrinho.
- **Esperado:** Exibição do ícone ilustrativo de carrinho vazio, texto "Seu carrinho está vazio", instrução "Adicione itens para continuar" e valor total em "R$ 0,00".

### WEB-006: Adição de Item Simples ao Carrinho e Atualização do Badge do Navbar
- **Pré-condição:** Usuário visualizando itens do cardápio.
- **Passos:** Clicar no botão de adicionar de um item do cardápio.
- **Esperado:** O contador numérico no badge do carrinho é incrementado para 1 e o item passa a constar na listagem do carrinho com seu nome e preço unitário.

### WEB-007: Botão de Incremento (+) de Quantidade de Item no Carrinho
- **Pré-condição:** Item presente no carrinho com quantidade inicial = 1.
- **Passos:** Abrir o carrinho e clicar no botão com ícone de soma (+) na linha do item.
- **Esperado:** A quantidade do item é atualizada para 2, o subtotal do item é recalculado e o valor total do carrinho é atualizado.

### WEB-008: Botão de Decremento (-) de Quantidade de Item no Carrinho
- **Pré-condição:** Item presente no carrinho com quantidade = 2.
- **Passos:** Abrir o carrinho e clicar no botão com ícone de subtração (-) na linha do item.
- **Esperado:** A quantidade do item diminui para 1 e o total do carrinho é recalculado de acordo.

### WEB-009: Remoção Automática do Item ao Decrementar Quantidade de 1 para 0
- **Pré-condição:** Item presente no carrinho com quantidade = 1.
- **Passos:** Clicar no botão (-) do item.
- **Esperado:** O item é completamente removido da lista do carrinho. Se não houver outros itens, a visualização alterna para o estado vazio.

### WEB-010: Atualização Reativa do Total do Pedido (R$) ao Alterar Quantidades
- **Pré-condição:** Dois itens de valores distintos no carrinho.
- **Passos:** Alterar a quantidade de um dos itens repetidas vezes.
- **Esperado:** O display de valor "Total" na base do carrinho reflete imediatamente a soma exata dos itens multiplicados por suas respectivas quantidades sem arredondamentos errados.

### WEB-011: Botão "No Restaurante" na Gaveta do Carrinho (Modalidade Salão)
- **Pré-condição:** Carrinho com ao menos 1 item adicionado.
- **Passos:** Clicar no botão "No Restaurante" no rodapé do carrinho.
- **Esperado:** O tipo de pedido é configurado como `LOCAL` e a aplicação avança para a etapa de seleção/inserção de mesa ou comanda.

### WEB-012: Botão "Deliver" / "Entrega" na Gaveta do Carrinho (Modalidade Delivery)
- **Pré-condição:** Carrinho com ao menos 1 item adicionado.
- **Passos:** Clicar no botão "Deliver" no rodapé do carrinho.
- **Esperado:** O tipo de pedido é configurado como `REMOTE` e a aplicação abre os campos de taxa de entrega e endereço.

### WEB-013: Botão "Finalizar Pedido" / Redirecionamento para Fluxo de Checkout
- **Pré-condição:** Carrinho preenchido e modalidade selecionada.
- **Passos:** Clicar no botão principal de finalizar pedido.
- **Esperado:** A gaveta fecha e o usuário é redirecionado para a tela de revisão e pagamento (`/order` ou `/checkout`).

### WEB-014: Persistência do Estado do Carrinho Entre Navegações de Páginas
- **Pré-condição:** Itens adicionados ao carrinho na página do restaurante.
- **Passos:** Navegar para a lista de restaurantes e depois retornar à página do restaurante.
- **Esperado:** O contador do carrinho permanece inalterado e os itens continuam preservados na memória/store.

### WEB-015: Clique no Logo Pratello na Barra de Navegação (Navegação para Início)
- **Pré-condição:** Usuário em qualquer página interna ou detalhe de restaurante.
- **Passos:** Clicar no logotipo Pratello posicionado na barra de navegação superior.
- **Esperado:** A rota corrente é alterada para a página principal (`/` ou `/home`), restaurando a visualização inicial.

### WEB-016: Clique no Link "Início" no Navbar
- **Pré-condição:** Links de navegação desktop visíveis no navbar.
- **Passos:** Clicar no link textual "Início".
- **Esperado:** Redirecionamento instantâneo para a página inicial com indicador visual de link ativo.

### WEB-017: Clique no Link "Cardápio" no Navbar
- **Pré-condição:** Usuário na visualização de um restaurante com link "Cardápio" visível.
- **Passos:** Clicar no link "Cardápio".
- **Esperado:** A página realiza scroll suave ou navega até a seção do catálogo/cardápio do restaurante.

### WEB-018: Clique no Link "Delivery" no Navbar
- **Pré-condição:** Usuário em qualquer página da aplicação.
- **Passos:** Clicar no link "Delivery" no navbar.
- **Esperado:** A página carrega a listagem geral de estabelecimentos disponíveis para entrega (`/restaurantes`).

### WEB-019: Botão de Menu Mobile (Hambúrguer) em Telas Menores
- **Pré-condição:** Visualização em viewport de dispositivo móvel (< 768px).
- **Passos:** Clicar no ícone de menu hambúrguer no canto do navbar.
- **Esperado:** O menu mobile expande exibindo os links de navegação acessíveis por toque.

### WEB-020: Chip de Endereço no Navbar: Abertura do Popover de Seleção
- **Pré-condição:** Navbar visível com o chip de endereço corrente (ex: "Jaboatão dos Guararapes, PE").
- **Passos:** Clicar no botão do chip de endereço.
- **Esperado:** O popover `AddressSection` abre abaixo do chip e a seta indicadora rotaciona 180 graus.

### WEB-021: Popover de Endereço: Seleção de Endereço Existente e Fechamento
- **Pré-condição:** Popover de endereço aberto exibindo lista de endereços previamente cadastrados.
- **Passos:** Clicar em um dos endereços sugeridos na lista.
- **Esperado:** O endereço selecionado passa a ser exibido no chip do navbar, o popover fecha e as taxas de entrega são recalculadas.

### WEB-022: Popover de Endereço: Botão "Adicionar novo endereço"
- **Pré-condição:** Popover de endereço aberto.
- **Passos:** Clicar no botão "Adicionar novo endereço" ou botão de ação de CEP.
- **Esperado:** Exibição do formulário de entrada com campos para CEP, rua, bairro e número.

### WEB-023: Popover de Endereço: Fechamento ao Clicar Fora (Click Outside)
- **Pré-condição:** Popover de endereço aberto.
- **Passos:** Clicar em qualquer elemento externo ao popover no corpo da página.
- **Esperado:** O popover de endereços fecha imediatamente sem alterar o endereço ativo.

### WEB-024: Botão "Entrar" no Navbar: Abertura do Menu Lateral de Acesso
- **Pré-condição:** Usuário anônimo (não logado). Botão "Entrar" estilizado em laranja suave visível no navbar.
- **Passos:** Clicar no botão "Entrar".
- **Esperado:** A sidebar lateral esquerda de autenticação (`LeftSidebarSection`) abre com animação suave, exibindo as opções de login.

### WEB-025: Menu Lateral de Acesso: Botão Fechar (X)
- **Pré-condição:** Menu lateral de acesso aberto na tela.
- **Passos:** Clicar no botão com ícone de fechar (X) no cabeçalho da sidebar.
- **Esperado:** A sidebar esquerda é recolhida e a navegação da página volta ao estado interativo.

### WEB-026: Menu Lateral de Acesso: Opção "Entrar como Cliente"
- **Pré-condição:** Menu lateral de acesso aberto.
- **Passos:** Clicar no botão "Entrar como Cliente".
- **Esperado:** A sidebar fecha e o usuário é redirecionado para a tela de login (`/login`) em modo cliente.

### WEB-027: Menu Lateral de Acesso: Opção "Entrar como Funcionário"
- **Pré-condição:** Menu lateral de acesso aberto.
- **Passos:** Clicar no botão "Entrar como Funcionário".
- **Esperado:** A sidebar fecha e o usuário é direcionado para a interface de autenticação do time operacional/PDV.

### WEB-028: Tela de Login: Foco e Digitação no Campo de Input "E-mail"
- **Pré-condição:** Usuário na página de login (`Auth.svelte`).
- **Passos:** Clicar no campo de texto de e-mail e digitar `cliente@exemplo.com`.
- **Esperado:** O campo recebe foco com realce de borda visual e o texto digitado é refletido com precisão sem travamentos.

### WEB-029: Tela de Login: Validação de E-mail com Formato Inválido ao Submeter
- **Pré-condição:** Campo de e-mail preenchido com string sem formato válido (ex: `cliente_invalido`).
- **Passos:** Clicar no botão "Entrar".
- **Esperado:** O formulário impede a submissão e o navegador ou componente apresenta mensagem de validação de formato.

### WEB-030: Tela de Login: Foco e Digitação no Campo de Input "Senha"
- **Pré-condição:** Usuário no formulário de login.
- **Passos:** Clicar no campo de senha e digitar uma sequência de caracteres.
- **Esperado:** Os caracteres são ocultados por padrão através de caracteres de máscara (tipo `password`).

### WEB-031: Tela de Login: Botão de Alternar Visibilidade da Senha (Ícone Olho)
- **Pré-condição:** Campo de senha preenchido.
- **Passos:** Clicar no botão com ícone de olho posicionado na extremidade direita do campo.
- **Esperado:** O tipo do input alterna de `password` para `text`, revelando a senha em texto legível; ao clicar novamente, volta a mascarar.

### WEB-032: Tela de Login: Checkbox Interativo "Lembrar de mim"
- **Pré-condição:** Checkbox "Lembrar de mim" desmarcado por padrão.
- **Passos:** Clicar no checkbox ou no rótulo de texto adjacente.
- **Esperado:** O checkbox alterna para o estado marcado (checked) com feedback visual de preenchimento.

### WEB-033: Tela de Login: Link "Esqueceu sua senha?"
- **Pré-condição:** Usuário visualizando o bloco de login.
- **Passos:** Clicar no link de recuperação de acesso "Esqueceu sua senha?".
- **Esperado:** Redirecionamento para a página ou modal de redefinição de credenciais de acesso.

### WEB-034: Tela de Login: Botão Submit "Entrar" com Campos Vazios
- **Pré-condição:** Campos de e-mail e senha sem nenhum caractere preenchido.
- **Passos:** Clicar no botão principal de submissão "Entrar".
- **Esperado:** O envio da requisição é abortado e os campos obrigatórios vazios são destacados visualmente.

### WEB-035: Tela de Login: Botão Submit "Entrar" com Credenciais Válidas
- **Pré-condição:** E-mail e senha válidos inseridos nos respectivos campos.
- **Passos:** Clicar no botão "Entrar" ou pressionar `Enter`.
- **Esperado:** O botão exibe estado de carregamento, a autenticação é concluída e o usuário é redirecionado à página inicial com o avatar logado no navbar.

### WEB-036: Tela de Login: Link "Crie uma conta" / "Novo aqui?"
- **Pré-condição:** Usuário na página de login.
- **Passos:** Clicar no link "Crie uma conta".
- **Esperado:** Navegação para o formulário de cadastro de novos usuários com campos adicionais de nome e documento.

### WEB-037: Perfil Autenticado no Navbar: Exibição de Iniciais e Nome
- **Pré-condição:** Sessão iniciada com conta autenticada (ex: "Gregório Silva").
- **Passos:** Observar a área esquerda do navbar.
- **Esperado:** O botão "Entrar" é substituído pelo chip de perfil contendo a inicial "G", saudação "Olá Gregório" e indicador de status verde.

### WEB-038: Menu Lateral do Usuário Autenticado: Botão "Fazer logout"
- **Pré-condição:** Usuário logado na aplicação.
- **Passos:** Clicar no chip de perfil para abrir a sidebar e em seguida clicar no botão "Fazer logout".
- **Esperado:** A sessão é limpa na store e no storage, a sidebar fecha e o navbar retorna ao estado inicial deslogado com o botão "Entrar".

### WEB-039: Campo de Busca de Restaurantes: Digitação e Filtragem em Tempo Real
- **Pré-condição:** Usuário na página principal de restaurantes (`RestaurantsPage`).
- **Passos:** Clicar no campo de busca com ícone de lupa e digitar "Burger".
- **Esperado:** A lista de restaurantes é filtrada imediatamente, exibindo apenas estabelecimentos cujo nome ou especialidade contenham o termo.

### WEB-040: Campo de Busca de Restaurantes: Limpeza do Texto de Pesquisa
- **Pré-condição:** Campo de pesquisa preenchido com filtro ativo.
- **Passos:** Apagar o texto da busca por completo.
- **Esperado:** A grade de restaurantes restaura a exibição de todos os estabelecimentos disponíveis na região.

### WEB-041: Botão "Explorar restaurantes" no Hero Banner da Página Principal
- **Pré-condição:** Usuário no topo da página principal.
- **Passos:** Clicar no botão de destaque com gradiente âmbar "Explorar restaurantes".
- **Esperado:** A visualização desliza com scroll animado suave até a âncora da seção de restaurantes próximos (`#restaurantes-proximos`).

### WEB-042: Carrossel de Destaques: Botões de Navegação Próximo/Anterior
- **Pré-condição:** Carrossel de banners de restaurantes em exibição.
- **Passos:** Clicar nas setas ou indicadores laterais de avanço e retrocesso de slide.
- **Esperado:** O banner ativo transiciona suavemente apresentando o próximo destaque promocional com título e chamada para ação atualizados.

### WEB-043: Card de Restaurante: Clique para Acessar o Estabelecimento
- **Pré-condição:** Lista de restaurantes visível na página.
- **Passos:** Clicar sobre o card de um restaurante específico (ex: "Pizzaria Bela Vista").
- **Esperado:** A rota muda para a visualização do restaurante (`/r/:id`), carregando seu cardápio próprio e layout temático.

### WEB-044: Filtros de Categoria no Cardápio (Combos, Pratos, Bebidas, Sobremesas)
- **Pré-condição:** Usuário visualizando a página de cardápio do restaurante.
- **Passos:** Clicar no botão de categoria de filtro "Bebidas".
- **Esperado:** O botão da categoria assume estado ativo selecionado e a lista apresenta apenas os itens associados à categoria clicada.

### WEB-045: Card de Item/Prato: Clique para Abrir o Modal de Detalhes
- **Pré-condição:** Itens visíveis no catálogo.
- **Passos:** Clicar no card de um item do cardápio (ex: "Hambúrguer Especial").
- **Esperado:** O modal `ItemModal` abre em sobreposição central, exibindo foto ampliada, descrição completa, preço e opções de customização.

### WEB-046: Modal do Item: Seleção e Desseleção de Opcionais/Adicionais
- **Pré-condição:** Modal de item aberto contendo grupo de adicionais (ex: "Bacon Extra", "Molho Especial").
- **Passos:** Clicar no seletor de um opcional e em seguida clicar novamente.
- **Esperado:** O item adicional é marcado e o preço base no botão de adicionar sofre acréscimo proporcional; ao desmarcar, o valor decrementa de volta ao base.

### WEB-047: Modal do Item: Botões de Controle de Quantidade (+ e -)
- **Pré-condição:** Modal de item aberto com quantidade padrão igual a 1.
- **Passos:** Clicar no botão (+) duas vezes e depois no botão (-) uma vez.
- **Esperado:** O mostrador numérico de quantidade avança para 3 e depois recua para 2, recalculando o valor total daquele prato no modal.

### WEB-048: Modal do Item: Botão "Adicionar ao Pedido"
- **Pré-condição:** Modal de item aberto com quantidade e opcionais configurados.
- **Passos:** Clicar no botão principal de ação "Adicionar ao Pedido (R$ XX,XX)".
- **Esperado:** O prato com todas as suas personalizações é incluído na store do carrinho, o modal se fecha automaticamente e o carrinho é atualizado.

### WEB-049: Modal do Item: Fechamento via Botão de Fechar ou Clique no Fundo
- **Pré-condição:** Modal de item aberto na tela.
- **Passos:** Clicar no botão (X) do modal ou no backdrop escurecido ao redor da caixa de diálogo.
- **Esperado:** O modal é dispensado sem adicionar nada ao carrinho e mantendo os itens anteriores intactos.

### WEB-050: Checkout: Seleção de Método de Pagamento e Botão de Confirmação
- **Pré-condição:** Usuário na página de checkout com itens no carrinho.
- **Passos:** Clicar no botão do método de pagamento desejado (PIX, Cartão ou Dinheiro) e clicar em "Confirmar Pedido".
- **Esperado:** O método escolhido recebe realce ativo de seleção e o clique no botão de confirmação dispara a criação do pedido, abrindo a tela de acompanhamento e comprovante.
