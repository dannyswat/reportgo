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

echo "Generating showcase example PDF..."
go run ./examples/showcase

echo "Generated PDFs:"
echo "  examples/simple/simple.pdf"
echo "  examples/payslip/payslip.pdf"
echo "  examples/showcase/showcase.pdf"