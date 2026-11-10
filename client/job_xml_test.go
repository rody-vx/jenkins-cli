package client

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("job config.xml editing", func() {
	Context("pipeline script", func() {
		It("reads the script with XML 1.1 declaration and escaped text", func() {
			pipeline, err := pipelineScriptFromConfig(SamplePipelineConfig("echo &apos;a &amp; b&apos;\nsh &quot;ls&quot;"))
			Expect(err).To(BeNil())
			Expect(pipeline.Script).To(Equal("echo 'a & b'\nsh \"ls\""))
			Expect(pipeline.Sandbox).To(BeTrue())
		})

		It("replaces only the script and keeps the rest", func() {
			result, err := replacePipelineScript(SamplePipelineConfig("old"), "echo '<a>' & \"b\"")
			Expect(err).To(BeNil())
			Expect(result).To(Equal(SamplePipelineConfig("echo &apos;&lt;a&gt;&apos; &amp; &quot;b&quot;")))

			pipeline, err := pipelineScriptFromConfig(result)
			Expect(err).To(BeNil())
			Expect(pipeline.Script).To(Equal("echo '<a>' & \"b\""))
		})

		It("adds a definition to a pipeline job which has no script yet", func() {
			config := "<?xml version='1.1' encoding='UTF-8'?>\n<flow-definition plugin=\"workflow-job@1.0\">\n  <properties/>\n</flow-definition>"
			result, err := replacePipelineScript(config, "echo 1")
			Expect(err).To(BeNil())
			Expect(result).To(Equal("<?xml version='1.1' encoding='UTF-8'?>\n<flow-definition plugin=\"workflow-job@1.0\">\n  <properties/>\n" +
				"  <definition class=\"org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition\">\n" +
				"    <script>echo 1</script>\n    <sandbox>true</sandbox>\n  </definition>\n</flow-definition>"))
		})

		It("keeps the lines of a multi-line script as they are when adding a definition", func() {
			config := "<flow-definition>\n  <properties/>\n</flow-definition>"
			script := "pipeline {\n    agent any\n}"
			result, err := replacePipelineScript(config, script)
			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring("    <script>pipeline {\n    agent any\n}</script>\n"))

			pipeline, err := pipelineScriptFromConfig(result)
			Expect(err).To(BeNil())
			Expect(pipeline.Script).To(Equal(script))
		})

		It("fills an empty script element", func() {
			config := "<flow-definition>\n  <definition class=\"org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition\">\n    <script/>\n  </definition>\n</flow-definition>"
			result, err := replacePipelineScript(config, "echo 1")
			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring("    <script>echo 1</script>\n"))
		})

		It("rejects other job types and SCM pipelines", func() {
			_, err := pipelineScriptFromConfig("<project><builders/></project>")
			Expect(err).To(HaveOccurred())

			scm := "<flow-definition><definition class=\"org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition\"><scriptPath>Jenkinsfile</scriptPath></definition></flow-definition>"
			_, err = pipelineScriptFromConfig(scm)
			Expect(err).To(HaveOccurred())
			_, err = replacePipelineScript(scm, "echo 1")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("parameters", func() {
		const withParams = `<?xml version='1.1' encoding='UTF-8'?>
<project>
  <properties>
    <hudson.model.ParametersDefinitionProperty>
      <parameterDefinitions>
        <hudson.model.StringParameterDefinition>
          <name>BRANCH</name>
          <defaultValue>main</defaultValue>
        </hudson.model.StringParameterDefinition>
        <net.uaznia.lukanus.hudson.plugins.gitparameter.GitParameterDefinition plugin="git-parameter@1.0">
          <name>TAG</name>
        </net.uaznia.lukanus.hudson.plugins.gitparameter.GitParameterDefinition>
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
</project>`

		It("parses an array or a single object", func() {
			specs, err := ParseParameterSpecs(`[{"name":"a"},{"name":"b"}]`)
			Expect(err).To(BeNil())
			Expect(specs).To(HaveLen(2))

			specs, err = ParseParameterSpecs(` {"name":"a","value":"1"}`)
			Expect(err).To(BeNil())
			Expect(specs).To(Equal([]ParameterSpec{{Name: "a", Value: "1"}}))

			_, err = ParseParameterSpecs(`name`)
			Expect(err).To(HaveOccurred())
		})

		It("appends to the existing parameter definitions", func() {
			result, err := addParametersToConfig(withParams, []ParameterSpec{{Name: "NAME", Value: "rick & co", Desc: "a name"}})
			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring(`        </net.uaznia.lukanus.hudson.plugins.gitparameter.GitParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>NAME</name>
          <description>a name</description>
          <defaultValue>rick &amp; co</defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
      </parameterDefinitions>`))
		})

		It("creates the parameter property in empty properties", func() {
			config := "<flow-definition>\n  <properties/>\n</flow-definition>"
			result, err := addParametersToConfig(config, []ParameterSpec{
				{Name: "DEBUG", Type: "boolean", Value: "true"},
				{Name: "ENV", Type: "ChoiceParameterDefinition", Choices: []string{"dev", "qa"}},
			})
			Expect(err).To(BeNil())
			Expect(result).To(Equal(`<flow-definition>
  <properties>
    <hudson.model.ParametersDefinitionProperty>
      <parameterDefinitions>
        <hudson.model.BooleanParameterDefinition>
          <name>DEBUG</name>
          <description></description>
          <defaultValue>true</defaultValue>
        </hudson.model.BooleanParameterDefinition>
        <hudson.model.ChoiceParameterDefinition>
          <name>ENV</name>
          <description></description>
          <choices class="java.util.Arrays$ArrayList">
            <a class="string-array">
              <string>dev</string>
              <string>qa</string>
            </a>
          </choices>
        </hudson.model.ChoiceParameterDefinition>
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
</flow-definition>`))

			root, err := parseXMLNodes(result)
			Expect(err).To(BeNil())
			_, definitions := parameterDefinitions(root)
			Expect(definitions.Children).To(HaveLen(2))
		})

		It("keeps multi-line values", func() {
			result, err := addParametersToConfig(withParams, []ParameterSpec{{Name: "NOTE", Type: "text", Value: "a\n  b", Desc: "x\ny"}})
			Expect(err).To(BeNil())

			root, err := parseXMLNodes(result)
			Expect(err).To(BeNil())
			_, definitions := parameterDefinitions(root)
			note := parameterNode(definitions, "NOTE")
			Expect(note.child("defaultValue").Text).To(Equal("a\n  b"))
			Expect(note.child("description").Text).To(Equal("x\ny"))
		})

		It("rejects invalid parameters", func() {
			for _, spec := range []ParameterSpec{
				{Name: "BRANCH"},
				{Name: ""},
				{Name: "A", Type: "boolean", Value: "yes"},
				{Name: "A", Type: "choice"},
				{Name: "A", Type: "file"},
			} {
				_, err := addParametersToConfig(withParams, []ParameterSpec{spec})
				Expect(err).To(HaveOccurred(), spec.Name+spec.Type)
			}
		})

		It("removes parameters and the property when none is left", func() {
			result, err := removeParametersFromConfig(withParams, []string{"TAG"})
			Expect(err).To(BeNil())
			Expect(result).NotTo(ContainSubstring("TAG"))
			Expect(result).To(ContainSubstring("        </hudson.model.StringParameterDefinition>\n      </parameterDefinitions>"))

			result, err = removeParametersFromConfig(result, []string{"BRANCH"})
			Expect(err).To(BeNil())
			Expect(result).To(Equal("<?xml version='1.1' encoding='UTF-8'?>\n<project>\n  <properties>\n  </properties>\n</project>"))

			_, err = removeParametersFromConfig(withParams, []string{"NONE"})
			Expect(err).To(HaveOccurred())
		})

		It("splits parameter names", func() {
			Expect(splitParameterNames(" a, b,,c ")).To(Equal([]string{"a", "b", "c"}))
		})
	})
})
