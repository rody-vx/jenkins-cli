package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rody-vx/jenkins-cli/client"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rody-vx/jenkins-cli/mock/mhttp"
)

var _ = Describe("job search command", func() {
	var (
		ctrl         *gomock.Controller
		roundTripper *mhttp.MockRoundTripper
		rootURL      string
		username     string
		token        string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		roundTripper = mhttp.NewMockRoundTripper(ctrl)
		jobParamOption.RoundTripper = roundTripper
		rootCmd.SetArgs([]string{})
		rootOptions.Jenkins = ""
		rootOptions.ConfigFile = "test.yaml"

		rootURL = "http://localhost:8080/jenkins"
		username = "admin"
		token = "111e3a2f0231198855dceaff96f20540a9"
	})

	AfterEach(func() {
		rootCmd.SetArgs([]string{})
		os.Remove(rootOptions.ConfigFile)
		rootOptions.ConfigFile = ""
		ctrl.Finish()
	})

	Context("basic cases, need RoundTripper", func() {
		It("without parameters", func() {
			data, err := GenerateSampleConfig()
			Expect(err).To(BeNil())
			err = ioutil.WriteFile(rootOptions.ConfigFile, data, 0664)
			Expect(err).To(BeNil())

			jobName := "fake"
			client.PrepareForGetJob(roundTripper, rootURL, jobName, username, token)
			rootCmd.SetArgs([]string{"job", "param", jobName})

			buf := new(bytes.Buffer)
			rootCmd.SetOutput(buf)
			_, err = rootCmd.ExecuteC()
			Expect(err).To(BeNil())

			Expect(buf.String()).To(Equal(""))
		})

		It("with one parameter", func() {
			data, err := GenerateSampleConfig()
			Expect(err).To(BeNil())
			err = ioutil.WriteFile(rootOptions.ConfigFile, data, 0664)
			Expect(err).To(BeNil())

			jobName := "fake"
			client.PrepareForGetJobWithParams(roundTripper, rootURL, jobName, username, token)
			rootCmd.SetArgs([]string{"job", "param", jobName})

			buf := new(bytes.Buffer)
			rootCmd.SetOutput(buf)
			_, err = rootCmd.ExecuteC()
			Expect(err).To(BeNil())

			Expect(buf.String()).To(Equal("[{\"Description\":\"\",\"name\":\"name\",\"Type\":\"StringParameterDefinition\",\"value\":\"\",\"DefaultParameterValue\":{\"Description\":\"\",\"Value\":\"jake\"}}]\n"))
		})

		It("with one parameter, output with indent", func() {
			data, err := GenerateSampleConfig()
			Expect(err).To(BeNil())
			err = ioutil.WriteFile(rootOptions.ConfigFile, data, 0664)
			Expect(err).To(BeNil())

			jobName := "fake"
			client.PrepareForGetJobWithParams(roundTripper, rootURL, jobName, username, token)
			rootCmd.SetArgs([]string{"job", "param", jobName, "--indent=true"})

			buf := new(bytes.Buffer)
			rootCmd.SetOutput(buf)
			_, err = rootCmd.ExecuteC()
			Expect(err).To(BeNil())

			Expect(buf.String()).To(Equal("[\n {\n  \"Description\": \"\",\n  \"name\": \"name\",\n  \"Type\": \"StringParameterDefinition\",\n  \"value\": \"\",\n  \"DefaultParameterValue\": {\n   \"Description\": \"\",\n   \"Value\": \"jake\"\n  }\n }\n]\n"))
		})

		Context("add or remove parameters via config.xml", func() {
			const paramsXML = `  <properties>
    <hudson.model.ParametersDefinitionProperty>
      <parameterDefinitions>
        <hudson.model.StringParameterDefinition>
          <name>A</name>
          <description>desc</description>
          <defaultValue>1</defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
`
			var (
				emptyConfig  string
				paramsConfig string
			)

			BeforeEach(func() {
				data, err := GenerateSampleConfig()
				Expect(err).To(BeNil())
				Expect(ioutil.WriteFile(rootOptions.ConfigFile, data, 0664)).To(Succeed())

				emptyConfig = client.SamplePipelineConfig("echo 1")
				paramsConfig = strings.Replace(emptyConfig, "  <properties/>\n", paramsXML, 1)
			})

			It("adds a parameter", func() {
				client.PrepareForGetJobConfig(roundTripper, rootURL, "/job/fake", emptyConfig, username, token)
				client.PrepareForUpdateJobConfig(roundTripper, rootURL, "/job/fake", paramsConfig, username, token)
				client.PrepareForGetJob(roundTripper, rootURL, "fake", username, token)

				rootCmd.SetArgs([]string{"job", "param", "fake", "--add", `[{"name":"A","value":"1","desc":"desc"}]`})
				_, err := rootCmd.ExecuteC()
				Expect(err).To(BeNil())
			})

			It("removes a parameter", func() {
				client.PrepareForGetJobConfig(roundTripper, rootURL, "/job/fake", paramsConfig, username, token)
				client.PrepareForUpdateJobConfig(roundTripper, rootURL, "/job/fake",
					strings.Replace(emptyConfig, "  <properties/>\n", "  <properties>\n  </properties>\n", 1), username, token)
				client.PrepareForGetJob(roundTripper, rootURL, "fake", username, token)

				rootCmd.SetArgs([]string{"job", "param", "fake", "--remove", "A", "--add", ""})
				_, err := rootCmd.ExecuteC()
				Expect(err).To(BeNil())
			})

			It("fails without saving when the parameter does not exist", func() {
				client.PrepareForGetJobConfig(roundTripper, rootURL, "/job/fake", emptyConfig, username, token)

				rootCmd.SetArgs([]string{"job", "param", "fake", "--remove", "NONE", "--add", ""})
				_, err := rootCmd.ExecuteC()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
