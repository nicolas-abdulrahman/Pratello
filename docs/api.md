# Pratello API

Base URL: `http://localhost:8080/api

## Restaurants
| Method | Route | Body | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/restaurants` | — | List all restaurants |
| `GET` | `/restaurants/{slug}` | — | Get restaurant by slug |
| `POST` | `/restaurants` | `{ name, slug, description }` | Create new restaurant |
| `PUT` | `/restaurants/{slug}` | `{ name, description }` | Update restaurant details |
| `DELETE` | `/restaurants/{slug}` | — | Remove restaurant |

## Dishes
| Method | Route | Body | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/restaurants/{slug}/dishes` | — | List dishes for a restaurant |
| `POST` | `/restaurants/{slug}/dishes` | `{ name, price }` | Add dish to restaurant |
| `PUT` | `/restaurants/{slug}/dishes/{id}` | `{ name, price }` | Update dish |
| `DELETE` | `/restaurants/{slug}/dishes/{id}` | — | Remove dish |
