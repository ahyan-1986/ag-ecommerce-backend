package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/jung-kurt/gofpdf"
)

// GenerateReceipt creates a standalone receipt PDF for an order.
// Returns the PDF as a byte buffer.
func GenerateReceipt(order *domain.Order) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ── Header ────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(79, 70, 229) // indigo
	pdf.CellFormat(0, 12, "HOPME", "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 5, "Premium Cat Litter Store", "", 1, "C", false, 0, "")
	pdf.Ln(2)
	pdf.SetDrawColor(79, 70, 229)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)

	// ── Receipt Title ─────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 8, "PAYMENT RECEIPT", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// ── Order Info Box ────────────────────────────────────────────────────
	drawInfoBox(pdf, [][2]string{
		{"Receipt No", order.OrderNo},
		{"Date", order.CreatedAt.Format("02 Jan 2006 03:04 PM")},
		{"Order Status", string(order.Status)},
		{"Payment Method", order.PaymentMethod},
	})
	pdf.Ln(3)

	// ── Customer Info ─────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 7, "Customer Details", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(70, 70, 70)
	if order.User != nil {
		name := order.User.FirstName + " " + order.User.LastName
		pdf.CellFormat(0, 6, "Name: "+name, "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 6, "Email: "+order.User.Email, "", 1, "L", false, 0, "")
	}
	pdf.Ln(3)

	// ── Shipping Address ──────────────────────────────────────────────────
	if order.ShippingLine1 != "" {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(0, 7, "Shipping Address", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(70, 70, 70)
		pdf.CellFormat(0, 6, order.ShippingName, "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 6, order.ShippingPhone, "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 6, order.ShippingLine1, "", 1, "L", false, 0, "")
		if order.ShippingLine2 != "" {
			pdf.CellFormat(0, 6, order.ShippingLine2, "", 1, "L", false, 0, "")
		}
		loc := order.ShippingCity
		if order.ShippingState != "" {
			loc += ", " + order.ShippingState
		}
		loc += " " + order.ShippingPostcode
		pdf.CellFormat(0, 6, loc, "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 6, order.ShippingCountry, "", 1, "L", false, 0, "")
		pdf.Ln(3)
	}

	// ── Items Table ────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(0, 7, "Items", "", 1, "L", false, 0, "")

	// Table header
	pdf.SetFillColor(245, 245, 255)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(60, 60, 60)
	colWidths := map[string]float64{"item": 85, "qty": 25, "price": 30, "total": 30}
	pdf.SetX(15)
	pdf.CellFormat(colWidths["item"], 7, "Item", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colWidths["qty"], 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colWidths["price"], 7, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(colWidths["total"], 7, "Amount", "1", 1, "R", true, 0, "")

	// Table rows
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(50, 50, 50)
	for _, item := range order.Items {
		pdf.SetX(15)
		name := item.ProductName
		if item.VariantName != "" && item.VariantName != name {
			name += " - " + item.VariantName
		}
		pdf.CellFormat(colWidths["item"], 7, truncate(name, 42), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths["qty"], 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths["price"], 7, fmt.Sprintf("MYR %.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths["total"], 7, fmt.Sprintf("MYR %.2f", item.TotalPrice), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(3)

	// ── Totals ─────────────────────────────────────────────────────────────
	totalWidth := 60.0
	startX := 195.0 - totalWidth
	drawTotalRow(pdf, startX, totalWidth, "Subtotal", fmt.Sprintf("MYR %.2f", order.SubTotal), false)
	if order.ShippingFee > 0 {
		drawTotalRow(pdf, startX, totalWidth, "Shipping", fmt.Sprintf("MYR %.2f", order.ShippingFee), false)
	}
	if order.DiscountAmount > 0 {
		drawTotalRow(pdf, startX, totalWidth, "Discount", fmt.Sprintf("-MYR %.2f", order.DiscountAmount), false)
	}
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(startX, pdf.GetY()+1, startX+totalWidth, pdf.GetY()+1)
	pdf.Ln(2)
	drawTotalRow(pdf, startX, totalWidth, "Total", fmt.Sprintf("MYR %.2f", order.TotalAmount), true)
	pdf.Ln(4)

	// ── Stamp if paid ──────────────────────────────────────────────────────
	if order.Status == domain.OrderStatusPaid {
		pdf.SetFont("Helvetica", "B", 18)
		pdf.SetTextColor(34, 197, 94) // green
		pdf.CellFormat(0, 10, "PAID", "", 1, "C", false, 0, "")
		if order.PaidAt != nil {
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(100, 100, 100)
			pdf.CellFormat(0, 6, "Paid on "+order.PaidAt.Format("02 Jan 2006 03:04 PM"), "", 1, "C", false, 0, "")
		}
		pdf.Ln(4)
	}

	// ── Payment Ref ────────────────────────────────────────────────────────
	if order.PaymentRef != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, "Payment Reference: "+order.PaymentRef, "", 1, "C", false, 0, "")
	}
	if order.VoucherCode != "" {
		pdf.CellFormat(0, 5, "Voucher: "+order.VoucherCode, "", 1, "C", false, 0, "")
	}

	// ── Footer ─────────────────────────────────────────────────────────────
	pdf.SetY(-25)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 4, "HOPME — Premium Cat Litter Store", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 4, "Thank you for your purchase!", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 4, "Generated on "+time.Now().Format("02 Jan 2006 03:04 PM"), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

func drawInfoBox(pdf *gofpdf.Fpdf, rows [][2]string) {
	pdf.SetFillColor(249, 250, 251)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(60, 60, 60)
	for _, row := range rows {
		pdf.SetX(15)
		pdf.CellFormat(45, 6, row[0], "1", 0, "L", true, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(40, 40, 40)
		pdf.CellFormat(120, 6, row[1], "1", 1, "L", true, 0, "")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(60, 60, 60)
	}
}

func drawTotalRow(pdf *gofpdf.Fpdf, x, w float64, label, value string, bold bool) {
	pdf.SetX(x)
	if bold {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(40, 40, 40)
	} else {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(70, 70, 70)
	}
	pdf.CellFormat(w/2, 7, label+":", "", 0, "R", false, 0, "")
	if bold {
		pdf.SetFont("Helvetica", "B", 11)
	} else {
		pdf.SetFont("Helvetica", "", 10)
	}
	pdf.CellFormat(w/2, 7, value, "", 1, "R", false, 0, "")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
