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
func GeneratePDFReport(allDevices, affectedDevices, unaffectedDevices []map[string]string, reportName string) error {
	m := GetMaroto(allDevices, affectedDevices, unaffectedDevices)
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

func GetMaroto(allDevices, affectedDevices, unaffectedDevices []map[string]string) core.Maroto {
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

	// All Devices Table
	m.AddRows(text.NewRow(10, "All PAN-OS NGFW Devices", props.Text{
		Top:   3,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))
	m.AddRow(7, text.NewCol(12, "All PAN-OS NGFW Devices", props.Text{
		Top:   1.5,
		Size:  9,
		Style: fontstyle.Bold,
		Align: align.Center,
		Color: &props.WhiteColor,
	})).WithStyle(&props.Cell{BackgroundColor: darkGrayColor})
	m.AddRows(getDeviceRows(allDevices, "allDevices")...)

	// Add some space between tables
	m.AddRow(10, col.New(12))

	// Affected Devices Table
	m.AddRows(text.NewRow(10, "NGFW Devices Requiring PAN-OS Upgrade", props.Text{
		Top:   3,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))
	m.AddRow(7, text.NewCol(12, "NGFW Devices Requiring PAN-OS Upgrade", props.Text{
		Top:   1.5,
		Size:  9,
		Style: fontstyle.Bold,
		Align: align.Center,
		Color: &props.WhiteColor,
	})).WithStyle(&props.Cell{BackgroundColor: darkGrayColor})
	m.AddRows(getDeviceRows(affectedDevices, "affectedDevices")...)

	// Add some space between tables
	m.AddRow(10, col.New(12))

	// Unaffected Devices Table
	m.AddRows(text.NewRow(10, "WildFire Registration Results", props.Text{
		Top:   3,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))
	m.AddRow(7, text.NewCol(12, "WildFire Registration Results", props.Text{
		Top:   1.5,
		Size:  9,
		Style: fontstyle.Bold,
		Align: align.Center,
		Color: &props.WhiteColor,
	})).WithStyle(&props.Cell{BackgroundColor: darkGrayColor})
	m.AddRows(getDeviceRows(affectedDevices, "unaffectedDevices")...)

	return m

}

func getDeviceRows(deviceList []map[string]string, tableType string) []core.Row {
	var headerRow core.Row
	var contentRows []core.Row

	switch tableType {
	case "allDevices":
		headerRow = getAllDevicesHeaderRow()
		contentRows = getAllDevicesContentRows(deviceList)
	case "affectedDevices":
		headerRow = getAffectedDevicesHeaderRow()
		contentRows = getAffectedDevicesContentRows(deviceList)
	case "unaffectedDevices":
		headerRow = getUnaffectedDevicesHeaderRow()
		contentRows = getUnaffectedDevicesContentRows(deviceList)
	default:
		log.Fatalf("Unknown table type: %s", tableType)
	}

	return append([]core.Row{headerRow}, contentRows...)
}

func getAllDevicesHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "SW Version", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "Serial", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getAffectedDevicesHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "SW Version", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Serial", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Min Update", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getUnaffectedDevicesHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(6, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(6, "Result", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getAllDevicesContentRows(deviceList []map[string]string) []core.Row {
	var rows []core.Row
	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(2, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["sw-version"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["model"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(3, device["ip-address"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(3, device["serial"], props.Text{Size: 7, Align: align.Left}),
		)
		if i%2 == 0 {
			r.WithStyle(&props.Cell{BackgroundColor: getGrayColor()})
		}
		rows = append(rows, r)
	}
	return rows
}

func getAffectedDevicesContentRows(deviceList []map[string]string) []core.Row {
	var rows []core.Row
	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(2, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["sw-version"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["model"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["ip-address"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["serial"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["minimumUpdateRelease"], props.Text{Size: 7, Align: align.Left}),
		)
		if i%2 == 0 {
			r.WithStyle(&props.Cell{BackgroundColor: getGrayColor()})
		}
		rows = append(rows, r)
	}
	return rows
}

func getUnaffectedDevicesContentRows(deviceList []map[string]string) []core.Row {
	var rows []core.Row
	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(6, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(6, device["result"], props.Text{Size: 7, Align: align.Left}),
		)
		if i%2 == 0 {
			r.WithStyle(&props.Cell{BackgroundColor: getGrayColor()})
		}
		rows = append(rows, r)
	}
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
		Red:   222,
		Green: 222,
		Blue:  222,
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
