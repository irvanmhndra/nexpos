package handler

import "github.com/labstack/echo/v5"

func getCompanyID(c *echo.Context) int64 {
	if v := c.Get("company_id"); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 1 // fallback for development
}

func getUserID(c *echo.Context) int64 {
	if v := c.Get("user_id"); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

func getBranchID(c *echo.Context) int64 {
	if v := c.Get("branch_id"); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 1 // fallback for development
}
