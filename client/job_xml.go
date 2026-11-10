package client

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// config.xml 을 고칠 때 바꾸려는 요소만 문자열로 교체하고 나머지는 원문 그대로 둔다.
// encoding/xml 로 통째로 다시 쓰면 플러그인 속성·순서·서식이 달라질 수 있기 때문이다.

const (
	cpsFlowDefinition    = "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition"
	cpsScmFlowDefinition = "org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition"
	parametersProperty   = "hudson.model.ParametersDefinitionProperty"
)

// xmlNode 는 config.xml 요소와 원문에서의 위치
type xmlNode struct {
	Name  string
	Attrs map[string]string
	// Start~End 는 요소 전체, InnerStart~InnerEnd 는 내용
	Start, InnerStart, InnerEnd, End int
	SelfClosing                      bool
	Text                             string
	Children                         []*xmlNode
}

func (n *xmlNode) child(name string) *xmlNode {
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

var xmlDeclPattern = regexp.MustCompile(`^\s*<\?xml[^>]*\?>`)

// parseXMLNodes 는 config.xml 을 위치 정보가 있는 요소 트리로 읽는다
func parseXMLNodes(doc string) (root *xmlNode, err error) {
	// Jenkins 는 XML 1.1 로 선언하지만 encoding/xml 은 1.0 만 받으므로, 위치가 어긋나지 않게 선언부를 공백으로 바꿔 읽는다
	src := doc
	if loc := xmlDeclPattern.FindStringIndex(src); loc != nil {
		src = strings.Repeat(" ", loc[1]) + src[loc[1]:]
	}

	decoder := xml.NewDecoder(strings.NewReader(src))
	var stack []*xmlNode
	for {
		offset := int(decoder.InputOffset())
		var token xml.Token
		if token, err = decoder.Token(); err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return nil, fmt.Errorf("cannot parse config.xml: %v", err)
		}
		end := int(decoder.InputOffset())

		switch t := token.(type) {
		case xml.StartElement:
			node := &xmlNode{Name: t.Name.Local, Attrs: map[string]string{}, Start: offset, InnerStart: end}
			for _, attr := range t.Attr {
				node.Attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			} else if root == nil {
				root = node
			}
			stack = append(stack, node)
		case xml.EndElement:
			node := stack[len(stack)-1]
			node.InnerEnd, node.End = offset, end
			node.SelfClosing = offset == end
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			}
		}
	}
	if root == nil {
		err = fmt.Errorf("config.xml is empty")
	}
	return
}

// escapeXMLText 는 Jenkins 가 저장하는 방식대로 텍스트를 이스케이프한다
func escapeXMLText(s string) string {
	return strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;", "\r", "&#xd;",
	).Replace(s)
}

// escapeXMLValue 는 한 줄로 넣을 값을 이스케이프한다. appendChild 가 줄마다 들여쓰기를 붙이므로 줄바꿈도 문자 참조로 바꾼다
func escapeXMLValue(s string) string {
	return strings.ReplaceAll(escapeXMLText(s), "\n", "&#xA;")
}

// scriptPlaceholder 는 appendChild 의 들여쓰기가 스크립트 본문에 붙지 않도록, 요소를 넣은 뒤에 본문으로 바꾸는 자리
const scriptPlaceholder = "\x00jcli-script\x00"

// lineIndent 는 pos 가 속한 줄의 들여쓰기
func lineIndent(doc string, pos int) string {
	lineStart := strings.LastIndex(doc[:pos], "\n") + 1
	line := doc[lineStart:pos]
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
}

// appendChild 는 parent 의 마지막 자식으로 elem 을 넣는다. elem 의 줄마다 자식 들여쓰기를 붙인다
func appendChild(doc string, parent *xmlNode, elem string) string {
	indent := lineIndent(doc, parent.Start)
	childIndent := indent + "  "
	body := childIndent + strings.ReplaceAll(elem, "\n", "\n"+childIndent)

	if parent.SelfClosing {
		tag := doc[parent.Start:parent.InnerStart]
		open := strings.TrimRight(strings.TrimSuffix(tag, "/>"), " ") + ">"
		return doc[:parent.Start] + open + "\n" + body + "\n" + indent + "</" + parent.Name + ">" + doc[parent.End:]
	}

	// 닫는 태그 앞의 줄바꿈·들여쓰기는 그대로 두고 그 앞에 넣는다
	inner := doc[parent.InnerStart:parent.InnerEnd]
	pos := parent.InnerStart + len(strings.TrimRight(inner, " \t\r\n"))
	if pos == parent.InnerStart && strings.TrimSpace(inner) == "" {
		return doc[:parent.InnerStart] + "\n" + body + "\n" + indent + doc[parent.InnerEnd:]
	}
	return doc[:pos] + "\n" + body + doc[pos:]
}

// removeNode 는 요소를 앞쪽 들여쓰기·줄바꿈과 함께 지운다
func removeNode(doc string, node *xmlNode) string {
	start := node.Start
	for start > 0 && (doc[start-1] == ' ' || doc[start-1] == '\t') {
		start--
	}
	if start > 0 && doc[start-1] == '\n' {
		start--
		if start > 0 && doc[start-1] == '\r' {
			start--
		}
	}
	return doc[:start] + doc[node.End:]
}

// replaceInner 는 요소의 내용을 바꾼다
func replaceInner(doc string, node *xmlNode, inner string) string {
	if node.SelfClosing {
		return doc[:node.Start] + "<" + node.Name + ">" + inner + "</" + node.Name + ">" + doc[node.End:]
	}
	return doc[:node.InnerStart] + inner + doc[node.InnerEnd:]
}

// pipelineDefinition 은 Pipeline job 의 스크립트 정의를 찾는다. 정의가 아직 없으면 nil
func pipelineDefinition(root *xmlNode) (definition *xmlNode, err error) {
	if root.Name != "flow-definition" {
		return nil, fmt.Errorf("only the pipeline job is supported, this job is %s", root.Name)
	}
	if definition = root.child("definition"); definition == nil {
		return
	}
	switch definition.Attrs["class"] {
	case cpsFlowDefinition:
	case cpsScmFlowDefinition:
		err = fmt.Errorf("this job reads the Jenkinsfile from SCM, edit it in the repository instead")
	default:
		err = fmt.Errorf("unsupported pipeline definition: %s", definition.Attrs["class"])
	}
	return
}

// pipelineScriptFromConfig 는 config.xml 에서 파이프라인 스크립트를 꺼낸다
func pipelineScriptFromConfig(config string) (pipeline *Pipeline, err error) {
	var root, definition *xmlNode
	if root, err = parseXMLNodes(config); err != nil {
		return
	}
	if definition, err = pipelineDefinition(root); err != nil {
		return
	}

	pipeline = &Pipeline{Sandbox: true}
	if definition == nil {
		return
	}
	if script := definition.child("script"); script != nil {
		pipeline.Script = script.Text
	}
	if sandbox := definition.child("sandbox"); sandbox != nil {
		pipeline.Sandbox = strings.TrimSpace(sandbox.Text) == "true"
	}
	return
}

// replacePipelineScript 는 config.xml 의 파이프라인 스크립트를 바꾼다
func replacePipelineScript(config, script string) (result string, err error) {
	var root, definition *xmlNode
	if root, err = parseXMLNodes(config); err != nil {
		return
	}
	if definition, err = pipelineDefinition(root); err != nil {
		return
	}

	escaped := escapeXMLText(script)
	if definition == nil {
		// 아직 스크립트를 넣은 적 없는 job. 샌드박스를 켜서 관리자 승인 없이 돌 수 있게 한다
		elem := fmt.Sprintf("<definition class=\"%s\">\n  <script>%s</script>\n  <sandbox>true</sandbox>\n</definition>",
			cpsFlowDefinition, scriptPlaceholder)
		result = appendChild(config, root, elem)
	} else if node := definition.child("script"); node != nil {
		return replaceInner(config, node, escaped), nil
	} else {
		result = appendChild(config, definition, "<script>"+scriptPlaceholder+"</script>")
	}
	return strings.Replace(result, scriptPlaceholder, escaped, 1), nil
}

// ParameterSpec 은 job param --add 로 받는 파라미터
type ParameterSpec struct {
	Name    string   `json:"name"`
	Value   string   `json:"value"`
	Desc    string   `json:"desc"`
	Type    string   `json:"type"`
	Choices []string `json:"choices"`
}

// ParseParameterSpecs 는 --add 의 JSON(배열 또는 객체 하나)을 읽는다
func ParseParameterSpecs(data string) (specs []ParameterSpec, err error) {
	data = strings.TrimSpace(data)
	if strings.HasPrefix(data, "{") {
		var spec ParameterSpec
		err = json.Unmarshal([]byte(data), &spec)
		specs = []ParameterSpec{spec}
	} else {
		err = json.Unmarshal([]byte(data), &specs)
	}
	if err != nil {
		err = fmt.Errorf("invalid parameters: %v", err)
	}
	return
}

// parameterXML 은 파라미터 정의 요소를 만든다
func (p ParameterSpec) parameterXML() (elem string, err error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return "", fmt.Errorf("parameter name is required")
	}
	lines := []string{
		"<name>" + escapeXMLValue(name) + "</name>",
		"<description>" + escapeXMLValue(p.Desc) + "</description>",
	}

	var class string
	switch strings.ToLower(strings.TrimSuffix(p.Type, "ParameterDefinition")) {
	case "", "string":
		class = "hudson.model.StringParameterDefinition"
		lines = append(lines, "<defaultValue>"+escapeXMLValue(p.Value)+"</defaultValue>", "<trim>false</trim>")
	case "text":
		class = "hudson.model.TextParameterDefinition"
		lines = append(lines, "<defaultValue>"+escapeXMLValue(p.Value)+"</defaultValue>", "<trim>false</trim>")
	case "boolean", "bool":
		class = "hudson.model.BooleanParameterDefinition"
		value := false
		if p.Value != "" {
			if value, err = strconv.ParseBool(p.Value); err != nil {
				return "", fmt.Errorf("parameter %s: value of boolean must be true or false", name)
			}
		}
		lines = append(lines, fmt.Sprintf("<defaultValue>%t</defaultValue>", value))
	case "choice":
		class = "hudson.model.ChoiceParameterDefinition"
		if len(p.Choices) == 0 {
			return "", fmt.Errorf("parameter %s: choices is required for choice", name)
		}
		lines = append(lines, "<choices class=\"java.util.Arrays$ArrayList\">", "  <a class=\"string-array\">")
		for _, choice := range p.Choices {
			lines = append(lines, "    <string>"+escapeXMLValue(choice)+"</string>")
		}
		lines = append(lines, "  </a>", "</choices>")
	default:
		return "", fmt.Errorf("parameter %s: unsupported type %s, use string, text, boolean or choice", name, p.Type)
	}

	elem = "<" + class + ">\n  " + strings.Join(lines, "\n  ") + "\n</" + class + ">"
	return
}

// parameterDefinitions 는 파라미터 정의 목록 요소와 그 상위 property 를 찾는다
func parameterDefinitions(root *xmlNode) (property, definitions *xmlNode) {
	properties := root.child("properties")
	if properties == nil {
		return
	}
	if property = properties.child(parametersProperty); property != nil {
		definitions = property.child("parameterDefinitions")
	}
	return
}

func parameterNode(definitions *xmlNode, name string) *xmlNode {
	if definitions == nil {
		return nil
	}
	for _, def := range definitions.Children {
		if n := def.child("name"); n != nil && strings.TrimSpace(n.Text) == name {
			return def
		}
	}
	return nil
}

// addParametersToConfig 는 config.xml 에 파라미터를 추가한다. 같은 이름이 이미 있으면 거부한다
func addParametersToConfig(config string, specs []ParameterSpec) (result string, err error) {
	result = config
	for _, spec := range specs {
		var elem string
		if elem, err = spec.parameterXML(); err != nil {
			return
		}

		var root *xmlNode
		if root, err = parseXMLNodes(result); err != nil {
			return
		}
		property, definitions := parameterDefinitions(root)
		if parameterNode(definitions, strings.TrimSpace(spec.Name)) != nil {
			return "", fmt.Errorf("parameter %s already exists", spec.Name)
		}

		switch {
		case definitions != nil:
			result = appendChild(result, definitions, elem)
		case property != nil:
			result = appendChild(result, property, "<parameterDefinitions>\n  "+
				strings.ReplaceAll(elem, "\n", "\n  ")+"\n</parameterDefinitions>")
		default:
			properties := root.child("properties")
			if properties == nil {
				return "", fmt.Errorf("cannot find the properties of the job in config.xml")
			}
			result = appendChild(result, properties, "<"+parametersProperty+">\n  <parameterDefinitions>\n    "+
				strings.ReplaceAll(elem, "\n", "\n    ")+"\n  </parameterDefinitions>\n</"+parametersProperty+">")
		}
	}
	return
}

// removeParametersFromConfig 는 config.xml 에서 파라미터를 지운다. 없는 이름이 있으면 거부한다.
// 파라미터가 하나도 남지 않으면 Jenkins 화면에서 끈 것처럼 파라미터 설정 자체를 지운다
func removeParametersFromConfig(config string, names []string) (result string, err error) {
	result = config
	for _, name := range names {
		var root *xmlNode
		if root, err = parseXMLNodes(result); err != nil {
			return
		}
		property, definitions := parameterDefinitions(root)
		node := parameterNode(definitions, name)
		if node == nil {
			return "", fmt.Errorf("parameter %s does not exist", name)
		}

		if len(definitions.Children) == 1 {
			result = removeNode(result, property)
		} else {
			result = removeNode(result, node)
		}
	}
	return
}

// splitParameterNames 는 "a, b,c" 형식의 이름 목록을 나눈다
func splitParameterNames(names string) (result []string) {
	for _, name := range strings.Split(names, ",") {
		if name = strings.TrimSpace(name); name != "" {
			result = append(result, name)
		}
	}
	return
}
