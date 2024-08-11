// Package pdfgenerate/pdf.go
package pdfgenerate

import (
	"log"
	"os"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// GeneratePDFReport creates a PDF report using the maroto library.
func GeneratePDFReport(deviceList []map[string]string, reportName string) error {
	m := GetMaroto(deviceList)
	document, err := m.Generate()
	if err != nil {
		return err
	}

	reportDir := "report"

	// Ensure the report directory exists
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		err = os.Mkdir(reportDir, 0755)
		if err != nil {
			return err
		}
	}

	err = document.Save(filepath.Join(reportDir, reportName))
	if err != nil {
		return err
	}

	return nil
}

func GetMaroto(deviceList []map[string]string) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	darkGrayColor := getDarkGrayColor()
	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	err := m.RegisterHeader(getPageHeader())
	if err != nil {
		log.Fatal(err.Error())
	}

	err = m.RegisterFooter(getPageFooter())
	if err != nil {
		log.Fatal(err.Error())
	}

	m.AddRows(text.NewRow(10, "Device Report", props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))

	m.AddRow(7,
		text.NewCol(3, "All PAN-OS NGFW Devices", props.Text{
			Top:   1.5,
			Size:  9,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.WhiteColor,
		}),
	).WithStyle(&props.Cell{BackgroundColor: darkGrayColor})

	// Add the device rows (including headers)
	m.AddRows(getDeviceRows(deviceList)...)

	return m
}

func getDeviceRows(deviceList []map[string]string) []core.Row {
	rows := []core.Row{
		row.New(5).Add(
			text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
			text.NewCol(2, "SW Version", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
			text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
			text.NewCol(2, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
			text.NewCol(2, "Serial", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		),
	}

	var contentsRow []core.Row

	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(2, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["sw-version"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["model"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["ip-address"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["serial"], props.Text{Size: 7, Align: align.Left}),
		)
		if i%2 == 0 {
			gray := getGrayColor()
			r.WithStyle(&props.Cell{BackgroundColor: gray})
		}

		contentsRow = append(contentsRow, r)
	}

	rows = append(rows, contentsRow...)

	return rows
}

func getPageHeader() core.Row {
	return row.New(20).Add(
		image.NewFromFileCol(3, "docs/assets/images/logo.png", props.Rect{
			Center:  true,
			Percent: 80,
		}),
		col.New(6),
		col.New(3).Add(
			text.New("CDSS Certificate Report", props.Text{
				Top:   5,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Right,
				Color: getBlueColor(),
			}),
		),
	)
}

func getPageFooter() core.Row {
	return row.New(20).Add(
		col.New(12).Add(
			text.New("github.com/cdot65/pan-os-cdss-certificate-registration", props.Text{
				Top:   13,
				Style: fontstyle.BoldItalic,
				Size:  8,
				Align: align.Left,
				Color: getBlueColor(),
			}),
		),
	)
}

func getDarkGrayColor() *props.Color {
	return &props.Color{
		Red:   55,
		Green: 55,
		Blue:  55,
	}
}

func getGrayColor() *props.Color {
	return &props.Color{
		Red:   200,
		Green: 200,
		Blue:  200,
	}
}

func getBlueColor() *props.Color {
	return &props.Color{
		Red:   10,
		Green: 10,
		Blue:  150,
	}
}

func getRedColor() *props.Color {
	return &props.Color{
		Red:   150,
		Green: 10,
		Blue:  10,
	}
}
