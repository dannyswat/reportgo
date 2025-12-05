// Package main demonstrates simple usage of the ReportGo library.
package main

import (
	"log"

	"github.com/dannyswat/reportgo/pkg/reportgo"
)

func main() {
	// Create a new report engine
	engine := reportgo.New()

	// Load template
	if err := engine.LoadTemplate("templates/examples/report.xml"); err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	// Prepare data
	data := map[string]interface{}{
		"ReportTitle":    "Monthly Sales Report",
		"ReportSubtitle": "November 2025",
		"Summary":        "This report provides a comprehensive overview of sales performance for November 2025. Overall revenue increased by 15% compared to the previous month.",
		"TotalRevenue":   "$125,450.00",
		"TotalOrders":    "1,234",
		"AvgOrderValue":  "$101.66",
		"SalesData": []map[string]interface{}{
			{"product": "Widget A", "quantity": 500, "unit_price": 25.00, "total": 12500.00},
			{"product": "Widget B", "quantity": 350, "unit_price": 45.00, "total": 15750.00},
			{"product": "Widget C", "quantity": 200, "unit_price": 75.00, "total": 15000.00},
			{"product": "Widget D", "quantity": 150, "unit_price": 120.00, "total": 18000.00},
		},
		"Notes": []interface{}{
			"Sales exceeded target by 10%",
			"New product launch contributed 25% of revenue",
			"Customer retention rate improved to 85%",
		},
	}

	// Generate PDF
	if err := engine.Generate(data, "output.pdf"); err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}

	log.Println("PDF generated successfully: output.pdf")
}
