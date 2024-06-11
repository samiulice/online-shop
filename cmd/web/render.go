package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"online_store/internal/models"
	"strings"
	"time"
)

type templateData struct {
	StringMap            map[string]string
	IntMap               map[string]int
	FloatMap             map[string]float64
	Data                 map[string]interface{}
	CSRFToken            string
	Flash                string
	Warning              string
	Error                string
	IsAuthenticated      int
	API                  string
	CSSVersion           string
	StripeSecretKey      string
	StripePublishableKey string
	User                 models.User
	Employee                 models.Employee
}

var funcMap = template.FuncMap{
	"titleCase":      titleCase,
	"formatCurrency": formatCurrency,
	"formatDate":     formatDate,
	"Userlink": Userlink,
}

// titleCase returns a copy of the string s with all Unicode letters mapped to their Unicode title case
func titleCase(s string) string {
	if len(strings.Split(s, " ")) < 2 {
		return strings.ToUpper(string(s[0])) + s[1:]
	}
	return strings.ToTitle(s)
}

// formatCurrency returns the currency with prefix $
func formatCurrency(n int) string {
	return fmt.Sprintf("$%.2f", float64(n)/100.0)
}

// FormatDate returns Date in a specific format
func formatDate(t time.Time, format string) string {
	return t.Format(format)
}
// Userlink returns link for user
func Userlink(str string) string {
	return str[:len(str)-1]
}

//go:embed templates/*
var templateFS embed.FS

// addDefaultData adds default variables to the templatedata
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripePublishableKey = app.config.stripe.key
	td.StripeSecretKey = app.config.stripe.secret
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
		td.User = app.Session.Get(r.Context(), "user").(models.User)
	} else {
		td.IsAuthenticated = 0
	}
	return td
}

func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	_, templateInMap := app.temlateCache[templateToRender]

	if templateInMap && app.config.env != "development" {
		t = app.temlateCache[templateToRender]
	} else {
		t, err = app.parseTemplate(partials, page)
		if err != nil {
			app.errorLog.Println(err)
			return nil
		}
	}

	if td == nil {
		td = &templateData{}
	}

	td = app.addDefaultData(td, r)
	err = t.Execute(w, td)

	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	return nil
}

func (app *application) parseTemplate(partials []string, page string) (*template.Template, error) {
	var t *template.Template
	var err error

	//Identifying Base Template path
	var baseTemplate string
	if strings.Contains(page, "admin") {
		baseTemplate = "templates/admin.layout.gohtml"
	} else if strings.Contains(page, "employee") {
		baseTemplate = "templates/employee.layout.gohtml"
	} else {
		baseTemplate = "templates/base.layout.gohtml"
	}

	//Identifying Template path that needs to be rendered
	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	//build partials
	if len(partials) > 0 {
		for i, v := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partials.gohtml", v)
		}
	}

	//patterns of Templates path that is to be parsed into the single template
	var templates []string
	templates = append(templates, baseTemplate, templateToRender)
	templates = append(templates, partials...)

	t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(funcMap).ParseFS(templateFS, templates...)

	if err != nil {
		return nil, err
	}

	app.temlateCache[templateToRender] = t

	return t, nil
}
