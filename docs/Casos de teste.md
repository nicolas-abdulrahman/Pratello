# Plano e Especificação Detalhada de Testes — Pratello (DBML Alinhado)

> **Versão:** 0.1.18
> **Classificação:** Suíte de Testes Profunda (Deep Testing Specification)  
> **Total de casos documentados:** 245

## Resumo Executivo e Metodologia de Teste

A suíte está estruturada em **7 módulos** cobrindo Caminho Feliz (*Happy Path*), Condições de Limite (*Boundary Testing*), Casos Negativos de Erro, Integridade Referencial (*ACID/FK*), Idempotência, Concorrência e Usabilidade Frontend:

1. **Módulo 1: Identidade, Contas, Papéis e Autoridade (IDP)** — Casos `IDP-001` a `IDP-027`: cadastro, contas sombra, CPF, métodos de acesso, sessões e autorização.
2. **Módulo 2: Estabelecimentos e Filiais (TEN)** — Casos `TEN-001` a `TEN-019`: organizações, aprovação de filiais, isolamento, vínculos por filial, arquivamento e resiliência operacional.
3. **Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT)** — Casos `CAT-001` a `CAT-055` (55 testes): layout visual, categorias, itens, ingredientes, grupos de opções, combos e promoções. Inclui regressões de isolamento entre restaurantes e filiais, privacidade de custos, validação de personalizações, preços, snapshots e concorrência.
4. **Módulo 4: Pedidos Anônimos, Especialização e Itens (ORD)** — Casos `ORD-001` a `ORD-040` (40 casos): rascunhos, autorização por pedido, salão com mesa opcional, entrega, checkout atômico, QR reutilizável, atribuição de conversão e transições operacionais. Integrações futuras estão marcadas no módulo.
5. **Módulo 5: Pagamentos, PIX, Cupons e Fidelidade (FIN)** — Casos `FIN-001` a `FIN-040` (40 casos): pagamento simulado, PIX pendente e confirmação, idempotência, devoluções, cupons e banners, descontos sem acúmulo e resgate concorrente de moedas. Cenários de integração futura estão identificados no módulo.
6. **Módulo 6: Avaliações de Itens, Combos e Restaurantes (REV)** — Casos `REV-001` a `REV-014` (14 casos automatizados): escalas, autorização por compra, comentários, concorrência, avaliações da organização e integridade dos totais.
7. **Módulo 7: Interface do Usuário, Navegação e Interações do Website (WEB)** — Casos `WEB-001` a `WEB-050` (50 roteiros manuais): carrinho, navegação, endereços, autenticação, cardápio e checkout.

### Contagem e execução verificadas nesta revisão

| Módulo | Casos documentados | PASS com vínculo por caso | XFAIL com vínculo por caso | Planejados | Roteiros manuais WEB | Sem resultado verificado por ID |
|---|---:|---:|---:|---:|---:|---:|
| IDP | 27 | 26 | 1 | 0 | 0 | 0 |
| TEN | 19 | 19 | 0 | 0 | 0 | 0 |
| CAT | 55 | 33 | 0 | 0 | 0 | 22 |
| ORD | 40 | 35 | 0 | 2 | 0 | 3 |
| FIN | 40 | 32 | 0 | 4 | 0 | 4 |
| REV | 14 | 14 | 0 | 0 | 0 | 0 |
| WEB | 50 | 0 | 0 | 0 | 50 | 0 |
| **Total** | **245** | **159** | **1** | **6** | **50** | **29** |

Os números da tabela contam IDs distintos. Os testes Go agrupados de CAT, ORD e FIN passaram com PostgreSQL e o detector de corridas (`go test -race -count=1 -v ./server/internal/menu ./server/internal/order ./server/internal/sales ./server/internal/transaction`). Um grupo aprovado só conta para os IDs com asserções correspondentes. Os roteiros WEB não foram executados.

| Módulo | IDs verificados como PASS | Evidência automatizada |
|---|---|---|
| CAT | `CAT-002`, `CAT-004–005`, `CAT-007–008`, `CAT-011–012`, `CAT-014–019`, `CAT-021–022`, `CAT-025`, `CAT-031–033`, `CAT-035`, `CAT-043–055` | `TestMenuPostgresLifecycle`, `TestSelectionPricingAndSnapshot`, `TestCAT050And051GroupCountsAndBranchStock`, `TestCAT053FutureSaleDoesNotOverrideEarly` e o snapshot persistido em `TestOrderCheckoutRevalidatesAndFreezesSnapshot`. |
| ORD | `ORD-001–002`, `ORD-004–023`, `ORD-027–035`, `ORD-037–040` | Testes HTTP, checkout e concorrência em `server/internal/order`, incluindo `TestORD013ComboPreparationLines`; auditoria de cupom em `TestPostgresCouponRedemption`; avaliações em `TestReviewCasesPostgres` (`REV-001`, `REV-004–005`, `REV-008–009`), também reexecutado com PostgreSQL. |
| FIN | `FIN-003–004`, `FIN-006`, `FIN-009–011`, `FIN-013–020`, `FIN-022–027`, `FIN-029–040` | `TestPostgresPaymentLifecycle`, `TestPostgresCouponRedemption`, `TestFIN024And030ConcurrentRedemptions`, `TestBestDiscountPerLine`, `TestFIN034ComboCouponsAndExtras`, `TestFIN025Through027CouponEligibility`, PIX em `TestOrderQRCodeCheckoutAndAnalyticsPostgres` e fidelidade por CPF em `TestFIN013014016GuestCPFLoyaltyAndSignup`. |

Para executar a suíte com PostgreSQL, configure `AUTH_TEST_DATABASE_URL` e rode `go test -count=1 -race ./...`. A contagem por ID não equivale ao número de funções Go. O XFAIL `IDP-021` é reportado como `SKIP` pelo Go após reproduzir a falha conhecida.

---

## Módulo 1: Identidade, Contas, Papéis e Autoridade (IDP-001 a IDP-027)

Os casos de identidade ficam em `server/internal/auth/auth_casos_de_testes_test.go`, `server/internal/user/user_casos_de_testes_test.go` e `server/internal/restaurants/restaurants_casos_de_testes_test.go`. Execute `go test -race ./server/internal/auth ./server/internal/user ./server/internal/restaurants`; os testes PostgreSQL usam `AUTH_TEST_DATABASE_URL` e um esquema temporário. A verificação de e-mail/WhatsApp é simulada. `XFAIL` registra a falha conhecida sem ocultar resultados inesperados.

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

### IDP-020: Sysadmin e Administrador são papéis distintos
- **Pré-condição:** Conta comum, sysadmin e conta alvo sem atribuições globais.
- **Passos:** Tentar promover o alvo a sysadmin sem token e com uma conta comum. Com o token de sysadmin, tentar promover a Administrador e depois a sysadmin.
- **Entradas:** ID da conta alvo nos endpoints `/sysadmins` e `/administrators`.
- **Esperado:** HTTP 401 e 403 para as primeiras tentativas; sysadmin não pode conceder Administrador (403), mas pode conceder sysadmin (201). Somente `is_sysadmin` fica ativo no alvo.
- **Teste:** `TestIDP020SysadminRoleAuthority` em `server/internal/user/user_casos_de_testes_test.go`.

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

### IDP-027: Analista tem vínculo somente de leitura
- **Pré-condição:** Organização com proprietário e uma conta sem vínculo.
- **Passos:** O proprietário cadastra o Analista. A conta consulta seus vínculos, tenta alterar a organização e nomear um proprietário. O proprietário revoga a atribuição.
- **Entradas:** ID do Analista e token da mesma conta antes e depois da revogação.
- **Esperado:** O vínculo `ANALYST` aparece em `/me/restaurants`; ambas as mutações retornam 403. Após revogação, a listagem não contém mais o vínculo.
- **Teste:** `TestIDP027AnalystMembershipDoesNotGrantMutation` em `server/internal/user/user_casos_de_testes_test.go`.

---

## Módulo 2: Estabelecimentos e Filiais (TEN-001 a TEN-019)

Os testes estão em `server/internal/restaurants/restaurants_casos_de_testes_test.go` e `server/internal/user/user_casos_de_testes_test.go`. Com `AUTH_TEST_DATABASE_URL` apontando para um PostgreSQL que permita criar esquemas, execute `go test -count=1 -race ./server/internal/restaurants ./server/internal/user -run '^TestTEN'`. Cada caso cria e remove seu próprio esquema.

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

### TEN-006: Administrador global atua em empresas distintas e perde acesso ao ser revogado
- **Pré-condição:** Duas empresas de proprietários distintos; Administrador existente e conta alvo sem vínculo.
- **Passos:** Tentar conceder Administrador como proprietário e revogar o único Administrador. Conceder a atribuição como Administrador, alterar ambas as empresas e criar e aprovar filial. Revogar a nova atribuição e repetir uma atualização com o mesmo token.
- **Entradas:** `account_id` do alvo e slugs `cafe` e `other`.
- **Esperado:** O proprietário recebe 403 e a tentativa de remover o último Administrador recebe 409. O novo Administrador pode atuar nas duas empresas sem vínculo local. Após revogação (204), o mesmo token recebe 403 e o último nome autorizado é preservado.
- **Teste:** `TestTEN006GlobalAdministratorAndRevocation`.

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

### TEN-014: Uma conta acumula vínculos independentes por empresa e filial
- **Pré-condição:** Três empresas e uma conta comum sem vínculo.
- **Passos:** Contratar a conta em duas filiais de `cafe` e atribuir gerência a uma filial de `bar`. Conferir seus vínculos e permissões; desativar um único registro de funcionário.
- **Entradas:** Um `account_id`, posição `WAITER`, dois IDs de filial de `cafe`, um de `bar` e nenhum vínculo com `diner`.
- **Esperado:** As duas contratações funcionam e a duplicata na mesma filial retorna 409. A conta pode listar mesas nas filiais onde trabalha, mas não gerenciá-las; pode alterar a filial gerenciada, mas não a organização `bar`. Em `diner`, não ganha acesso operacional. Desativar um registro só revoga a filial correspondente.
- **Teste:** `TestTEN014EmployeeAndManagerMembershipsAreIndependent` em `server/internal/user/user_casos_de_testes_test.go`.

### TEN-015: Cancelamento e recuperação com pool ocupado
- **Pré-condição:** Há uma organização cadastrada e o pool de teste permite uma conexão, reservada pelo teste.
- **Passos:** Consultar a organização com prazo limitado, aguardar o pedido entrar na fila do pool, deixar o prazo expirar, liberar a conexão e repetir a consulta.
- **Entradas:** `GET /restaurants/cafe`, primeiro com prazo de um segundo e depois com contexto novo.
- **Esperado:** O primeiro pedido termina sem sucesso enquanto a conexão ainda está reservada. A nova consulta retorna HTTP 200 e o pool termina sem conexões em uso.
- **Teste:** `TestTEN015PoolCancellationAndRecovery`.

### TEN-016: Limite de corpo JSON antes de persistir
- **Pré-condição:** Conta autenticada sem organização com o slug `large`.
- **Passos:** Enviar uma criação cujo JSON válido excede o limite do endpoint e conferir o banco e a consulta pública.
- **Entradas:** `POST /restaurants` com slug `large`, nome válido e descrição de 32 KiB.
- **Esperado:** HTTP 413 com erro de tamanho; nenhuma organização é criada e `GET /restaurants/large` retorna 404.
- **Teste:** `TestTEN016OversizedRequestDoesNotPersist`.

### TEN-017: Resposta 500 não expõe detalhes internos
- **Pré-condição:** Handler de restaurantes com pool de PostgreSQL fechado.
- **Passos:** Solicitar a listagem pública e inspecionar o status, o tipo e todos os campos da resposta.
- **Entradas:** `GET /restaurants`.
- **Esperado:** HTTP 500 e JSON com apenas `error: internal server error`, sem mensagem do driver ou detalhes de infraestrutura.
- **Teste:** `TestTEN017DatabaseErrorDoesNotLeakDetails`.

### TEN-018: Reaplicar a inicialização preserva os dados
- **Pré-condição:** Esquema de teste inicializado, com uma organização e seu primeiro proprietário.
- **Passos:** Reexecutar a inicialização, consultar o mesmo registro e vínculo e tentar cadastrar o mesmo slug novamente.
- **Entradas:** Organização `cafe` e nova tentativa de cadastro com o mesmo slug.
- **Esperado:** A inicialização termina sem erro; ID e vínculo original são preservados, sem cópias; o slug duplicado continua retornando HTTP 409.
- **Teste:** `TestTEN018SchemaReapplyPreservesOrganization`.

### TEN-019: Escritas concorrentes não duplicam mesa nem auditoria
- **Pré-condição:** Filial sem mesa número 7, proprietário autenticado e pool com duas conexões.
- **Passos:** Disparar dois cadastros simultâneos, com prazo, para a mesma mesa; consultar os registros de mesa e de auditoria.
- **Entradas:** Dois `POST /restaurants/cafe/branches/{branch_id}/tables` com `{"number":7}`.
- **Esperado:** Uma resposta HTTP 201 e uma HTTP 409, sem travamento. Persiste uma mesa e uma entrada de auditoria para a criação.
- **Teste:** `TestTEN019ConcurrentTableCreationIsAtomic`.

---

## Módulo 3: Cardápio, Catálogo, Ficha Técnica e Combos (CAT-001 a CAT-055)

**Contrato executável:** os testes do domínio `menu` usam `docs/openapi.yaml` como referência para rotas, corpos e códigos HTTP. Preços são inteiros em centavos, a exclusão de item arquiva o registro e produtos indisponíveis continuam visíveis com `is_available: false`. Nos exemplos, `cafe` é o slug do restaurante.

| Cenário adicional | Resultado esperado |
| --- | --- |
| Filial não verificada, mesmo com outra filial verificada | Consulta pública com o `branch_id` da filial não verificada retorna 404; equipe autorizada pode visualizar seu escopo. |
| Custo de produção opcional | Somente Owner, Administrator e Analyst recebem um custo informado. `PATCH` com `base_cost_cents: null` remove o custo. |
| Exclusão forçada de categoria | `force=true` move itens para `Uncategorized`; sem `force`, a categoria em uso retorna 409. |
| Estoque de opção obrigatória | Se não há opções disponíveis suficientes para o mínimo, o item deixa de ser comprável apenas naquela filial. |
| Estoque de componente de combo | A indisponibilidade ou o arquivamento de um componente obrigatório impede a compra do combo. |
| Promoções sobrepostas | A promoção da filial tem prioridade. No mesmo escopo e intervalo, vence a mais nova, sem acumular descontos. |
| Remoção por unidade do combo | A exclusão de ingrediente removível é validada e vira observação de preparo, sem dedução do preço. |

### CAT-001: Criação de Menu com Estilo Válido (`menu_style_enum`)
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** PUT `/restaurants/cafe/menu`.
- **Entradas:**
  - Estilo Clássico: `{"name": "Menu Principal", "template_style": "classic", "accent_color": "#FF5722"}`
  - Estilo Experiência: `{"name": "Menu Degustação", "template_style": "experience", "accent_color": "#4CAF50"}`
- **Esperado:** HTTP 200 OK para ambos. Um layout por restaurante.

### CAT-002: Rejeição de Menu com Estilo Fora do Enum
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter PUT `/restaurants/cafe/menu` com `{"template_style": "moderno"}`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-003: Atualização de Cor de Destaque (`accent_color`) do Menu
- **Pré-condição:** Menu cadastrado.
- **Passos:** PUT `/restaurants/cafe/menu` com `{"accent_color": "#1A2B3C"}`.
- **Esperado:** HTTP 200 OK. Cor atualizada.

### CAT-004: Criação de Categoria (`category`) Vinculada a Restaurante
- **Pré-condição:** Restaurante ID 1 existente.
- **Passos:** POST `/restaurants/cafe/categories` com `{"name": "Pizzas Tradicionais"}`.
- **Esperado:** HTTP 201 Created. Registro em `category` associado ao `restaurant_id = 1`.

### CAT-005: Listagem de Categorias do Restaurante
- **Pré-condição:** 3 categorias no Restaurante 1 e 2 categorias no Restaurante 2.
- **Passos:** GET `/restaurants/cafe/categories`.
- **Esperado:** HTTP 200 OK retornando estritamente as 3 categorias do Restaurante 1.

### CAT-006: Exclusão de Categoria Vazia
- **Pré-condição:** Categoria sem produtos vinculados.
- **Passos:** DELETE `/restaurants/cafe/categories/10`.
- **Esperado:** HTTP 204 No Content.

### CAT-007: Bloqueio de Exclusão de Categoria com Itens
- **Pré-condição:** Categoria possui itens cadastrados em `menu_item`.
- **Passos:** DELETE `/restaurants/cafe/categories/1`.
- **Esperado:** HTTP 409 Conflict, pois a categoria está em uso.

### CAT-008: Criação de Item do Cardápio (`menu_item`)
- **Pré-condição:** Restaurante 1 e Categoria 1 existentes.
- **Passos:** POST `/restaurants/cafe/items`.
- **Entradas:**
  ```json
  {
    "category_id": 1,
    "name": "Pizza Calabresa",
    "description": "Molho de tomate, calabresa fatiada e cebola",
    "base_price_cents": 4250,
    "photo_url": "https://cdn.pratello.com/calabresa.png",
    "is_available": true,
    "is_combo": false
  }
  ```
- **Esperado:** HTTP 201 Created. Registro inserido em `menu_item`.

### CAT-009: Validação de Preço de Item (`base_price_cents` não negativo)
- **Pré-condição:** Nenhuma.
- **Passos:** Submeter item com `{"base_price_cents": -500}`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-010: Alternância de Disponibilidade de Item (`is_available`)
- **Pré-condição:** Item disponível.
- **Passos:** PATCH `/restaurants/cafe/items/5` com `{"is_available": false}`.
- **Esperado:** HTTP 200 OK. O item aparece na listagem pública sem override de filial com `is_available: false`.

### CAT-011: Cadastro de Ingrediente Base no Catálogo Global (`ingredient`)
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/restaurants/cafe/ingredients` com `{"name": "Orégano Desidratado"}`.
- **Esperado:** HTTP 201 Created. Registro na tabela `ingredient`.

### CAT-012: Composição de Ficha Técnica (`item_ingredient`) com Medida e Quantidade
- **Pré-condição:** Item ID 5 e Ingrediente ID 2 existentes.
- **Passos:** POST `/restaurants/cafe/items/5/ingredients`.
- **Entradas:** `{"ingredient_id": 2, "is_removable": true, "quantity": 100, "unit": "g"}`.
- **Esperado:** HTTP 201 Created. Chave composta `(item_id, ingredient_id)` gravada com sucesso.

### CAT-013: Ingrediente Não Removível na Ficha Técnica (`is_removable = false`)
- **Pré-condição:** Ficha técnica configurada com `is_removable = false`.
- **Passos:** Consultar ficha técnica do produto.
- **Esperado:** Flag retornada como `false` para que a UI do cliente desabilite a opção de remover.

### CAT-014: Criação de Grupo de Opções (`item_side`)
- **Pré-condição:** Item ID 5 existente.
- **Passos:** POST `/restaurants/cafe/items/5/sides`.
- **Entradas:** `{"title": "Borda Recheada", "min_options": 0, "max_options": 1}`.
- **Esperado:** HTTP 201 Created com registro gravado na tabela `item_side`.

### CAT-015: Customização de Pedido com Opcional e Ingrediente Removido (`order_item.notes`)
- **Pré-condição:** Item com uma opção adicional de 800 centavos e um ingrediente removível.
- **Passos:** Gerar o snapshot no serviço `menu` com os IDs da opção e do ingrediente a excluir, a quantidade e eventuais observações.
- **Esperado:** O snapshot preserva os nomes, a exclusão e as observações, e soma o adicional ao preço sem descontar o ingrediente. A persistência na submissão do pedido pertence ao domínio de pedidos.

### CAT-016: Criação de Item do Tipo Combo (`is_combo = true`)
- **Pré-condição:** Categoria "Combos" existente.
- **Passos:** POST `/restaurants/cafe/items` com `{"category_id": 1, "name": "Combo Casal", "base_price_cents": 7990, "is_combo": true}`.
- **Esperado:** HTTP 201 Created com flag `is_combo = true`.

### CAT-017: Composição de Receita do Pacote Combo (`combo_item`)
- **Pré-condição:** Combo ID 10 ("Combo Casal") e Itens unitários ID 1 ("Pizza") e ID 2 ("Refrigerante").
- **Passos:** POST `/restaurants/cafe/items/10/combo-items`.
- **Entradas:**
  - `{"sub_item_id": 1, "quantity": 1}`
  - `{"sub_item_id": 2, "quantity": 2}`
- **Esperado:** HTTP 201 Created. Chave composta `(combo_id, included_item_id)` inserida em `combo_item`.

### CAT-018: Prevenção de Inclusão do Próprio Combo nele Mesmo
- **Pré-condição:** Combo ID 10.
- **Passos:** POST `/restaurants/cafe/items/10/combo-items` com `{"sub_item_id": 10, "quantity": 1}`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-019: Cadastro de Promoção com Desconto Fixo (`fixed_cents`)
- **Pré-condição:** Item com preço de 50.00.
- **Passos:** POST `/restaurants/cafe/items/1/sales`.
- **Entradas:**
  ```json
  {
    "discount_type": "fixed_cents",
    "discount_value": 1000,
    "starts_at": "2026-10-17T10:00:00Z",
    "ends_at": "2026-10-17T23:59:59Z"
  }
  ```
- **Esperado:** HTTP 201 Created. Promoção associada ao item.

### CAT-020: Cadastro de Promoção com Desconto Percentual (`percent`)
- **Pré-condição:** Item com preço de 50.00.
- **Passos:** POST `/restaurants/cafe/items/1/sales` com `discount_type: percent`, `discount_value: 20` e intervalo válido.
- **Esperado:** HTTP 201 Created (representando 20% de desconto).

### CAT-021: Desativação Manual de Promoção (`is_active = false`)
- **Pré-condição:** Promoção ativa cadastrada.
- **Passos:** DELETE `/restaurants/cafe/items/1/sales/{sale_id}`.
- **Esperado:** HTTP 204 No Content. A promoção cancelada não é mais aplicada ao preço.

### CAT-022: Validação de Janela Temporal de Promoção (`starts_at` e `ends_at`)
- **Pré-condição:** Promoção expirada no passado.
- **Passos:** Consultar preço efetivo do produto.
- **Esperado:** Preço integral do produto retornado, pois a data atual está fora do intervalo.

### CAT-023: Consulta Completa de Detalhes de Item com Ficha e Opcionais
- **Pré-condição:** Item cadastrado com ingredientes e adicionais.
- **Passos:** GET `/restaurants/cafe/items/1`.
- **Esperado:** HTTP 200 OK com árvore de dados aninhada.

### CAT-024: Atualização de Descrição e Foto do Item
- **Pré-condição:** Item existente.
- **Passos:** PATCH `/restaurants/cafe/items/1` com novos `description` e `photo_url`.
- **Esperado:** HTTP 200 OK.

### CAT-025: Arquivamento de Item sem Vendas Anteriores
- **Pré-condição:** Item sem pedidos em `order_item`.
- **Passos:** DELETE `/restaurants/cafe/items/99`.
- **Esperado:** HTTP 204 No Content.

### CAT-026: Arquivamento de Item com Pedidos Realizados
- **Pré-condição:** Item presente em compras na tabela `order_item`.
- **Passos:** DELETE `/restaurants/cafe/items/1`.
- **Esperado:** HTTP 204 No Content. O item sai do catálogo, e o registro referenciado pelos pedidos é preservado.

### CAT-027: Desvinculação de Ingrediente da Ficha Técnica
- **Pré-condição:** Ingrediente vinculado ao item.
- **Passos:** DELETE `/restaurants/cafe/items/1/ingredients/2`.
- **Esperado:** HTTP 204 No Content. Linha em `item_ingredient` excluída.

### CAT-028: Atualização de Quantidade de Sub-Itens no Combo
- **Pré-condição:** Combo com 1 refrigerante na receita.
- **Passos:** PUT `/restaurants/cafe/items/10/combo-items/{combo_item_id}` com `{"quantity": 3}`.
- **Esperado:** HTTP 200 OK. Quantidade alterada para 3.

### CAT-029: Remoção de Sub-Item da Receita do Combo
- **Pré-condição:** Combo cadastrado.
- **Passos:** DELETE `/restaurants/cafe/items/10/combo-items/{combo_item_id}`.
- **Esperado:** HTTP 204 No Content.

### CAT-030: Consulta Pública de Disponibilidade do Cardápio
- **Pré-condição:** 10 itens no restaurante, 2 com `is_available = false`.
- **Passos:** GET `/restaurants/cafe/items`.
- **Esperado:** Retorno com 10 itens, sendo 2 com `is_available: false`.

### CAT-031: Validação de Desconto Superior a 100% no Tipo `percent`
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/restaurants/cafe/items/1/sales` com `discount_type: percent` e `discount_value: 110`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-032: Validação de Desconto Fixo Superior ao Preço do Item
- **Pré-condição:** Item com preço de 30.00.
- **Passos:** POST `/restaurants/cafe/items/1/sales` com `discount_type: fixed_cents` e `discount_value: 4000`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-033: Tentativa de Criar Item Vinculado a Categoria de Outro Restaurante
- **Pré-condição:** Categoria pertence ao Restaurante 2.
- **Passos:** Criar item no Restaurante 1 apontando para essa categoria.
- **Esperado:** HTTP 404 Not Found, sem expor uma categoria de outro restaurante.

### CAT-034: Validação do Campo `unit` na Ficha Técnica (ex.: `g`, `ml`, `un`)
- **Pré-condição:** Nenhuma.
- **Passos:** Associar ingrediente a um item com `"unit": "g"`.
- **Esperado:** HTTP 201 Created.

### CAT-035: Remoção de Adicional (`item_side`)
- **Pré-condição:** Adicional `item_side` ID 5 cadastrado para o item ID 1.
- **Passos:** DELETE `/restaurants/cafe/items/1/sides/5`.
- **Esperado:** HTTP 204 No Content. Registro excluído da tabela `item_side`.

### CAT-036: Validação de Preço de Opção em `item_side_option`
- **Pré-condição:** Nenhuma.
- **Passos:** POST `/restaurants/cafe/items/1/sides/5/options` com `{"name": "Opção", "additional_price_cents": -300}`.
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
- **Passos:** GET `/restaurants/cafe/ingredients`.
- **Esperado:** HTTP 200 OK com array de ingredientes.

### CAT-042: Exclusão de Ingrediente sem Uso
- **Pré-condição:** Ingrediente nunca associado a nenhum item.
- **Passos:** DELETE `/restaurants/cafe/ingredients/50`.
- **Esperado:** HTTP 204 No Content.

### CAT-043: Bloqueio de Exclusão de Ingrediente Usado em Ficha Técnica
- **Pré-condição:** Ingrediente presente em `item_ingredient`.
- **Passos:** DELETE `/restaurants/cafe/ingredients/1`.
- **Esperado:** HTTP 409 Conflict.

### CAT-044: Validação de URL de Imagem Válida (`photo_url`)
- **Pré-condição:** Nenhuma.
- **Passos:** Salvar URL inválida `"not-a-url"`.
- **Esperado:** HTTP 400 Bad Request.

### CAT-045: Alteração de Preço de Item no Catálogo Não Altera Pedidos Anteriores
- **Passos:** Gerar um snapshot de checkout; depois alterar preço, nome ou opções no catálogo e gerar outro snapshot.
- **Esperado no menu:** O primeiro snapshot conserva os valores de compra; o novo reflete o catálogo atual.
- **Integração:** O Módulo 4 verifica a persistência do snapshot pelo fluxo real de submissão do pedido.

### Regressões de negócio executáveis (CAT-046 a CAT-055)

Casos de rejeição passam apenas se o sistema recusar a operação com o resultado esperado **e** preservar o estado que a operação não podia alterar. `xfail` fica reservado para falhas conhecidas e reproduzíveis, com motivo registrado; um dado inválido ou uma ação proibida não é motivo para `xfail`.

| ID | Cenário | Resultado esperado |
| --- | --- | --- |
| CAT-046 | Owner usa IDs de outro restaurante, ou IDs de ingredientes, grupos, opções, componentes e promoções sob um item pai incorreto. | `404`; as relações originais e os preços permanecem intactos. |
| CAT-047 | Owner, Administrator, Analyst, Manager, Employee e visitante consultam itens. Depois um Analyst perde sua associação usando o mesmo token. | Custo é omitido no detalhe e na lista sem permissão; revogação tem efeito na próxima consulta; `null` limpa o custo e zero explícito permanece zero para liderança. |
| CAT-048 | QR de filial não verificada ou inativa, enquanto outra filial da organização está publicada. | Leituras públicas daquela filial retornam `404`, inclusive sub-recursos; a filial publicada continua acessível e a equipe autorizada pode visualizar sua filial. |
| CAT-049 | Exclusão forçada em duas categorias, incluindo itens ativos e arquivados. | Sem `force`, `409` não altera a categoria nem os itens. Com `force`, ambos são movidos para uma única `Uncategorized`; itens arquivados continuam ocultos; a categoria de destino não pode ser excluída. |
| CAT-050 | Grupos opcionais, escolha única obrigatória e exatamente três acompanhamentos; seleção repetida ou de outro item. | Contagens são independentes por grupo; itens fora do grupo, duplicatas e contagens inválidas são recusados; preços extras só incidem nas opções selecionadas. |
| CAT-051 | Uma opção desabilitada ainda deixa opções suficientes, depois outra desabilitação torna o mínimo impossível. | A opção desabilitada não pode ser escolhida; o item segue comprável até o mínimo ficar impossível, apenas na filial afetada. |
| CAT-052 | Combo sem componentes, componente repetido, componente que é combo ou arquivado; remoção de ingrediente de uma unidade específica. | Combo vazio ou sem componente disponível não é comprável. Duplicatas e combos aninhados são recusados. Remoções válidas viram observações sem alterar o preço; unidades inexistentes e ingredientes não removíveis são recusados. |
| CAT-053 | Promoções de organização e filial se sobrepõem; uma promoção nova começa depois; gerente tenta alterar promoção de outra filial. | A filial ganha da organização; dentro do mesmo escopo vence a promoção nova somente durante sua janela `[início, fim)`. Acesso indevido não altera preço nem agenda. |
| CAT-054 | Outra edição ocorre imediatamente após o snapshot, incluindo mudança de nome e remoção de opção. | O snapshot anterior permanece igual; o novo reflete as mudanças. Valores que extrapolam o limite monetário são recusados em vez de gerar totais incorretos. |
| CAT-055 | Duas requisições tentam criar simultaneamente a mesma associação de ingrediente no item. | Exatamente uma retorna `201` e a outra `409`, com uma única relação gravada. |

---

## Módulo 4: Pedidos Anônimos, Especialização e Itens (ORD-001 a ORD-040)

As regressões do domínio `order` exercitam o HTTP e o PostgreSQL. `POST /restaurants/{slug}/branches/{branch_id}/orders/local` ou `/orders/remote` abre um rascunho quando não há itens e submete imediatamente quando há itens. No rascunho, `submitted_at` é nulo; depois da submissão, o pedido entra na fila operacional. Valores monetários são centavos inteiros. **Planejado** identifica integração ainda indisponível; não representa um teste executado ou um `xfail`.

| ID | Cenário | Resultado esperado |
| --- | --- | --- |
| ORD-001 | Abrir salão na mesa existente da filial pelo seu número. | `201`, subtipo local e mesa persistidos, sem exigir CPF. |
| ORD-002 | Abrir entrega como visitante, com endereço e itens válidos. | `201`, subtipo remoto, frete configurado e credencial de acesso. |
| ORD-003 | Enviar endereço no pedido local ou mesa no pedido remoto. | `400`, sem criar pedido parcial. |
| ORD-004 | Omitir a mesa ao abrir salão. | `201`, rascunho com `table_id` e `table_number` nulos. |
| ORD-005 | Alterar nome e preço do catálogo depois da submissão. | Nome, preço e total do pedido submetido não mudam. |
| ORD-006 | Adicionar durante uma promoção e submeter com ela ativa. | Preço bruto unitário preservado no snapshot; desconto refletido no total líquido. |
| ORD-007 | Promoção termina entre adição e submissão. | Checkout revalida e registra o preço integral vigente. |
| ORD-008 | Enviar quantidade zero, negativa ou preço definido pelo cliente. | `400`, sem alterar linhas ou total. |
| ORD-009 | Comprar várias linhas com opcionais, observações e remoções válidas. | Valores e instruções são preservados nas linhas submetidas. |
| ORD-010 | Item fica indisponível na filial depois de entrar no rascunho. | `409`; sem consumo de cupom, snapshot definitivo ou conversão. O rascunho continua editável. |
| ORD-011 | Avaliar itens de compras entregues e o restaurante, incluindo visitante com token do pedido. | Uma avaliação por linha comprada; para restaurante, uma por conta e organização ou uma por pedido de visitante. Comentário exige conta. |
| ORD-012 | Enviar nota fora de 1 a 3 para item ou de 1 a 5 para restaurante. | `400` sem alterar avaliações ou totais. |
| ORD-013 | Submeter múltiplas unidades de um combo com uma remoção de ingrediente e um item avulso; depois editar o catálogo. | A API retorna linhas filhas com quantidade, instruções e `parent_order_item_id`. O preço e o total das linhas filhas são zero; o item avulso mantém seu preço e o pedido não muda após editar o catálogo. |
| ORD-014 | Submeter com cupom elegível. | Desconto entra no total e em `order_coupon` uma única vez. |
| ORD-015 | Submeter cupom exclusivo de outro restaurante. | `404`, sem conversão, consumo ou perda do rascunho. |
| ORD-016 | Subtotal abaixo de `min_order_cents`. | Submissão mantém preço integral e não consome o cupom. |
| ORD-017 | Equipe avança pedido submetido de `pending` para `preparing`. | `200`; pular estados ou avançar rascunho recebe `409`. |
| ORD-018 | Equipe avança de `preparing` para `ready`. | `200`; fila passa a filtrar o pedido como pronto. |
| ORD-019 | Equipe avança de `ready` para `delivered`. | `200`; o pedido fica no estado final. |
| ORD-020 | Equipe cancela um pedido operacional pendente. | `200`; nova cobrança é bloqueada pelo domínio `transaction`. |
| ORD-021 | Tentar cancelar pedido já entregue. | `409`, sem mudar o estado. |
| ORD-022 | Usar mesa de outra filial, inativa ou identificadores conflitantes. | Mesa alheia ou inativa: `404`; campos conflitantes: `400`. |
| ORD-023 | Usar QR de mesa e QR sem mesa. | Ambos atribuem `qr_id`; só o primeiro preenche a mesa. |
| ORD-024 | Solicitar entrega com telefone ou endereço inválido. | `400` sem pedido persistido. |
| ORD-025 | Registrar instruções de entrega no pedido. **Planejado.** | Instruções persistem no subtipo remoto. |
| ORD-026 | Publicar estimativa de entrega. **Planejado.** | Estimativa fica disponível ao cliente autorizado. |
| ORD-027 | Alterar o frete sem permissão, definir frete negativo ou falsificar o frete no checkout. | Ações rejeitadas; só o valor configurado pelo Owner ou Administrator é cobrado. |
| ORD-028 | Adicionar item a rascunho autorizado. | `201`, totais atualizados e evento `cart_item_add` com item e quantidade. |
| ORD-029 | Adicionar ou remover uma linha após submissão. | `409`, sem alterar o snapshot. |
| ORD-030 | Remover linha do próprio rascunho; tentar remover linha de outro pedido. | Remoção válida fica auditada e recalcula total; linha alheia retorna `404`. |
| ORD-031 | Consultar ou alterar pedido existente com conta ou token alheio. | `403`; conta proprietária, equipe da filial ou token correto podem consultar. |
| ORD-032 | Consultar ID inexistente. | `404`. |
| ORD-033 | Listar fila por filial e status, com rascunhos e pedidos de outra filial no banco. | Apenas pedidos submetidos daquela filial e do filtro; cliente não acessa a fila. |
| ORD-034 | Tentar enviar ou pagar rascunho vazio. | `409`; abertura do rascunho em si é permitida. |
| ORD-035 | Adicionar itens em paralelo e, depois, submeter o mesmo pedido duas vezes em paralelo. | Nenhuma linha perdida; apenas uma submissão, uma conversão e um consumo de cupom. |
| ORD-036 | Abrir pedido em filial inexistente, de outro slug ou não publicada. | `404` sem criar pedido ou sessão. |
| ORD-037 | Enviar status desconhecido ou editar status como cliente. | `400` para valor desconhecido; `403` para cliente; estado permanece intacto. |
| ORD-038 | Abrir pedido de visitante e inspecionar resposta e dados persistidos. | Sem associação de conta ou CPF; token não aparece no JSON; banco guarda apenas seu hash. |
| ORD-039 | Repetir o GET do QR e depois abrir duas vezes via POST. | GET não cria pedidos; POSTs iniciam rascunhos distintos com mesma origem e credenciais diferentes. |
| ORD-040 | Abrir, editar, submeter e pagar via QR de mesa. | Eventos do próprio pedido preservam QR e mesa, registram inclusão e remoção e geram uma conversão; pagamento só é aceito após submissão. |

---

## Módulo 5: Pagamentos, PIX, Cupons e Fidelidade (FIN-001 a FIN-040)

As regressões executáveis usam os contratos de `sales` e `transaction`, valores inteiros em centavos e PostgreSQL para verificar persistência e concorrência. A cobrança usa um provedor simulado. Casos marcados como **planejados** representam integrações ainda não disponíveis e não são contabilizados como testes Go executados.

### FIN-001: Dinheiro exige equipe autorizada
- **Passos:** Enviar `{"payment_method":"cash","amount_cents":5000}` a `POST /orders/100/transactions` com `Idempotency-Key`, primeiro como cliente e depois como membro autorizado da equipe.
- **Esperado:** Cliente recebe `403`; equipe recebe `201` com status `paid`.

### FIN-002: Cartão de crédito e débito simulados
- **Passos:** Pagar pedidos usando `credit_card` e `debit_card`, sem enviar número de cartão.
- **Esperado:** Sucesso do mock produz `paid` e chama o provedor uma vez por cobrança.

### FIN-003: Pagamento Dividido da Comanda (Split Payment)
- **Pré-condição:** Pedido de 1000 centavos.
- **Passos:** Registrar 400 centavos em `cash` pela equipe e 600 centavos em `debit_card`; tentar cobrar mais 1 centavo.
- **Esperado:** As duas cobranças liquidadas totalizam 1000 centavos. A cobrança adicional recebe `409`.

### FIN-004: Criação de cobrança PIX simulada
- **Passos:** Iniciar transação `pix` e consultar `GET /transactions/{id}/pix`.
- **Esperado:** Transação `pending`, cobrança única com `copy_paste_code` simulado, QR em `qr_code_base64`, `expires_at` e `paid_at` nulo. O provedor ainda não foi chamado.

### FIN-005: Confirmação explícita do PIX
- **Passos:** Chamar `POST /transactions/{id}/pix/confirm` duas vezes antes do vencimento.
- **Esperado:** Status `paid` e `paid_at` gravado. Repetir a confirmação não chama o provedor novamente.

### FIN-006: Expiração de PIX Não Pago
- **Pré-condição:** Transação PIX que ultrapassou o `expires_at`.
- **Passos:** Tentar confirmar no vencimento e iniciar outra cobrança com chave nova.
- **Esperado:** Confirmação retorna `409`; cobrança expirada não reserva mais saldo nem chama o provedor.

### FIN-007: Emissão de recibo estruturado — planejado
- **Esperado:** Pedido pago tem um recibo associado por `order_id`, mesmo com múltiplas transações.

### FIN-008: Unicidade do Código de Recibo — planejado
- **Pré-condição:** Código `REC-123456` já emitido.
- **Passos:** Tentar gerar outro recibo com mesmo código.
- **Esperado:** HTTP 409 Conflict.

### FIN-009: Criação de Cupom com Restrição de Unicidade Composta `(restaurant_id, code)`
- **Pré-condição:** Restaurante ID 1.
- **Passos:** Owner ou Administrator envia `POST /restaurants/{slug}/coupons` com código, tipo e valor.
- **Esperado:** HTTP 201 Created. Registro inserido na tabela `coupon`.

### FIN-010: Mesmo Código de Cupom Permitido em Restaurantes Distintos
- **Pré-condição:** Cupom `PRATELLO10` existe no Restaurante ID 1.
- **Passos:** Cadastrar cupom `PRATELLO10` no Restaurante ID 2.
- **Esperado:** HTTP 201 Created devido ao índice composto `idx_coupons_restaurant_id_code`.

### FIN-011: Bloqueio de Cupom Duplicado no Mesmo Restaurante
- **Pré-condição:** Cupom `PRATELLO10` já existente no Restaurante 1.
- **Passos:** Tentar recriar no Restaurante 1 usando outra capitalização.
- **Esperado:** `409`; o código é único por restaurante sem distinguir maiúsculas e minúsculas.

### FIN-012: Limites de descontos (`percent` e `fixed_cents`)
- **Esperado:** Percentual acima de 100 e valores não positivos são inválidos. Desconto fixo nunca reduz o total abaixo de zero.

### FIN-013: Acúmulo na primeira compra por CPF
- **Esperado:** O PIX pendente não credita moedas. A confirmação do pagamento integral credita uma vez a carteira do CPF no restaurante. O visitante consulta o saldo sem login.

### FIN-014: Acúmulo em carteira existente
- **Esperado:** Uma nova compra paga credita a carteira existente. Reenviar o mesmo pagamento não duplica o crédito.

### FIN-015: Fidelidade: Bloqueio de Resgate de Moedas para Usuário sem Conta (`account_id IS NULL`)
- **Passos:** Enviar `coins > 0` na cotação de um carrinho sem autenticação.
- **Esperado:** `403`; não há resgate anônimo.

### FIN-016: Carteira acumulada por CPF após registro
- **Esperado:** O cadastro conserva a mesma conta e carteira do CPF. A consulta autenticada retorna o saldo acumulado antes do cadastro.

### FIN-017: Fidelidade: Resgate de Moedas com Sucesso para Usuário com Conta (`account_id IS NOT NULL`)
- **Passos:** Cliente com 50 moedas resgata 20 após os descontos de cupom.
- **Esperado:** 20 centavos a menos. A cotação não consome moedas; `ApplyToOrder` reduz o saldo para 30 e grava o lançamento da carteira.

### FIN-018: Fidelidade: Bloqueio de Resgate de Moedas Superior ao Saldo
- **Pré-condição:** Usuário com conta possui 10 moedas.
- **Passos:** Tentar resgatar 25 moedas.
- **Esperado:** `409`; saldo, cupons e pedido permanecem intactos.

### FIN-019: Fidelidade: Isolamento de Moedas Entre Restaurantes Distintos
- **Pré-condição:** Cliente tem 50 moedas no Restaurante 1 e 0 moedas no Restaurante 2.
- **Passos:** Tentar resgatar moedas no Restaurante 2.
- **Esperado:** Rejeitado. Moedas são intransferíveis entre diferentes `restaurant_id`.

### FIN-020: Pedido cancelado não pode ser cobrado
- **Esperado:** Pedido `cancelled` rejeita nova cobrança e confirmação de PIX pendente com `409`, antes de chamar o provedor.

### FIN-021: Reversão de crédito de fidelidade — planejado
- **Esperado:** Cancelamento de compra com crédito de fidelidade reverte o crédito uma vez, com trilha de auditoria.

### FIN-022: Falha injetada no provedor
- **Passos:** Substituir o mock padrão por um provedor de teste que retorna erro.
- **Esperado:** Não persistir uma cobrança quitada nem reservar saldo; nova tentativa válida continua possível.

### FIN-023: Validação de Chave Estrangeira: Transação com `order_id` Inexistente
- **Pré-condição:** `order_id = 999999` não cadastrado.
- **Passos:** POST transação.
- **Esperado:** `404`; não há transação ou chamada ao provedor.

### FIN-024: Limite Máximo de Usos de Cupom (`coupon.max_uses`)
- **Pré-condição:** Cupom com `max_uses = 3`, já utilizado 3 vezes.
- **Passos:** Quarto pedido tenta usar o cupom.
- **Esperado:** A validação apresenta `MAX_USES_REACHED`; o checkout ignora o cupom e não incrementa o contador. Dois pedidos concorrentes pelo último uso concedem o desconto apenas uma vez.

### FIN-025: Cupom Fora da Validade (`coupon.expires_at`)
- **Pré-condição:** Cupom com data de expiração no passado.
- **Passos:** Aplicar cupom no pedido.
- **Esperado:** A validação apresenta `EXPIRED` a partir do instante de expiração; o checkout não dá desconto.

### FIN-026: Desativação de Cupom (`is_active = false`)
- **Pré-condição:** Cupom desativado manualmente.
- **Passos:** Tentar utilizar o cupom.
- **Esperado:** A validação apresenta `INACTIVE`; o cupom não aparece como banner ou cupom utilizável e não dá desconto.

### FIN-027: Validação de valor mínimo (`min_order_cents`)
- **Pré-condição:** Cupom exige mínimo de 60.00; Pedido com total de 40.00.
- **Passos:** Tentar aplicar.
- **Esperado:** Sem desconto abaixo do mínimo (`MIN_ORDER_NOT_MET`); o subtotal igual ao limite é elegível.

### FIN-028: Cálculo e auditoria de recibo em JSONB — planejado
- **Pré-condição:** Recibo gravado para pedido finalizado.
- **Passos:** Consultar coluna `receipt.receipt`.
- **Esperado:** Objeto JSON íntegro contendo discriminação de itens, taxas e formas de pagamento.

### FIN-029: Prevenção de Transação com Valor Negativo ou Zero
- **Pré-condição:** Nenhuma.
- **Passos:** Criar transação com `amount_cents` igual a zero ou negativo.
- **Esperado:** `400` antes de chamar o provedor.

### FIN-030: Concorrência no Resgate de Moedas
- **Pré-condição:** Usuário com conta possui 30 moedas.
- **Passos:** Dois pedidos simultâneos solicitam o resgate de 20 moedas cada.
- **Esperado:** Apenas um resgate conclui com sucesso; o segundo falha por saldo insuficiente (prevenção de double-spending).

### Regressões de negócio executáveis (FIN-031 a FIN-040)

| ID | Cenário | Resultado esperado |
| --- | --- | --- |
| FIN-031 | Manager tenta criar cupom; Owner referencia produto de outra organização ou declara combo para produto simples. | `403` para Manager; alvo inválido é rejeitado e nenhum cupom é criado. |
| FIN-032 | Banner consultado em filial sem publicação; visitante toca o banner; cliente autenticado repete o toque. | Filial não publicada recebe `404`. Visitante recebe o código sem criar vínculo; conta autenticada ganha um único vínculo e outra conta não o herda. |
| FIN-033 | Cupom específico concorre com promoção do item e cupom de pedido; outro item não tem desconto próprio. | Maior desconto vence por linha, sem acumular; cupom geral ainda pode descontar a outra linha. |
| FIN-034 | Cupom de combo, preço de opcionais e dois cupons gerais candidatos. | Cupons só descontam os alvos corretos. Percentual do cupom inclui opcionais; promoção do item afeta apenas a base. Apenas o cupom geral mais forte concorre. |
| FIN-035 | Cotação repetida, finalização repetida, código e ID salvo do mesmo cupom. | Cotação não altera contadores ou carteira. Finalização concede um uso, grava snapshot e débito uma vez e rejeita repetição. |
| FIN-036 | Moedas excedem o valor devido depois dos descontos. | Rejeição antes de debitar a carteira; o total nunca fica negativo. |
| FIN-037 | Cobranças concorrentes com mesma chave e com chaves distintas; mesma chave com outro valor. | Mesmo pedido e mesma chave retornam a cobrança original; payload alterado ou tentativa que excede saldo recebe `409`. |
| FIN-038 | Conta alheia consulta transação ou PIX; cliente tenta reembolso do próprio pagamento. | `403`; nenhum movimento no provedor nem mudança no saldo. |
| FIN-039 | Reembolso parcial repetido, tentativa acima do restante e reembolso integral final. | Uma devolução por chave; excesso recebe `409`; devolver o restante muda o status para `refunded`. |
| FIN-040 | PIX pendente tenta ser confirmado exatamente no vencimento; pagamento alternativo compete antes e depois. | PIX reserva saldo até o vencimento exclusivo; confirmação no limite é recusada e nova cobrança pode usar o valor liberado. |

---

## Módulo 6: Avaliações de Itens, Combos e Restaurantes (REV-001 a REV-014)

Os subtestes em `server/internal/review/reviews_caso_de_testes_test.go` usam o HTTP real com PostgreSQL. As rejeições esperadas são testes normais: verificam o status e, quando há tentativa de escrita, que o estado não foi alterado.

| ID | Cenário | Resultado esperado |
| --- | --- | --- |
| REV-001 | Enviar nota ausente, zero, acima de 3 para item ou acima de 5 para restaurante. | `400`, sem inserir avaliações; os extremos válidos das duas escalas são aceitos. |
| REV-002 | Escrever sem credencial, com conta de outro cliente, com token de outro pedido ou com linha de um pedido diferente. | Sem credencial: `401`; conta ou token alheio: `403`; linha fora do pedido: `404`. |
| REV-003 | Avaliar pedido pendente, cancelado ou não submetido, linha anulada, item com avaliações desabilitadas ou componente sem compra independente. | `409`, sem registrar avaliação. |
| REV-004 | Visitante com token válido tenta enviar comentário, depois envia somente a nota. | Comentário recebe `403`; nota é aceita e publicada sem comentário. |
| REV-005 | Cliente avalia duas compras do mesmo item e repete a avaliação da primeira compra. | Uma avaliação por linha comprada; nova compra aceita, repetição recebe `409`. |
| REV-006 | Duas requisições concorrentes avaliam a mesma linha de visitante. | Uma aceita, uma recebe `409`; o total aumenta somente uma vez. |
| REV-007 | Cliente avalia um combo e um prato comprado separadamente. | Cada produto tem contagem, soma e média próprias; a nota do combo não se propaga ao prato. |
| REV-008 | Cliente avalia a organização após compras em duas filiais. | A segunda avaliação atualiza a mesma nota e o mesmo comentário; o total da organização não cresce. |
| REV-009 | Visitante avalia a organização após a entrega e repete a tentativa no mesmo pedido. | Só a primeira nota é aceita; comentário exige conta e repetição não muda os totais. |
| REV-010 | Consultar avaliações públicas sem login, com paginação. | A página contém só as avaliações solicitadas, com o resumo completo e sem IDs privados ou credenciais. |
| REV-011 | Consultar item sob o slug errado, arquivado ou de restaurante sem filial publicada. | `404`, sem expor avaliações nem volume de vendas. |
| REV-012 | Consultar unidades vendidas com compras submetidas, canceladas, rascunhos, linhas anuladas e componentes de combo. | Soma somente quantidades submetidas, não canceladas, não anuladas e de linhas compradas independentemente. |
| REV-013 | Atualizar e remover uma avaliação na persistência, inclusive a última. | Contagem e soma continuam exatas; sem avaliações, a média é nula. |
| REV-014 | Enviar comentário com caracteres multibyte no limite e acima de 2000 caracteres. | Limite aceito; excesso recebe `400` sem alterar a avaliação existente. |

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
