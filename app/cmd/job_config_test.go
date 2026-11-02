package cmd

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rody-vx/jenkins-cli/client"
	"github.com/rody-vx/jenkins-cli/mock/mhttp"
)

var _ = Describe("job config command", func() {
	var (
		ctrl         *gomock.Controller
		roundTripper *mhttp.MockRoundTripper
		dir          string
		configURL    = "http://localhost:8080/jenkins/job/folder/job/deploy/config.xml"
		config       = "<flow-definition/>\n"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		roundTripper = mhttp.NewMockRoundTripper(ctrl)
		jobConfigOption.RoundTripper = roundTripper
		jobConfigOption.Filename = ""
		jobConfigOption.Batch = false
		rootOptions.Jenkins = ""
		rootOptions.ConfigFile = "test.yaml"

		data, err := GenerateSampleConfig()
		Expect(err).To(BeNil())
		Expect(ioutil.WriteFile(rootOptions.ConfigFile, data, 0664)).To(Succeed())

		dir, err = ioutil.TempDir("", "job-config")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		rootCmd.SetArgs([]string{})
		os.Remove(rootOptions.ConfigFile)
		os.RemoveAll(dir)
		rootOptions.ConfigFile = ""
		ctrl.Finish()
	})

	It("get saves config.xml into a private file", func() {
		request, _ := http.NewRequest(http.MethodGet, configURL, nil)
		request.SetBasicAuth("admin", "111e3a2f0231198855dceaff96f20540a9")
		roundTripper.EXPECT().RoundTrip(client.NewRequestMatcher(request)).Return(&http.Response{
			StatusCode: 200,
			Request:    request,
			Body:       ioutil.NopCloser(bytes.NewBufferString(config)),
		}, nil)

		file := filepath.Join(dir, "config.xml")
		rootCmd.SetArgs([]string{"job", "config", "get", "folder/deploy", "-f", file})
		Expect(rootCmd.Execute()).To(Succeed())

		data, err := ioutil.ReadFile(file)
		Expect(err).To(BeNil())
		Expect(string(data)).To(Equal(config))

		info, err := os.Stat(file)
		Expect(err).To(BeNil())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0600)))
	})

	It("update refuses a file which is not XML, without any request", func() {
		file := filepath.Join(dir, "config.xml")
		Expect(ioutil.WriteFile(file, []byte("not xml"), 0600)).To(Succeed())

		rootCmd.SetArgs([]string{"job", "config", "update", "folder/deploy", "-f", file, "-b"})
		Expect(rootCmd.Execute()).To(HaveOccurred())
	})

	It("update refuses an empty file, without any request", func() {
		file := filepath.Join(dir, "config.xml")
		Expect(ioutil.WriteFile(file, []byte(" \n"), 0600)).To(Succeed())

		rootCmd.SetArgs([]string{"job", "config", "update", "folder/deploy", "-f", file, "-b"})
		Expect(rootCmd.Execute()).To(HaveOccurred())
	})
})
