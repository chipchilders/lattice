package command_factory

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner/command_factory/presentation"
	"github.com/cloudfoundry-incubator/lattice/ltc/colors"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler"
	"github.com/cloudfoundry-incubator/lattice/ltc/output"
	"github.com/cloudfoundry-incubator/lattice/ltc/output/cursor"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/clock"
)

const TimestampDisplayLayout = "2006-01-02 15:04:05 (MST)"

type AppExaminerCommandFactory struct {
    appExaminer app_examiner.AppExaminer
    output      *output.Output
    clock       clock.Clock
    exitHandler exit_handler.ExitHandler
}

func NewAppExaminerCommandFactory(appExaminer app_examiner.AppExaminer, output *output.Output, clock clock.Clock, exitHandler exit_handler.ExitHandler) *AppExaminerCommandFactory {
	return &AppExaminerCommandFactory{appExaminer, output, clock, exitHandler}
}

func (factory *AppExaminerCommandFactory) MakeListAppCommand() cli.Command {

	var startCommand = cli.Command{
		Name:        "list",
		ShortName:   "li",
		Description: "List all applications running on Lattice",
		Usage:       "ltc list",
		Action:      factory.listApps,
		Flags:       []cli.Flag{},
	}

	return startCommand
}

func (factory *AppExaminerCommandFactory) MakeVisualizeCommand() cli.Command {

	var visualizeFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "rate, r",
			Usage: "The rate at which to refresh the visualization.\n\te.g. -r=\".5s\"\n\te.g. -r=\"1000ns\"",
		},
	}

	var startCommand = cli.Command{
		Name:        "visualize",
        ShortName:   "v",
		Description: "Visualize the workload distribution across the Lattice Cells",
		Usage:       "ltc visualize",
		Action:      factory.visualizeCells,
		Flags:       visualizeFlags,
	}

	return startCommand
}

func (factory *AppExaminerCommandFactory) MakeStatusCommand() cli.Command {
	return cli.Command{
		Name:        "status",
        ShortName:   "st",
		Description: "Displays detailed status information about the given application and its instances",
		Usage:       "ltc status APP_NAME",
		Action:      factory.appStatus,
		Flags:       []cli.Flag{},
	}
}

func (factory *AppExaminerCommandFactory) listApps(context *cli.Context) {
	appList, err := factory.appExaminer.ListApps()
	if err != nil {
		factory.output.Say("Error listing apps: " + err.Error())
		return
	} else if len(appList) == 0 {
		factory.output.Say("No apps to display.")
		return
	}

	w := &tabwriter.Writer{}
	w.Init(factory.output, 10+colors.ColorCodeLength, 8, 1, '\t', 0)

	header := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", colors.Bold("App Name"), colors.Bold("Instances"), colors.Bold("DiskMB"), colors.Bold("MemoryMB"), colors.Bold("Route"))
	fmt.Fprintln(w, header)

	for _, appInfo := range appList {
		var displayedRoute string
		if appInfo.Routes != nil && len(appInfo.Routes) > 0 {
			arbitraryPort := appInfo.Ports[0]
			displayedRoute = fmt.Sprintf("%s => %d", strings.Join(appInfo.Routes.HostnamesByPort()[arbitraryPort], ", "), arbitraryPort)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", colors.Bold(appInfo.ProcessGuid), colorInstances(appInfo), colors.NoColor(strconv.Itoa(appInfo.DiskMB)), colors.NoColor(strconv.Itoa(appInfo.MemoryMB)), colors.Cyan(displayedRoute))
	}

	w.Flush()
}

func printHorizontalRule(w io.Writer, pattern string) {
	header := strings.Repeat(pattern, 80) + "\n"
	fmt.Fprintf(w, header)
}

func (factory *AppExaminerCommandFactory) appStatus(context *cli.Context) {
	if len(context.Args()) < 1 {
		factory.output.IncorrectUsage("App Name required")
		return
	}

	appName := context.Args()[0]
	appInfo, err := factory.appExaminer.AppStatus(appName)

	if err != nil {
		factory.output.Say(err.Error())
		return
	}

	minColumnWidth := 13
	w := tabwriter.NewWriter(factory.output, minColumnWidth, 8, 1, '\t', 0)

	headingPrefix := strings.Repeat(" ", minColumnWidth/2)

	titleBar := func(title string) {
		printHorizontalRule(w, "=")
		fmt.Fprintf(w, "%s%s\n", headingPrefix, title)
		printHorizontalRule(w, "-")
	}

	titleBar(colors.Bold(appName))

	printAppInfo(w, appInfo)

	fmt.Fprintln(w, "")
	printHorizontalRule(w, "=")

	printInstanceInfo(w, headingPrefix, appInfo.ActualInstances)
	w.Flush()
}

func printAppInfo(w io.Writer, appInfo app_examiner.AppInfo) {

	fmt.Fprintf(w, "%s\t%s\n", "Instances", colorInstances(appInfo))
	fmt.Fprintf(w, "%s\t%s\n", "Stack", appInfo.Stack)

	fmt.Fprintf(w, "%s\t%d\n", "Start Timeout", appInfo.StartTimeout)
	fmt.Fprintf(w, "%s\t%d\n", "DiskMB", appInfo.DiskMB)
	fmt.Fprintf(w, "%s\t%d\n", "MemoryMB", appInfo.MemoryMB)
	fmt.Fprintf(w, "%s\t%d\n", "CPUWeight", appInfo.CPUWeight)

	portStrings := make([]string, 0)
	for _, port := range appInfo.Ports {
		portStrings = append(portStrings, fmt.Sprint(port))
	}

	fmt.Fprintf(w, "%s\t%s\n", "Ports", strings.Join(portStrings, ","))

	formatRoute := func(hostname string, port uint16) string {
		return colors.Cyan(fmt.Sprintf("%s => %d", hostname, port))
	}

	routeStringsByPort := appInfo.Routes.HostnamesByPort()
	var i int
	for port, routeStrs := range routeStringsByPort {
		if i == 0 {
			fmt.Fprintf(w, "%s\t%s\n", "Routes", formatRoute(routeStrs[0], port))
			if len(routeStrs) > 1 {
				for _, routeStr := range routeStrs[1:] {
					fmt.Fprintf(w, "\t%s\n", formatRoute(routeStr, port))
				}
			}
		} else {
			for _, routeStr := range routeStrs {
				fmt.Fprintf(w, "\t%s\n", formatRoute(routeStr, port))
			}
		}
		i++
	}

	if appInfo.Annotation != "" {
		fmt.Fprintf(w, "%s\t%s\n", "Annotation", appInfo.Annotation)
	}

	printHorizontalRule(w, "-")
	var envVars string
	for _, envVar := range appInfo.EnvironmentVariables {
		envVars += envVar.Name + `="` + envVar.Value + `" ` + "\n"
	}
	fmt.Fprintf(w, "%s\n\n%s", "Environment", envVars)

}

func printInstanceInfo(w io.Writer, headingPrefix string, actualInstances []app_examiner.InstanceInfo) {
	instanceBar := func(index, state string) {
		fmt.Fprintf(w, "%sInstance %s  [%s]\n", headingPrefix, index, state)
		printHorizontalRule(w, "-")
	}

	for _, instance := range actualInstances {
		instanceBar(fmt.Sprint(instance.Index), presentation.ColorInstanceState(instance))

		if instance.PlacementError == "" && instance.State != "CRASHED" {
			fmt.Fprintf(w, "%s\t%s\n", "InstanceGuid", instance.InstanceGuid)
			fmt.Fprintf(w, "%s\t%s\n", "Cell ID", instance.CellID)
			fmt.Fprintf(w, "%s\t%s\n", "Ip", instance.Ip)

			portMappingStrings := make([]string, 0)
			for _, portMapping := range instance.Ports {
				portMappingStrings = append(portMappingStrings, fmt.Sprintf("%d:%d", portMapping.HostPort, portMapping.ContainerPort))
			}
			fmt.Fprintf(w, "%s\t%s\n", "Port Mapping", strings.Join(portMappingStrings, ";"))

			fmt.Fprintf(w, "%s\t%s\n", "Since", fmt.Sprint(time.Unix(0, instance.Since).Format(TimestampDisplayLayout)))

		} else if instance.State != "CRASHED" {
			fmt.Fprintf(w, "%s\t%s\n", "Placement Error", instance.PlacementError)
		}
		fmt.Fprintf(w, "%s \t%d \n", "Crash Count", instance.CrashCount)
		printHorizontalRule(w, "-")
	}
}

func (factory *AppExaminerCommandFactory) visualizeCells(context *cli.Context) {
	rate := context.Duration("rate")

	factory.output.Say(colors.Bold("Distribution\n"))
	linesWritten := factory.printDistribution()

	if rate == 0 {
		return
	}

	closeChan := make(chan bool)
	factory.output.Say(cursor.Hide())

	factory.exitHandler.OnExit(func() {
		closeChan <- true
		factory.output.Say(cursor.Show())
	})

	for {
		select {
		case <-closeChan:
			return
		case <-factory.clock.NewTimer(rate).C():
			factory.output.Say(cursor.Up(linesWritten))
			linesWritten = factory.printDistribution()
		}
	}
}

func (factory *AppExaminerCommandFactory) printDistribution() int {
	defer factory.output.Say(cursor.ClearToEndOfDisplay())

	cells, err := factory.appExaminer.ListCells()
	if err != nil {
		factory.output.Say("Error visualizing: " + err.Error())
		factory.output.Say(cursor.ClearToEndOfLine())
		factory.output.NewLine()
		return 1
	}

	for _, cell := range cells {
		factory.output.Say(cell.CellID)
		if cell.Missing {
			factory.output.Say(colors.Red("[MISSING]"))
		}
		factory.output.Say(": ")

		if cell.RunningInstances == 0 && cell.ClaimedInstances == 0 && !cell.Missing {
			factory.output.Say(colors.Red("empty"))
		} else {
			factory.output.Say(colors.Green(strings.Repeat("•", cell.RunningInstances)))
			factory.output.Say(colors.Yellow(strings.Repeat("•", cell.ClaimedInstances)))
		}
		factory.output.Say(cursor.ClearToEndOfLine())
		factory.output.NewLine()
	}

	return len(cells)
}

func colorInstances(appInfo app_examiner.AppInfo) string {
	instances := fmt.Sprintf("%d/%d", appInfo.ActualRunningInstances, appInfo.DesiredInstances)
	if appInfo.ActualRunningInstances == appInfo.DesiredInstances {
		return colors.Green(instances)
	} else if appInfo.ActualRunningInstances == 0 {
		return colors.Red(instances)
	}

	return colors.Yellow(instances)
}
