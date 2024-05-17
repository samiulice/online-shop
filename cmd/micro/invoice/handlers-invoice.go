package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
)

// Order holds the necessary info to build invoice
type Order struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Items     []Product `json:"items"`
}
type Product struct {
	ID       int    `json:"product_id"`
	Name     string `json:"product_name"`
	Quantity int    `json:"quantity"`
	Amount   int    `json:"amount"`
}

func (app *application) GenerateInvoiceAndSend(w http.ResponseWriter, r *http.Request) {
	//receive json
	var order Order

	err := app.readJSON(w, r, &order)

	if err != nil {
		app.badRequest(w, err)
		return
	}

	//generate a pdf invoice
	err = app.CreateInvoicePDF(order)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	//create mail
	attachments := []string{
		fmt.Sprintf("./invoices/%d_%s.pdf", order.ID, order.CreatedAt.Format("02-Jan-2006_15-04-05")),
	}
	//send mail with attachment //"coding.samiul@gmail.com" will be replace by oreder.Email in production mode
	err = app.SendMail("info@demomailtrap.com", "coding.samiul@gmail.com", "Order Invoice", "invoice", attachments, nil)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	//send response
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = fmt.Sprintf("Invoice %d.pdf created and sent to %s", order.ID, order.Email)
	app.writeJSON(w, http.StatusCreated, resp)
}

// GenerateTestInvoice generates test invoice
func (app *application) GenerateTestInvoice(w http.ResponseWriter, r *http.Request) {
	var items []Product
	var name = []string{
		"AMD Ryzen 9 7950X3D Gaming Processor",
		"Lian Li Galahad II TRINITY 360 ARGB AIO Liquid CPU Cooler",
		"Asus TUF GAMING B650M-E WIFI AMD AM5 micro-ATX Motherboard",
		"Corsair DOMINATOR PLATINUM RGB 32GB DDR5 5600MHz CL40 RAM White",
		"Samsung 980 Pro 1TB PCIe 4.0 M.2 NVMe SSD",
		"ASUS ROG LOKI SFX-L 750W 80 Plus Platinum Power Supply",
		"Corsair iCUE 5000X RGB Tempered Glass Mid-Tower ATX PC Smart Case"}
	var id = []int{7950, 23360, 6505, 3255600, 980142, 75080, 5000}
	var price = []int{79500, 16600, 31200, 17500, 1200, 22200, 19000}
	for i, v := range name {
		var p = Product{
			ID:       id[i],
			Name:     v,
			Quantity: 10,
			Amount:   price[i],
		}
		items = append(items, p)
	}
	var order = Order{
		ID:        7710526,
		FirstName: "Samiul",
		LastName:  "Islam",
		Email:     "coding.samiul@gmail.com",
		CreatedAt: time.Now(),
		Items:     items,
	}
	//generate a pdf invoice
	err := app.CreateInvoicePDF(order)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	//send response
	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = fmt.Sprintf("Test Invoice %d_%s.pdf Generated", order.ID, order.CreatedAt.Format("02-Jan-2006_15-04-05"))
	app.writeJSON(w, http.StatusCreated, resp)
}

func (app *application) CreateInvoicePDF(order Order) error {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)
	pdf.SetAutoPageBreak(true, 0)

	importer := gofpdi.NewImporter()

	t := importer.ImportPage(pdf, "./pdf-templates/invoice.pdf", 1, "/MediaBox")

	pdf.AddPage()
	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)

	//write invoice header
	pdf.SetX(10)
	pdf.SetY(50)
	pdf.SetFont("Times", "", 11)
	pdf.CellFormat(97, 8, fmt.Sprintf("Attention: %s %s", order.FirstName, order.LastName), "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.Email, "", 0, "L", false, 0, "")
	pdf.Ln(5)
	pdf.CellFormat(97, 8, order.CreatedAt.Format("Mon, 02 Jan 2006 15:04:05 MST"), "", 0, "L", false, 0, "")

	//write invoice descrption
	y := float64(93)
	var total float64
	for _, item := range order.Items {
		pdf.SetY(y)
		pdf.SetX(10)
		pdf.CellFormat(155, 8, item.Name, "", 0, "L", false, 0, "")
		pdf.SetX(166)
		pdf.CellFormat(20, 8, fmt.Sprintf("%d", item.Quantity), "", 0, "C", false, 0, "")
		pdf.SetX(185)
		pdf.CellFormat(20, 8, fmt.Sprintf("$%.2f", float64(item.Amount/100)), "", 0, "R", false, 0, "")

		total += float64(item.Amount*item.Quantity / 100)
		y += 5
		pdf.SetY(y)
		pdf.CellFormat(155, 8, fmt.Sprintf("(ID: %d)", item.ID), "", 0, "L", false, 0, "")
		y += 10		
	}
	

	//write invoice footer
	pdf.SetY(238)
	pdf.SetX(185)
	pdf.CellFormat(20, 8, fmt.Sprintf("$%.2f", total), "", 0, "R", false, 0, "")
	

	//Export PDF
	invoicePath := fmt.Sprintf("./invoices/%d_%s.pdf", order.ID, order.CreatedAt.Format("02-Jan-2006_15-04-05"))

	err := pdf.OutputFileAndClose(invoicePath)

	return err
}
