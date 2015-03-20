package command_factory

import (
	"fmt"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/receptor_json_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/codegangsta/cli"
    "os"
)

//appRunner:             config.AppRunner,
//dockerMetadataFetcher: config.DockerMetadataFetcher,
//output:                config.Output,
//timeout:               config.Timeout,
//domain:                config.Domain,
//env:                   config.Env,
//clock:                 config.Clock,
//tailedLogsOutputter:   config.TailedLogsOutputter,

type ReceptorJsonCommandFactory struct {
	receptorJsonRunner receptor_json_runner.ReceptorJsonRunner
	ui                 terminal.UI
	tlo                console_tailed_logs_outputter.TailedLogsOutputter
}

func NewReceptorJsonRunnerCommandFactory(receptorJsonRunner receptor_json_runner.ReceptorJsonRunner, ui terminal.UI, tlo console_tailed_logs_outputter.TailedLogsOutputter) *ReceptorJsonCommandFactory {
	return &ReceptorJsonCommandFactory{
		receptorJsonRunner: receptorJsonRunner,
		ui:                 ui,
		tlo:                tlo,
	}
}

func (factory *ReceptorJsonCommandFactory) MakeCreateAppFromJsonCommand() cli.Command {
	var createCommand = cli.Command{
		Name:        "create-from-json",
		ShortName:   "cfj",
		Usage:       "ltc create-from-json /path/to/file.json",
		Description: `Create a LRP on lattice from JSON object`,
		Action:      factory.createAppFromJson,
	}
	return createCommand
}

func (factory *ReceptorJsonCommandFactory) createAppFromJson(c *cli.Context) {
    filePath := c.Args().First()

    file, err := os.Open(filePath)
    if err != nil {
        factory.ui.Say(err.Error())
        return
    }

    data := make([]byte, 100)
    _, err = file.Read(data)
    if err != nil {
        factory.ui.Say(err.Error())
        return
    }

    factory.ui.Say(fmt.Sprintf("Attempting to Create LRP from %s", c.Args().First()))

    err = factory.receptorJsonRunner.CreateAppFromJson(string(data[:]))
    if err != nil {
        factory.ui.Say(err.Error())
        return
    }

//    factory.tlo.OutputTailedLogs("the app guid goes here")
    // stream logs
    // poll until start
}
