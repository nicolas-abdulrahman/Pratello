const { test, expect } = require('@playwright/test');

const storefront = '/r/loja-teste';
const shop = { name: 'Loja teste', slug: 'loja-teste' };
const dish = { id: 1, name: 'Pizza teste', description: 'Pizza artesanal', category_id: 1, base_price_cents: 1200, effective_price_cents: 1000, is_available: true };

test.beforeEach(async ({ page }) => {
  await page.route('https://viacep.com.br/ws/50000000/json/', route => route.fulfill({ json: { uf: 'PE', localidade: 'Recife' } }));
  await page.route('**/api/**', route => {
    const { pathname: path, searchParams } = new URL(route.request().url());
    const restaurant = '/api/restaurants/loja-teste';
    if (path === restaurant) return route.fulfill({ json: shop });
    if (path === '/api/restaurants') return route.fulfill({ json: { data: [shop], total: 1 } });
    if (path === `${restaurant}/items`) {
      const data = searchParams.get('category_id') === '2'
        ? Array.from({ length: 6 }, (_, index) => ({ ...dish, id: index + 2, category_id: 2, name: `Combo ${index + 1}`, is_combo: true }))
        : [dish];
      return route.fulfill({ json: { data, total: data.length } });
    }
    if (path === `${restaurant}/categories`) return route.fulfill({ json: [{ id: 1, name: 'Pizzas' }, { id: 2, name: 'Combos' }] });
    if (path.endsWith('/reviews')) return route.fulfill({ json: { data: [], summary: { average_rating: null, review_count: 0 } } });
    if (path.endsWith('/municipios')) return route.fulfill({ json: [{ id: 1, nome: 'Recife' }] });
    if (path.endsWith('/estados')) return route.fulfill({ json: [{ sigla: 'PE', nome: 'Pernambuco' }] });
    if (path === '/api/auth/login') return route.fulfill({ json: { access_token: 'fixture-token', refresh_token: 'fixture-refresh', expires_in: 3600 } });
    if (path === '/api/auth/me') return route.fulfill({ json: { id: 1, name: 'Ana Silva', email: 'ana@example.com', roles: [] } });
    if (path === `${restaurant}/menu`) return route.fulfill({ json: {
      display_config: {
        dark: false,
        colors: { bg: '#eeeeee', bg_dark: '#222222' },
        first_page: [{ id: 'hero', type: 'header1', title: { text: 'Loja teste' } },
          { id: 'combos', type: 'section', display_items: { by: 'category', value: 'combos' }, container: 'horizontal', title: { text: 'Combos' } }],
      },
    } });
    if (path === `${restaurant}/branches`) return route.fulfill({ json: { data: [] } });
    if (path.startsWith(`${restaurant}/`)) return route.fulfill({ json: [] });
    return route.fulfill({ status: 404, json: {} });
  });
  await page.goto(storefront);
});

const cartButton = page => page.getByRole('button', { name: 'Carrinho', exact: true });
const badge = page => cartButton(page).locator('span').filter({ hasText: /^\d+$/ });
const cartBackdrop = page => page.getByRole('button', { name: 'Fechar carrinho', exact: true });
const cartHeading = page => page.getByRole('heading', { name: 'Seu Pedido' });
const button = (page, name) => page.getByRole('button', { name, exact: typeof name === 'string' });
const add = page => button(page, 'Adicionar ao carrinho').click();
const openCart = page => cartButton(page).click();
const total = (page, amount) => page.getByText(new RegExp(`^R\\$\\s*${amount},00$`)).last();
const order = async page => { await add(page); await openCart(page); await button(page, 'No Restaurante').click(); };
const detail = async page => { await page.getByText(dish.name, { exact: true }).click(); return page.getByRole('dialog'); };
const loginMenu = page => button(page, 'Entrar').click();
const email = page => page.getByLabel('Endereço de Email');
const password = page => page.getByLabel('Senha', { exact: true });
const submit = page => page.locator('form button[type=submit]');
const signIn = async page => {
  await page.goto('/login');
  await email(page).fill('ana@example.com');
  await password(page).fill('fixture-password');
  const request = page.waitForRequest('**/api/auth/login');
  await submit(page).click();
  expect((await request).postDataJSON()).toEqual({ email: 'ana@example.com', password: 'fixture-password' });
  await expect(button(page, /Olá Ana/)).toBeVisible();
};

test('WEB-006 adds the promotion directly and updates the badge', async ({ page }) => {
  await add(page);
  await expect(badge(page)).toHaveText('1');
  await openCart(page);
  await expect(page.getByText(dish.name, { exact: true }).last()).toBeInViewport();
});

for (const [id, amount] of [['WEB-007', 20], ['WEB-008', 10], ['WEB-010', 30]]) {
  test(`${id} recalculates the order total to ${amount}`, async ({ page }) => {
    await order(page);
    await button(page, 'Aumentar quantidade').click();
    if (amount === 10) await button(page, 'Diminuir quantidade').click();
    if (amount === 30) await button(page, 'Aumentar quantidade').click();
    await expect(total(page, amount)).toBeInViewport();
    await expect(page.locator('.qty-value')).toHaveText(String(amount / 10));
  });
}

test('WEB-009 removes an item with the drawer remove button', async ({ page }) => {
  await add(page);
  await openCart(page);
  await button(page, 'Remover item').click();
  await expect(page.getByText('Seu carrinho está vazio')).toBeInViewport();
  await expect(total(page, 0)).toBeInViewport();
});

test('WEB-011 opens the dine-in order', async ({ page }) => {
  await order(page);
  await expect(page).toHaveURL('/order');
  await expect(button(page, 'Ir para o Pagamento')).toBeVisible();
});

test('WEB-012 opens delivery from the drawer', async ({ page }) => {
  await add(page);
  await openCart(page);
  await button(page, 'Deliver').click();
  await expect(page).toHaveURL(`${storefront}/delivery`);
  await expect(page.getByRole('heading', { name: 'Área de Delivery' })).toBeVisible();
});

test('WEB-013 opens checkout from the order', async ({ page }) => {
  await order(page);
  await button(page, 'Ir para o Pagamento').click();
  await expect(page).toHaveURL('/checkout');
  await expect(page.getByRole('tablist', { name: 'Método de pagamento' })).toBeVisible();
});

test('WEB-014 retains the cart during navbar navigation', async ({ page }) => {
  await add(page);
  await page.getByRole('navigation').getByRole('link', { name: 'Cardápio' }).click();
  await expect(page).toHaveURL(`${storefront}/cardapio`);
  await expect(badge(page)).toHaveText('1');
  await openCart(page);
  await expect(page.getByText(dish.name, { exact: true })).toBeInViewport();
});

test('WEB-020 opens location controls', async ({ page }) => {
  await button(page, 'Escolha um endereço').click();
  await expect(page.getByLabel('Estado', { exact: true })).toBeVisible();
  await expect(page.getByLabel('Buscar por CEP')).toBeVisible();
});

test('WEB-021 selects a city', async ({ page }) => {
  await button(page, 'Escolha um endereço').click();
  await page.getByLabel('Estado', { exact: true }).selectOption('PE');
  await page.getByLabel('Cidade', { exact: true }).selectOption('Recife');
  await expect(page.getByRole('button', { name: /Recife/i })).toBeVisible();
});

test('WEB-023 dismisses the location controls outside', async ({ page }) => {
  await button(page, 'Escolha um endereço').click();
  await expect(page.getByLabel('Buscar por CEP')).toBeVisible();
  await page.getByRole('banner').click({ position: { x: 900, y: 100 } });
  await expect(page.getByLabel('Buscar por CEP')).toBeHidden();
});

test('WEB-022 resolves a postal code and closes the picker', async ({ page }) => {
  await button(page, 'Escolha um endereço').click();
  await page.getByLabel('Buscar por CEP').fill('50000000');
  await button(page.getByRole('dialog', { name: 'Escolher endereço' }), 'Buscar').click();
  await expect(button(page, 'Recife, PE')).toBeVisible();
  await expect(page.getByLabel('Buscar por CEP')).toBeHidden();
});

test('WEB-024 opens the access menu', async ({ page }) => {
  await loginMenu(page);
  await expect(button(page, /Cliente/)).toBeInViewport();
  await expect(button(page, /Funcionário/)).toBeInViewport();
});

test('WEB-025 closes the access menu', async ({ page }) => {
  await loginMenu(page);
  await expect(button(page, /Cliente/)).toBeInViewport();
  await button(page.locator('aside').filter({ has: button(page, /Cliente/) }), 'Fechar').click();
  await expect(button(page, /Cliente/)).not.toBeInViewport();
});

for (const [id, role] of [['WEB-026', 'Cliente'], ['WEB-027', 'Funcionário']]) {
  test(`${id} opens login as ${role}`, async ({ page }) => {
    await loginMenu(page);
    await button(page, new RegExp(role)).click();
    await expect(page).toHaveURL('/login');
    await expect(email(page)).toBeVisible();
  });
}

test.describe('login form', () => {
  test.beforeEach(async ({ page }) => { await page.goto('/login'); });
  test('WEB-028 enters an email address', async ({ page }) => {
    await email(page).fill('ana@example.com');
    await expect(email(page)).toBeFocused();
    await expect(email(page)).toHaveValue('ana@example.com');
  });
  test('WEB-029 rejects malformed email', async ({ page }) => {
    await email(page).fill('ana');
    await password(page).fill('fixture-password');
    await submit(page).click();
    await expect(email(page)).toBeFocused();
    expect(await email(page).evaluate(input => input.validity.typeMismatch)).toBe(true);
  });
  test('WEB-030 enters a masked password', async ({ page }) => {
    await password(page).fill('fixture-password');
    await expect(password(page)).toHaveValue('fixture-password');
    await expect(password(page)).toHaveAttribute('type', 'password');
  });
  test('WEB-031 reveals and masks the password', async ({ page }) => {
    await page.getByTitle('Exibir senha').click();
    await expect(password(page)).toHaveAttribute('type', 'text');
    await page.getByTitle('Ocultar senha').click();
    await expect(password(page)).toHaveAttribute('type', 'password');
  });
  test('WEB-032 toggles remember me', async ({ page }) => {
    const remember = page.getByRole('checkbox', { name: 'Lembrar de mim' });
    await remember.uncheck();
    await expect(remember).not.toBeChecked();
    await remember.check();
    await expect(remember).toBeChecked();
  });
  test('WEB-033 displays password reset support instructions', async ({ page }) => {
    await button(page, 'Esqueceu a senha?').click();
    await expect(page.getByRole('alert')).toContainText('solicite suporte');
  });
  test('WEB-034 requires an email for submission', async ({ page }) => {
    await submit(page).click();
    await expect(email(page)).toBeFocused();
    expect(await email(page).evaluate(input => input.validity.valueMissing)).toBe(true);
  });
  test('WEB-036 switches to registration and back', async ({ page }) => {
    await button(page, 'Crie uma conta').click();
    await expect(page.getByLabel(/Nome completo/)).toBeVisible();
    await button(page, 'Faça login').click();
    await expect(page.getByLabel(/Nome completo/)).toBeHidden();
  });
});

test('WEB-035 authenticates through the login form', async ({ page }) => { await signIn(page); });
test('WEB-037 shows the authenticated name and initials', async ({ page }) => {
  await signIn(page);
  await expect(page.getByText('AS', { exact: true }).first()).toBeVisible();
});
test('WEB-038 logs out through the account menu', async ({ page }) => {
  await signIn(page);
  await button(page, /Olá Ana/).click();
  await button(page, 'Fazer logout').click();
  await expect(button(page, 'Entrar')).toBeVisible();
});

test('WEB-041 explores nearby restaurants from the marketplace hero', async ({ page }) => {
  await page.goto('/restaurantes');
  await page.getByRole('link', { name: 'Explorar restaurantes' }).click();
  await expect(page).toHaveURL('/restaurantes#restaurantes-proximos');
  await expect(page.getByPlaceholder('Buscar restaurantes ou pratos próximos...')).toBeInViewport();
});

test('WEB-043 opens the restaurant from its marketplace card', async ({ page }) => {
  await page.goto('/restaurantes');
  await page.getByRole('link', { name: 'Ver restaurante' }).click();
  await expect(page).toHaveURL(storefront);
  await expect(page.getByRole('heading', { name: shop.name })).toBeVisible();
});

test('WEB-042 scrolls the combos carousel in both directions', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const previous = button(page, 'Combos anteriores');
  await expect(previous).toBeHidden();
  await button(page, 'Próximos combos').click();
  await expect(previous).toBeVisible();
  await previous.click();
  await expect(previous).toBeHidden();
});

test('WEB-044 expands and collapses a menu category', async ({ page }) => {
  await page.goto(`${storefront}/cardapio`);
  const category = button(page, /Pizzas/);
  await category.click();
  await expect(category).toHaveAttribute('aria-expanded', 'true');
  await expect(page.getByText(dish.name, { exact: true })).toBeVisible();
  await category.click();
  await expect(category).toHaveAttribute('aria-expanded', 'false');
  await expect(page.getByText(dish.name, { exact: true })).toBeHidden();
});

test('WEB-045 opens item details inside the tenant theme', async ({ page }) => {
  const modal = await detail(page);
  await expect(modal).toBeVisible();
  await expect(page.locator('[data-theme="restaurantTheme"]').getByRole('dialog')).toHaveCount(1);
  await expect(modal.getByText(dish.description)).toBeVisible();
});

test('WEB-048 adds the displayed item from its dialog', async ({ page }) => {
  const modal = await detail(page);
  await button(modal, 'Adicionar ao Carrinho').click();
  await expect(badge(page)).toHaveText('1');
});

test('WEB-049 dismisses item details via Voltar and backdrop', async ({ page }) => {
  const modal = await detail(page);
  await button(modal, 'Voltar').click();
  await expect(modal).toBeHidden();
  await detail(page);
  await button(page, 'Fechar detalhes do prato').click({ position: { x: 5, y: 5 } });
  await expect(modal).toBeHidden();
});

test('WEB-050 selects checkout methods and acknowledges simulated PIX', async ({ page }) => {
  await order(page);
  await button(page, 'Ir para o Pagamento').click();
  const card = page.getByRole('tab', { name: 'Cartão de Crédito' });
  const pix = page.getByRole('tab', { name: 'PIX' });
  await card.click();
  await expect(card).toHaveAttribute('aria-selected', 'true');
  await expect(page.getByPlaceholder('0000 0000 0000 0000')).toBeVisible();
  await pix.click();
  await expect(pix).toHaveAttribute('aria-selected', 'true');
  const dialog = page.waitForEvent('dialog');
  page.once('dialog', alert => alert.accept());
  await button(page, 'Já realizei o pagamento').click();
  expect((await dialog).message()).toBe('Pagamento PIX confirmado!');
});

test('WEB-001 opens the cart', async ({ page }) => {
  await cartButton(page).click();
  await expect(cartHeading(page)).toBeInViewport();
});

test('WEB-002 closes the cart with its close button', async ({ page }) => {
  await cartButton(page).click();
  await page.getByRole('button', { name: 'Fechar', exact: true }).click();
  await expect(cartHeading(page)).not.toBeInViewport();
});

test('WEB-003 closes the cart with its backdrop', async ({ page }) => {
  await cartButton(page).click();
  await cartBackdrop(page).click({ position: { x: 20, y: 200 } });
  await expect(cartBackdrop(page)).toHaveCount(0);
  await expect(cartHeading(page)).not.toBeInViewport();
});

test('WEB-004 closes the cart with Escape', async ({ page }) => {
  await cartButton(page).click();
  await expect(cartHeading(page)).toBeInViewport();
  await page.keyboard.press('Escape');
  await expect(cartBackdrop(page)).toHaveCount(0);
  await expect(cartHeading(page)).not.toBeInViewport();
});

test('WEB-005 displays the empty cart and zero total', async ({ page }) => {
  await cartButton(page).click();
  await expect(page.getByText('Seu carrinho está vazio')).toBeInViewport();
  await expect(page.getByText('Adicione itens para continuar.')).toBeInViewport();
  await expect(page.getByText(/^R\$\s*0,00$/)).toBeInViewport();
});

for (const [id, name, suffix] of [['WEB-017', 'Cardápio', '/cardapio'], ['WEB-018', 'Delivery', '/delivery']]) {
  test(`${id} navigates through ${name}`, async ({ page }) => {
    await page.getByRole('navigation').getByRole('link', { name, exact: true }).click();
    await expect(page).toHaveURL(`${storefront}${suffix}`);
    await expect(page.locator('[data-theme="restaurantTheme"]')).toBeVisible();
  });
  const heroLabel = name === 'Cardápio' ? 'Ver cardápio' : name;
  test(`${id} navigates through the hero link ${heroLabel}`, async ({ page }) => {
    await page.getByRole('banner').getByRole('link', { name: heroLabel, exact: true }).click();
    await expect(page).toHaveURL(`${storefront}${suffix}`);
    await expect(page.locator('[data-theme="restaurantTheme"]')).toBeVisible();
  });
}

test('WEB-016 returns from the menu to the tenant homepage', async ({ page }) => {
  await page.goto(`${storefront}/cardapio`);
  await page.getByRole('link', { name: 'Início', exact: true }).click();
  await expect(page).toHaveURL(storefront);
});

test('WEB-051 toggles the storefront palette in both directions', async ({ page }) => {
  const toggle = page.getByRole('button', { name: 'Modo escuro' });
  const canvas = page.getByRole('heading', { name: 'Loja teste' }).locator('xpath=ancestor::div[contains(@class, "selection:bg-primary")]');
  await expect(canvas).toHaveCSS('background-color', 'rgb(238, 238, 238)');
  await toggle.click();
  await expect(toggle).toHaveAttribute('aria-pressed', 'true');
  await expect(canvas).toHaveCSS('background-color', 'rgb(34, 34, 34)');
  await toggle.click();
  await expect(toggle).toHaveAttribute('aria-pressed', 'false');
  await expect(canvas).toHaveCSS('background-color', 'rgb(238, 238, 238)');
});
