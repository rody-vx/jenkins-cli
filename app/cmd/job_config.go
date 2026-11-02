package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/rody-vx/jenkins-cli/app/cmd/common"
	"github.com/rody-vx/jenkins-cli/app/i18n"
	"github.com/rody-vx/jenkins-cli/client"
	"github.com/spf13/cobra"
)

// JobConfigOption is the option for the job config commands
type JobConfigOption struct {
	common.BatchOption
	common.Option

	Filename string
}

var jobConfigOption JobConfigOption

func init() {
	jobCmd.AddCommand(jobConfigCmd)
	jobConfigCmd.AddCommand(jobConfigGetCmd, jobConfigUpdateCmd)

	jobConfigGetCmd.Flags().StringVarP(&jobConfigOption.Filename, "filename", "f", "",
		i18n.T("The file to save config.xml. Print to stdout if it is empty"))

	jobConfigUpdateCmd.Flags().StringVarP(&jobConfigOption.Filename, "filename", "f", "",
		i18n.T("The config.xml file to apply"))
	_ = jobConfigUpdateCmd.MarkFlagRequired("filename")
	jobConfigOption.SetFlag(jobConfigUpdateCmd)

	jobConfigOption.Option.Stdio = common.GetSystemStdio()
	jobConfigOption.BatchOption.Stdio = common.GetSystemStdio()
}

func (o *JobConfigOption) newClient() *client.JobClient {
	jclient := &client.JobClient{
		JenkinsCore: client.JenkinsCore{
			RoundTripper: o.RoundTripper,
		},
	}
	getCurrentJenkinsAndClientOrDie(&(jclient.JenkinsCore))
	return jclient
}

var jobConfigCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("Get or update the config.xml of a job"),
	Long:  i18n.T("Get or update the config.xml of a job. It supports every job type without any plugin"),
}

var jobConfigGetCmd = &cobra.Command{
	Use:   "get <jobName>",
	Short: i18n.T("Get the config.xml of a job"),
	Long:  i18n.T("Get the config.xml of a job"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var config string
		if config, err = jobConfigOption.newClient().GetConfig(args[0]); err != nil {
			return
		}

		if jobConfigOption.Filename != "" {
			// config.xml can contain encrypted secrets
			return ioutil.WriteFile(jobConfigOption.Filename, []byte(config), 0600)
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), config)
		return
	},
}

var jobConfigUpdateCmd = &cobra.Command{
	Use:   "update <jobName> -f <file>",
	Short: i18n.T("Replace the config.xml of a job"),
	Long:  i18n.T("Replace the config.xml of a job. Save the current one with 'job config get' before updating it"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var data []byte
		if data, err = ioutil.ReadFile(jobConfigOption.Filename); err != nil {
			return
		}
		if !strings.HasPrefix(strings.TrimSpace(string(data)), "<") {
			return fmt.Errorf("%s is not an XML file", jobConfigOption.Filename)
		}

		jclient := jobConfigOption.newClient()
		if !jobConfigOption.Confirm(fmt.Sprintf("Are you sure to replace the config of job %s on %s ?", args[0], jclient.URL)) {
			return
		}
		return jclient.UpdateConfig(args[0], string(data))
	},
}
