package casbin

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
	xormadapter "github.com/casbin/xorm-adapter/v2"
)

const (
	host     = "postgres2"
	port     = "5432"
	dbname   = "casbin"
	username = "postgres"
	password = "1234"
)

func CasbinEnforcer(logger *slog.Logger) (*casbin.Enforcer, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", host, port, username, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Error connecting to database", "error", err.Error())
		return nil, err
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS casbin")
	if err != nil {
		logger.Error("Error dropping Casbin database", "error", err.Error())
		return nil, err
	}

	adapter, err := xormadapter.NewAdapter("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, username, dbname, password))
	if err != nil {
		logger.Error("Error creating Casbin adapter", "error", err.Error())
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer("casbin/model.conf", adapter)
	if err != nil {
		logger.Error("Error creating Casbin enforcer", "error", err.Error())
		return nil, err
	}

	err = enforcer.LoadPolicy()
	if err != nil {
		logger.Error("Error loading Casbin policy", "error", err.Error())
		return nil, err
	}

	policies := [][]string{
		//card
		{"consumer", "/cards", "POST"},
		{"seller", "/cards", "POST"},
		{"consumer", "/cards", "GET"},
		{"seller", "/cards", "GET"},
		{"consumer", "/cards/amount/:card_number", "GET"},
		{"seller", "/cards/amount/:card_number", "GET"},
		{"seller", "/cards/:card_number", "DELETE"},
		{"consumer", "/cards/:card_number", "DELETE"},

		// product
		{"seller", "/products", "POST"},
		{"seller", "/products/list", "POST"},
		{"consumer", "/products/list", "POST"},
		{"admin", "/products/list", "POST"},
		{"seller", "/products/:id", "GET"},
		{"consumer", "/products/:id", "GET"},
		{"admin", "/products/:id", "GET"},
		{"seller", "/products", "GET"},
		{"seller", "/products/:id", "PUT"},
		{"seller", "/products/:id", "DELETE"},
		{"seller", "/products/photo/:product_id", "POST"},
		{"seller", "/products/photo/:product_id", "DELETE"},
		{"seller", "/products/limit/:product_id", "PUT"},

		// process
		{"consumer", "/process/buy", "POST"},
		{"seller", "/process/products/:product_id", "GET"},
		{"admin", "/process/products/:product_id", "GET"},
		{"admin", "/process/user/:product_id/:user_id", "GET"},
		{"consumer", "/process", "GET"},
		{"consumer", "/process/:id", "GET"},
		{"seller", "/process/:id", "GET"},
		{"admin", "/process/:id", "GET"},
		{"seller", "/process/:id", "PUT"},
		{"admin", "/process/:id", "PUT"},
		{"consumer", "/process/:id", "DELETE"},

		// Wishlist
		{"consumer", "/wishlist/:product_id", "POST"},
		{"consumer", "/wishlist", "GET"},
		{"consumer", "/wishlist/:id", "GET"},

		// Feedback
		{"consumer", "/feedback/:product_id", "POST"},
		{"consumer", "/feedback", "GET"},
		{"consumer", "/feedback/:product_id", "GET"},
		{"seller", "/feedback/:product_id", "GET"},
		{"admin", "/feedback/:product_id", "GET"},

		// Bought
		{"seller", "/bought/:product_id", "GET"},
		{"admin", "/bought/:product_id", "GET"},
		{"consumer", "/bought", "GET"},

		// Notifications
		{"consumer", "/notifications/all", "GET"},
		{"seller", "/notifications/all", "GET"},
		{"consumer", "/notifications/unreaden", "GET"},
		{"seller", "/notifications/unreaden", "GET"},
	}

	_, err = enforcer.AddPolicies(policies)
	if err != nil {
		logger.Error("Error adding Casbin policy", "error", err.Error())
		return nil, err
	}

	err = enforcer.SavePolicy()
	if err != nil {
		logger.Error("Error saving Casbin policy", "error", err.Error())
		return nil, err
	}
	return enforcer, nil
}
