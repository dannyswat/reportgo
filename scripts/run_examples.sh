#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"

cd "$repo_root"

echo "Generating simple example PDF..."
go run ./cmd/reportgo \
	-template templates/examples/report.xml \
	-data examples/simple/data.json \
	-output examples/simple/simple.pdf

echo "Generating payslip example PDF..."
go run ./cmd/reportgo \
	-template templates/examples/payslip.xml \
	-data examples/payslip/data.json \
	-output examples/payslip/payslip.pdf

echo "Generating rowgrid example PDF..."
go run ./cmd/reportgo \
	-template templates/examples/rowgrid.xml \
	-data examples/rowgrid/data.json \
	-output examples/rowgrid/rowgrid.pdf

echo "Generating header_footer example PDF..."
go run ./cmd/reportgo \
	-template templates/examples/header_footer.xml \
	-data examples/header_footer/data.json \
	-output examples/header_footer/header_footer.pdf

echo "Generating showcase example PDF..."
go run ./examples/showcase

echo "Generated PDFs:"
echo "  examples/simple/simple.pdf"
echo "  examples/payslip/payslip.pdf"
echo "  examples/rowgrid/rowgrid.pdf"
echo "  examples/header_footer/header_footer.pdf"
echo "  examples/showcase/showcase.pdf"