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
