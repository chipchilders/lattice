package receptor_json_runner_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/cloudfoundry-incubator/lattice/ltc/receptor_json_runner"
    "github.com/cloudfoundry-incubator/receptor/fake_receptor"
    "github.com/cloudfoundry-incubator/receptor"
    "errors"
)

var _ = Describe("Receptor Json Runner", func() {

    var(
        fakeReceptorClient          *fake_receptor.FakeClient
        createRequest               receptor.DesiredLRPCreateRequest
        receptorJsonRunner          receptor_json_runner.ReceptorJsonRunner
    )

    BeforeEach(func() {
        fakeReceptorClient = &fake_receptor.FakeClient{}
        createRequest = receptor.DesiredLRPCreateRequest{}
        receptorJsonRunner = receptor_json_runner.New(fakeReceptorClient, createRequest)
    })

    Describe("CreateAppFromJson", func() {
        It("Takes a string of valid JSON", func(){
            err := receptorJsonRunner.CreateAppFromJson("{\"name\":\"cool-web-app\", \"dockerimagepath\":\"superfun/app\",\"startCommand\":\"sampleStartCommand\"}")

            Expect(err).ToNot(HaveOccurred())
        })

        Context("Given invalid JSON", func(){
            It("returns an error and exits", func(){
                err := receptorJsonRunner.CreateAppFromJson("{\"definitely not valid json}")

                Expect(err).To(HaveOccurred())
            })
        })

        Context("Given the desired app is already running", func(){
            It("prints a message to that effect and exits", func(){

            })
        })

        Context("Given upsert domain error", func(){
            It("returns an error and exits", func(){
                fakeReceptorClient.UpsertDomainReturns(errors.New("Upsert Domain Error"))
                err := receptorJsonRunner.CreateAppFromJson("{\"name\":\"cool-web-app\", \"dockerimagepath\":\"superfun/app\",\"startCommand\":\"sampleStartCommand\"}")

                Expect(err).To(HaveOccurred())
                Expect(err.Error()).To(Equal("Upsert Domain Error"))
            })
        })
    })
})
