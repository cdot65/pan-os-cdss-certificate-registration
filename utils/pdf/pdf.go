// Package pdfgenerate/pdf.go
package pdf

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
func GeneratePDFReport(allDevices, ineligibleHardware, unsupportedVersions, registrationCandidates []map[string]string, reportName string) error {
	m := GetMaroto(allDevices, ineligibleHardware, unsupportedVersions, registrationCandidates)
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

func GetMaroto(allDevices, ineligibleHardware, unsupportedVersions, registrationCandidates []map[string]string) core.Maroto {
	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

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
	addDevicesTable(m, allDevices, "All PAN-OS NGFW Devices", "List of all NGFW devices that will be considered for this job", "allDevices")

	// Ineligible Hardware Table
	addDevicesTable(m, ineligibleHardware, "Skipped Because of Hardware", "Devices with hardware platforms unaffected by services registration with Device Certificate", "ineligibleHardware")

	// Unsupported Versions Table
	addDevicesTable(m, unsupportedVersions, "Skipped Because of PAN-OS Versions", "Devices that require a PAN-OS upgrade to support Device Certificate registration to CDSS services", "unsupportedVersions")

	// Registration Candidates Table
	addDevicesTable(m, registrationCandidates, "WildFire Registration Candidates", "Devices eligible for WildFire registration with device certificate", "registrationCandidates")

	return m

}

func addDevicesTable(m core.Maroto, devices []map[string]string, title, description, tableType string) {
	darkGrayColor := getDarkGrayColor()

	m.AddRows(text.NewRow(10, title, props.Text{
		Top:   3,
		Size:  12,
		Style: fontstyle.Bold,
		Align: align.Center,
	}))
	m.AddRow(7, text.NewCol(12, description, props.Text{
		Top:   1.5,
		Size:  9,
		Style: fontstyle.Bold,
		Align: align.Center,
		Color: &props.WhiteColor,
	})).WithStyle(&props.Cell{BackgroundColor: darkGrayColor})
	m.AddRows(getDeviceRows(devices, tableType)...)

	// Add some space between tables
	m.AddRow(10, col.New(12))
}

func getDeviceRows(deviceList []map[string]string, tableType string) []core.Row {
	var headerRow core.Row
	var contentRows []core.Row

	switch tableType {
	case "allDevices":
		headerRow = getAllDevicesHeaderRow()
		contentRows = getAllDevicesContentRows(deviceList)
	case "ineligibleHardware":
		headerRow = getIneligibleHardwareHeaderRow()
		contentRows = getIneligibleHardwareContentRows(deviceList)
	case "unsupportedVersions":
		headerRow = getUnsupportedVersionsHeaderRow()
		contentRows = getUnsupportedVersionsContentRows(deviceList)
	case "registrationCandidates":
		headerRow = getRegistrationCandidatesHeaderRow()
		contentRows = getRegistrationCandidatesContentRows(deviceList)
	default:
		log.Fatalf("Unknown table type: %s", tableType)
	}

	return append([]core.Row{headerRow}, contentRows...)
}

func getIneligibleHardwareHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Family", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "Serial", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getIneligibleHardwareContentRows(deviceList []map[string]string) []core.Row {
	var rows []core.Row
	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(2, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["model"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["family"], props.Text{Size: 7, Align: align.Left}),
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

func getUnsupportedVersionsHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "SW Version", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "Serial", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getUnsupportedVersionsContentRows(deviceList []map[string]string) []core.Row {
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

func getRegistrationCandidatesHeaderRow() core.Row {
	return row.New(5).Add(
		text.NewCol(2, "Hostname", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "SW Version", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(2, "Model", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "IP Address", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
		text.NewCol(3, "Result", props.Text{Size: 8, Align: align.Left, Style: fontstyle.Bold}),
	)
}

func getRegistrationCandidatesContentRows(deviceList []map[string]string) []core.Row {
	var rows []core.Row
	for i, device := range deviceList {
		r := row.New(4).Add(
			text.NewCol(2, device["hostname"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["sw-version"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(2, device["model"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(3, device["ip-address"], props.Text{Size: 7, Align: align.Left}),
			text.NewCol(3, device["result"], props.Text{Size: 7, Align: align.Left}),
		)
		if i%2 == 0 {
			r.WithStyle(&props.Cell{BackgroundColor: getGrayColor()})
		}
		rows = append(rows, r)
	}
	return rows
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

func getPageHeader() core.Row {
	return row.New(20).Add(
		image.NewFromFileCol(3, "docs/assets/images/logo.png", props.Rect{
			Center:  true,
			Percent: 80,
		}),
		col.New(6),
		col.New(3).Add(
			text.New("CDSS Services Registration With Device Certificate Report", props.Text{
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
