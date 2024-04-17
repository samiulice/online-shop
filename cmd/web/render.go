package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type templateData struct {
	StringMap       map[string]string
	IntMap          map[string]int
	FloatMap        map[string]float64
	Data            map[string]interface{}
	CSRFToken       string
	Flash           string
	Warning         string
	Error           string
	IsAuthenticated string
	API             string
	CSSVersion      string
	StripeSecretKey      string
	StripePublishableKey      string
}

var functions = template.FuncMap{
	"formatCurrency" : formatCurrency,
}

func formatCurrency(n int)(string) {
	return fmt.Sprintf("$%.2f", float32(n/100))
}

//go:embed templates
var templateFS embed.FS

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	td.API = app.config.api
	td.StripePublishableKey = app.config.stripe.key
	td.StripeSecretKey = app.config.stripe.secret
	return td
}
func (app *application) renderTemplate(w http.ResponseWriter, r *http.Request, page string, td *templateData, partials ...string) error {
	var t *template.Template
	var err error

	templateToRender := fmt.Sprintf("templates/%s.page.gohtml", page)

	_, templateInMap := app.temlateCache[templateToRender]

	if app.config.env == "production" && templateInMap {
		t = app.temlateCache[templateToRender]
	} else {
		t, err = app.parseTemplate(partials, page, templateToRender)
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

	if err != nil{
		app.errorLog.Println(err)
		return err
	}
	return nil
}

func (app *application) parseTemplate(partials []string, page string, templateToRender string) (*template.Template, error) {
	var t *template.Template
	var err error

	//build partials
	if len(partials) > 0 {
		for i, v := range partials {
			partials[i] = fmt.Sprintf("templates/%s.partials.gohtml", v)
		}
	}

	if len(partials) > 0 {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.gohtml", strings.Join(partials, ","), templateToRender)
	} else {
		t, err = template.New(fmt.Sprintf("%s.page.gohtml", page)).Funcs(functions).ParseFS(templateFS, "templates/base.layout.gohtml", templateToRender)
	}

	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	app.temlateCache[templateToRender] = t

	return t, nil
}
