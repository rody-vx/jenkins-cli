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
})
