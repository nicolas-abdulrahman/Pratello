# Pratello API

Base URL: `http://localhost:8080/api`

## Endpoints

| Method | Route | Body | Description |
| :--- | :--- | :--- | :--- |
| **Autenticação & Contas** | | | |
| `POST` | `/auth/register` | `{ email, password, phone }` | Criar nova conta global (`account`) |
| `POST` | `/auth/login` | `{ email, password }` | Autenticar usuário e gerar sessão/token |
| `GET` | `/auth/me` | — | Obter dados da conta autenticada |
| `GET` | `/sysadmins` | — | Listar superadministradores (`sysadmin`) |
| `POST` | `/sysadmins` | `{ account_id }` | Promover conta a superadministrador |
| `GET` | `/restaurants/{slug}/managers` | — | Listar gerentes do restaurante (`manager`) |
| `POST` | `/restaurants/{slug}/managers` | `{ account_id }` | Vincular gerente ao restaurante |
| `DELETE` | `/restaurants/{slug}/managers/{account_id}` | — | Desvincular gerente do restaurante |
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
| `POST` | `/restaurants/{slug}/branches` | `{ name, address: { cep, street, number, apt, complement } }` | Criar filial física com endereço (`Address`) |
| `PUT` | `/restaurants/{slug}/branches/{branch_id}` | `{ name }` | Atualizar informações da filial |
| `DELETE` | `/restaurants/{slug}/branches/{branch_id}` | — | Desativar filial física |
| `PUT` | `/restaurants/{slug}/branches/{branch_id}/address` | `{ cep, street, number, apt, complement }` | Atualizar endereço da filial (`Address`) |
| `GET` | `/restaurants/{slug}/branches/{branch_id}/phones` | — | Listar telefones da filial (`p_restaurant_phone`) |
| `POST` | `/restaurants/{slug}/branches/{branch_id}/phones` | `{ phone_number, phone_type }` | Adicionar telefone à filial |
| `DELETE` | `/restaurants/{slug}/branches/{branch_id}/phones/{phone_id}` | — | Remover telefone da filial |
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
