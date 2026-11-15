package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
)

// App wires together Layer 1 (HTTP Transport / Handlers), Layer 2 (Business Services), and Layer 3 (Persistence Stores).
type App struct {
	RestaurantService *RestaurantService
	DishService       *DishService
	AuthService       *AuthService
	Restaurants       RestaurantStore
	Dishes            DishStore
	Accounts          AccountStore
	DB                *sql.DB
}

// NewApp initializes dependencies across all layers.
func NewApp(db *sql.DB) *App {
	app := &App{
		DB:          db,
		Restaurants: NewSQLRestaurantStore(db),
		Dishes:      NewSQLDishStore(db),
		Accounts:    NewSQLAccountStore(db),
	}
	app.RestaurantService = NewRestaurantService(app.Restaurants)
	app.DishService = NewDishService(app.Restaurants, app.Dishes)
	app.AuthService = NewAuthService(app.Accounts)
	return app
}

// RegisterRoutes registers all route endpoints for both "/" and "/api/" prefixes.
func (a *App) RegisterRoutes(mux *http.ServeMux) {
	prefixes := []string{"", "/api"}

	// Static File Server: serve public uploads from ./media/ directory
	_ = os.MkdirAll("media/dishes", 0755)
	mux.Handle("GET /media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))

	for _, p := range prefixes {
		// Autenticação & Contas
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/register", p), a.handleAuthRegister)
		mux.HandleFunc(fmt.Sprintf("POST %s/auth/login", p), a.handleAuthLogin)
		mux.HandleFunc(fmt.Sprintf("GET %s/auth/me", p), a.handleAuthMe)

		// Superadministradores Globais
		mux.HandleFunc(fmt.Sprintf("GET %s/sysadmins", p), a.handleListSysadmins)
		mux.HandleFunc(fmt.Sprintf("POST %s/sysadmins", p), a.handleCreateSysadmin)

		// Gerentes de Restaurante
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/managers", p), a.handleListManagers)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/managers", p), a.handleCreateManager)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/managers/{account_id}", p), a.handleDeleteManager)

		// Proprietários (Owner Role)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/owners", p), a.handleListOwners)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/owners", p), a.handlePromoteOwner)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/owners/{account_id}", p), a.handleRemoveOwner)

		// Filiais e Unidades Físicas (Physical Restaurants)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/branches", p), a.handleListBranches)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches", p), a.handleCreateBranch)

		// Quadro de Funcionários (HR Records)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/branches/{branch_id}/employees", p), a.handleListEmployees)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/employees", p), a.handleCreateEmployee)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/employees/{employee_id}/invite", p), a.handleInviteEmployee)

		// Terminal de Salão / Sessão de Turno (PIN de 6 dígitos)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/branches/{branch_id}/terminal/session", p), a.handleOpenTerminalSession)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/branches/{branch_id}/terminal/orders", p), a.handleTerminalOrders)

		// Restaurants
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants", p), a.handleListRestaurants)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants", p), a.handleCreateRestaurant)
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}", p), a.handleGetRestaurant)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}", p), a.handleUpdateRestaurant)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}", p), a.handleDeleteRestaurant)

		// Dishes
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/dishes", p), a.handleListDishes)
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes", p), a.handleCreateDish)
		mux.HandleFunc(fmt.Sprintf("PUT %s/restaurants/{slug}/dishes/{id}", p), a.handleUpdateDish)
		mux.HandleFunc(fmt.Sprintf("DELETE %s/restaurants/{slug}/dishes/{id}", p), a.handleDeleteDish)

		// Full Menu / Page Endpoint (returns { restaurant, categories, dishes })
		mux.HandleFunc(fmt.Sprintf("GET %s/restaurants/{slug}/menu", p), a.handleRestaurantMenu)
		mux.HandleFunc(fmt.Sprintf("GET %s/r/{slug}", p), a.handleRestaurantMenu)

		// Images / Uploads
		mux.HandleFunc(fmt.Sprintf("POST %s/restaurants/{slug}/dishes/{id}/image", p), a.handleUploadDishImage)
		mux.HandleFunc(fmt.Sprintf("POST %s/upload/image", p), a.handleGenericImageUpload)
	}
}
