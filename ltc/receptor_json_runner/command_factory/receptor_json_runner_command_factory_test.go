package command_factory_test

import (
    "errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/fake_tailed_logs_outputter"
	"github.com/cloudfoundry-incubator/lattice/ltc/receptor_json_runner/command_factory"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal"
	"github.com/cloudfoundry-incubator/lattice/ltc/test_helpers"
	"github.com/codegangsta/cli"
    "github.com/cloudfoundry-incubator/lattice/ltc/receptor_json_runner/fake_receptor_json_runner"
)

var _ = Describe("CommandFactory", func() {

	var (
		outputBuffer            *gbytes.Buffer
		ui                      terminal.UI
		fakeTailedLogsOutputter *fake_tailed_logs_outputter.FakeTailedLogsOutputter
        fakeReceptorJsonRunner  *fake_receptor_json_runner.FakeReceptorJsonRunner
	)

	BeforeEach(func() {
		outputBuffer = gbytes.NewBuffer()
		ui = terminal.NewUI(nil, outputBuffer, nil)
		fakeTailedLogsOutputter = fake_tailed_logs_outputter.NewFakeTailedLogsOutputter()
        fakeReceptorJsonRunner = &fake_receptor_json_runner.FakeReceptorJsonRunner{}
	})

	Describe("CreateJSONCommand", func() {
		var createJsonCommand cli.Command

		BeforeEach(func() {
			commandFactory := command_factory.NewReceptorJsonRunnerCommandFactory(fakeReceptorJsonRunner, ui, fakeTailedLogsOutputter)
			createJsonCommand = commandFactory.MakeCreateAppFromJsonCommand()
		})

		It("reads in json from the specified path and sends it off to the receptor", func() {
			args := []string{"test.json"}

            fakeReceptorJsonRunner.CreateAppFromJsonReturns(nil)  // happy path

			test_helpers.ExecuteCommandWithArgs(createJsonCommand, args)

			Expect(outputBuffer).To(test_helpers.Say("Attempting to Create LRP from test.json"))

            Expect(fakeReceptorJsonRunner.CreateAppFromJsonCallCount()).To(Equal(1))
            Expect(fakeReceptorJsonRunner.CreateAppFromJsonArgsForCall(0)).To(Equal("{\"fake\":\"json\"}"))
            Expect(fakeTailedLogsOutputter.OutputTailedLogsCallCount()).To(Equal(1))
		})

        Context("with an invalid file path", func(){
            It("reports an error and exits", func() {
                args := []string{"~/mad/wrong/path/bro/lrp.json"}

                fakeReceptorJsonRunner.CreateAppFromJsonReturns(nil)

                test_helpers.ExecuteCommandWithArgs(createJsonCommand, args)

                Expect(outputBuffer).To(test_helpers.Say("no such file or directory"))
            })
        })

        Context("with a file path that requires privileges", func(){
            It("reports an error and...", func(){
                // TODO: test for privileges message
            })
        })

        Context("with invalid json", func(){
            It("Bubbles up an error and exits", func(){
                args := []string{"test_invalid_json.json"}

                fakeReceptorJsonRunner.CreateAppFromJsonReturns(errors.New("Invalid JSON"))

                test_helpers.ExecuteCommandWithArgs(createJsonCommand, args)

                Expect(outputBuffer).To(test_helpers.Say("Invalid JSON"))
            })
        })
	})

})
