package org.pratello;

import com.microsoft.playwright.*;
import org.junit.jupiter.api.*;

import static com.microsoft.playwright.assertions.PlaywrightAssertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertTrue;

@TestInstance(TestInstance.Lifecycle.PER_CLASS)
public class PratelloE2ETest {
    private Playwright playwright;
    private Browser browser;
    private String baseUrl;

    @BeforeAll
    void init() {
        playwright = Playwright.create();
        // Headless by default; can run in headed mode if HEADLESS=false
        boolean headless = !"false".equalsIgnoreCase(System.getenv("HEADLESS"));
        browser = playwright.chromium().launch(new BrowserType.LaunchOptions().setHeadless(headless));

        baseUrl = System.getenv("BASE_URL");
        if (baseUrl == null || baseUrl.isBlank()) {
            baseUrl = "http://localhost:5173";
        }
    }

    @AfterAll
    void tearDown() {
        if (browser != null) {
            browser.close();
        }
        if (playwright != null) {
            playwright.close();
        }
    }

    @Test
    @DisplayName("Homepage renders Pratello branding and navigation")
    void testHomepageRenders() {
        BrowserContext context = browser.newContext();
        Page page = context.newPage();

        page.navigate(baseUrl);

        // Check main heading
        Locator heading = page.locator("h1");
        assertThat(heading).containsText("Discover Places on Pratello");

        // Check that quick links are present
        assertThat(page.getByText("Bob's Pizzeria").first()).isVisible();
        assertThat(page.getByText("Delícia de Prazeres").first()).isVisible();

        context.close();
    }

    @Test
    @DisplayName("Navigate to Bob's Pizzeria and verify menu items from API")
    void testNavigateToBobPizzeria() {
        BrowserContext context = browser.newContext();
        Page page = context.newPage();

        page.navigate(baseUrl);

        // Click Bob's Pizzeria card
        page.getByText("Bob's Pizzeria").first().click();

        // Verify URL route
        assertTrue(page.url().contains("/bobpizzeria"), "URL should contain /bobpizzeria");

        // Verify restaurant details rendered
        assertThat(page.locator("h1")).containsText("Bob's Pizzeria");
        assertThat(page.getByText("Margherita DOP")).isVisible();

        // Click Back to Home
        page.getByText("Back to Pratello Home").click();
        assertThat(page.locator("h1")).containsText("Discover Places on Pratello");

        context.close();
    }

    @Test
    @DisplayName("Custom restaurant input navigates correctly")
    void testCustomRestaurantSearch() {
        BrowserContext context = browser.newContext();
        Page page = context.newPage();

        page.navigate(baseUrl);

        // Type in custom restaurant slug
        Locator input = page.locator("input[placeholder*='bobpizzeria']");
        input.fill("taco-fiesta");
        page.getByRole(AriaRole.BUTTON, new Page.GetByRoleOptions().setName("Visit")).click();

        // Verify URL
        assertTrue(page.url().contains("/taco-fiesta"), "URL should contain /taco-fiesta");
        assertThat(page.locator("h1")).containsText("Taco-Fiesta");

        context.close();
    }
}
