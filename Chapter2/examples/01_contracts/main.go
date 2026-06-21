package main

import (
	"context"
	"fmt"
)

// --- Контракт определяет ПОТРЕБИТЕЛЬ (report), не реализация ---

type ReportGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type ReportService struct {
	gen ReportGenerator
}

func NewReportService(gen ReportGenerator) *ReportService {
	return &ReportService{gen: gen}
}

func (s *ReportService) BuildDaily(ctx context.Context) (string, error) {
	return s.gen.Generate(ctx)
}

// --- Реализации: не знают про ReportGenerator ---

type PDFReport struct{}

func (PDFReport) Generate(ctx context.Context) (string, error) {
	return "PDF report content", nil
}

type CSVReport struct{}

func (CSVReport) Generate(ctx context.Context) (string, error) {
	return "id,amount\n1,100", nil
}

// Compile-time check: CSVReport satisfies ReportGenerator
var _ ReportGenerator = CSVReport{}

func main() {
	ctx := context.Background()

	// Подмена реализации без изменения ReportService
	for _, gen := range []ReportGenerator{PDFReport{}, CSVReport{}} {
		svc := NewReportService(gen)
		out, err := svc.BuildDaily(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println(out)
	}
}
