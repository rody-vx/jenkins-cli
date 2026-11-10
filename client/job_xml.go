package client

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// config.xml 을 고칠 때 바꾸려는 요소만 문자열로 교체하고 나머지는 원문 그대로 둔다.
// encoding/xml 로 통째로 다시 쓰면 플러그인 속성·순서·서식이 달라질 수 있기 때문이다.

const (
	cpsFlowDefinition    = "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition"
	cpsScmFlowDefinition = "org.jenkinsci.plugins.workflow.cps.CpsScmFlowDefinition"
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
