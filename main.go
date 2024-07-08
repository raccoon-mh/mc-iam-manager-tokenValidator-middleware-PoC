package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/m-cmp/mc-iam-manager/iamtokenvalidator"
)

func init() {
	err := iamtokenvalidator.GetPubkeyIamManager("https://example.com:5000/api/auth/certs") // mc-iam-manager certs endpoint is require, this endpoint is v0.2.0..
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	e := echo.New()

	e.Any("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	protectedPath := e.Group("/protected")
	protectedPath.Use(isTokenValid)
	protectedPath.Use(setUserRole)
	protectedPath.Any("/", func(c echo.Context) error {
		roles := strings.Join(c.Get("realmAccess").([]string), ", ")
		return c.String(http.StatusOK, "Hello, World! Protected. your roles is : "+roles)
	})

	e.Logger.Fatal(e.Start(":1323"))
}

func isTokenValid(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accesstoken := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
		err := iamtokenvalidator.IsTokenValid(accesstoken)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Authorization is not valid..")
		}
		c.Set("accesstoken", accesstoken)
		return next(c)
	}
}

func setUserRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		accesstoken := c.Get("accesstoken").(string)
		claims, err := iamtokenvalidator.GetTokenClaimsByIamManagerClaims(accesstoken)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Authorization is not valid..")
		}
		c.Set("realmAccess", claims.RealmAccess.Roles)
		return next(c)
	}
}
