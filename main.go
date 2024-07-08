package main

import (
	"fmt"
	"net/http"
	"strings"

	iamtokenvalidator "github.com/raccoon-mh/iamtokenvalidatorpoc"

	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
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
	protectedPath.Use(echojwt.WithConfig(echojwt.Config{
		KeyFunc:        iamtokenvalidator.Keyfunction,
		SuccessHandler: setRolesInContext,
	}))

	// token 이 valid 할 경우
	protectedPath.Any("", func(c echo.Context) error {
		roles := strings.Join(c.Get("roles").([]string), ", ")
		userId := c.Get("userId")
		msg := fmt.Sprintf("Hello, %s! Protected. roles : %s", userId, roles)
		return c.String(http.StatusOK, msg)
	})

	// token 이 valid 할 경우 && admin 역할을 강제하는 경우
	protectedPath.Any("/admin", SetGrantedRolesMiddleware([]string{"admin"})(func(c echo.Context) error {
		roles := strings.Join(c.Get("roles").([]string), ", ")
		userId := c.Get("userId")
		msg := fmt.Sprintf("Hello, %s Admin! Protected. roles : %s", userId, roles)
		return c.String(http.StatusOK, msg)
	}))

	e.Logger.Fatal(e.Start(":1323"))
}

func setRolesInContext(c echo.Context) {
	accesstoken := c.Get("user").(*jwt.Token).Raw
	claims, _ := iamtokenvalidator.GetTokenClaimsByIamManagerClaims(accesstoken)
	c.Set("userId", claims.UserId)
	c.Set("userName", claims.UserName)
	c.Set("preferredUsername", claims.PreferredUsername)
	c.Set("roles", claims.RealmAccess.Roles)
}

func SetGrantedRolesMiddleware(roles []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRolesArr := c.Get("roles").([]string)
			userRolesArrSet := make(map[string]struct{}, len(userRolesArr))
			for _, v := range userRolesArr {
				userRolesArrSet[v] = struct{}{}
			}
			for _, v := range roles {
				if _, found := userRolesArrSet[v]; found {
					return next(c)
				}
			}
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
	}
}
