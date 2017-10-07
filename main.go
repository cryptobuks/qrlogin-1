package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
)

var (
	loginTokensMap    = make(map[string]string)
	useridpasswordMap = make(map[string]string)
	useridsessionMap  = make(map[string]string)
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo()
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func init() {
	useridpasswordMap["testuser1"] = "testuser1"
	useridpasswordMap["testuser2"] = "testuser2"
	useridpasswordMap["testuser3"] = "testuser3"

}

func main() {
	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}
	e.Renderer = renderer

	e.Static("/", "static")

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "qr_code.html", map[string]interface{}{})
	})

	e.GET("/login2", func(c echo.Context) error {
		return c.Render(http.StatusOK, "another.html", map[string]interface{}{})
	})

	e.POST("/home", func(c echo.Context) error {
		templateCtx := make(map[string]interface{})
		c.Request().ParseForm()
		templateCtx["id"] = c.Request().Form["id"]
		templateCtx["authenticated"] = c.Request().Form["authenticated"]

		return c.Render(http.StatusOK, "home.html", templateCtx)
	})

	e.POST("/home2", func(c echo.Context) error {
		templateCtx := make(map[string]interface{})
		c.Request().ParseForm()
		templateCtx["id"] = c.Request().Form["id"]
		templateCtx["authenticated"] = c.Request().Form["authenticated"]

		return c.Render(http.StatusOK, "home2.html", templateCtx)
	})

	e.GET("/home", func(c echo.Context) error { return c.Render(http.StatusOK, "home.html", map[string]interface{}{}) })
	e.GET("/home2", func(c echo.Context) error { return c.Render(http.StatusOK, "home2.html", map[string]interface{}{}) })

	e.GET("/health", func(c echo.Context) error { return c.JSON(http.StatusOK, "hello, sQuiRrel!") })

	e.GET("/generateLoginToken", generateLoginToken)

	e.POST("/validateLoginToken", validateLoginToken)

	e.POST("/checkLoginTokenStatus", checkLoginTokenStatus)

	e.POST("/doLogin", doLogin)

	e.Logger.Fatal(e.Start(":8089"))
}

func generateLoginToken(c echo.Context) error {
	fmt.Println("Inside generateLoginToken function")

	u1 := uuid.NewV4()
	loginTokensMap[u1.String()] = ""
	loginToken := LoginToken{Token: u1.String()}
	b, _ := json.Marshal(loginToken)

	fmt.Println("loginToken = ", loginToken)

	return c.JSON(http.StatusOK, string(b))
}

func validateLoginToken(context echo.Context) error {

	fmt.Println("Inside validateLoginToken function")

	requestBody := context.Request().Body
	body, _ := ioutil.ReadAll(requestBody)
	loginIdTokenRequest := LoginIdTokenRequest{}

	if body != nil {
		if err := json.Unmarshal([]byte(body), &loginIdTokenRequest); isError(err) {
			fmt.Println("Error unmarshal request")
			return context.JSON(http.StatusInternalServerError, getDefaultError())
		}
	}

	fmt.Println("Request body :", loginIdTokenRequest)

	_, ok := loginTokensMap[loginIdTokenRequest.Token]
	if ok == false {
		fmt.Println("Invalid login token")
		return context.JSON(http.StatusOK, "Invalid login token ")
	}

	fmt.Println("Valid login token. Userid : ", loginIdTokenRequest.Userid)
	loginTokensMap[loginIdTokenRequest.Token] = loginIdTokenRequest.Userid

	fmt.Println("Returning 200 success")
	return context.NoContent(http.StatusOK)
}

func checkLoginTokenStatus(context echo.Context) error {

	fmt.Println("Inside checkLoginTokenStatus function")

	requestBody := context.Request().Body
	body, _ := ioutil.ReadAll(requestBody)
	loginToken := LoginToken{}

	if body != nil {
		if err := json.Unmarshal([]byte(body), &loginToken); isError(err) {
			fmt.Println("Error unmarshal request")
			return context.JSON(http.StatusInternalServerError, getDefaultError())
		}
	}

	fmt.Println("Request body :", loginToken)

	uid, ok := loginTokensMap[loginToken.Token]
	if ok == false {
		fmt.Println("Invalid login token")
		return context.JSON(http.StatusOK, "Invalid login token")
	}

	var loginTokenStatus LoginTokenStatus

	if uid == "" {
		fmt.Println("uid = nil. Authenticated: false")
		loginTokenStatus = LoginTokenStatus{Id: "", Authenticated: false}
		return context.JSON(http.StatusOK, loginTokenStatus)
	} else {
		fmt.Println("uid =", uid, " Authenticated: true")
		loginTokenStatus = LoginTokenStatus{Id: uid, Authenticated: true}
		return context.JSON(http.StatusOK, loginTokenStatus)
	}

}

func doLogin(context echo.Context) error {

	fmt.Println("Inside doLogin function")

	requestBody := context.Request().Body
	body, _ := ioutil.ReadAll(requestBody)
	loginRequest := LoginRequest{}

	if body != nil {
		if err := json.Unmarshal([]byte(body), &loginRequest); isError(err) {
			fmt.Println("Error unmarshal request")
			return context.JSON(http.StatusInternalServerError, getDefaultError())
		}
	}

	fmt.Println("Request body :", loginRequest)

	password, ok := useridpasswordMap[loginRequest.Userid]
	if ok == false {
		fmt.Println("Invalid userid")
		return context.JSON(http.StatusOK, "Invalid userid or password")
	}
	if password != loginRequest.Password {
		fmt.Println("Invalid password")
		return context.JSON(http.StatusOK, "Invalid userid or password")
	}

	u1 := uuid.NewV4()

	fmt.Println("Login successful. Sessionid : ", u1)

	useridsessionMap[loginRequest.Userid] = u1.String()

	return context.JSON(http.StatusOK, "Login success")
}

func getDefaultError() *ErrorResponse {
	return &ErrorResponse{
		"200000",
		"An error occurred",
	}
}

type ErrorResponse struct {
	ResponseId string `json:"id"`
	Text       string `json:"text"`
}

func isError(err error) bool {
	if err != nil {
		return true
	}
	return false
}

type LoginToken struct {
	Token string `json:"token"`
}

type LoginTokenStatus struct {
	Id            string `json:"id"`
	Authenticated bool   `json:"authenticated"`
}

type LoginRequest struct {
	Userid   string `json:"userid"`
	Password string `json:"password"`
}

type LoginIdTokenRequest struct {
	Token  string `json:"token"`
	Userid string `json:"userid"`
}
