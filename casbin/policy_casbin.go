package casbin

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
	xormadapter "github.com/casbin/xorm-adapter/v2"
)

const (
	host     = "localhost"
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
		{"consumer", "/cards", "POST"},
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
