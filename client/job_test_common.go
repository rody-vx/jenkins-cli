package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	httpdownloader "github.com/linuxsuren/http-downloader/pkg"
	"github.com/rody-vx/jenkins-cli/mock/mhttp"
)

// PrepareForGetJobInputActions only for test
func PrepareForGetJobInputActions(roundTripper *mhttp.MockRoundTripper, rootURL, user, password, jobName string, buildID int) (
	request *http.Request, response *http.Response) {
	request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/job/%s/%d/wfapi/pendingInputActions", rootURL, jobName, buildID), nil)
	response = &http.Response{
		StatusCode: 200,
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(`
[{"id":"Eff7d5dba32b4da32d9a67a519434d3f","proceedText":"继续","message":"message","inputs":[],
"proceedUrl":"/job/test/5/wfapi/inputSubmit?inputId=Eff7d5dba32b4da32d9a67a519434d3f",
"abortUrl":"/job/test/5/input/Eff7d5dba32b4da32d9a67a519434d3f/abort","redirectApprovalUrl":"/job/test/5/input/"}]`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)

	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	}
	return
}

// PrepareForSubmitInput only for test
func PrepareForSubmitInput(roundTripper *mhttp.MockRoundTripper, rootURL, jobPath, user, password string) (
	request *http.Request, response *http.Response) {
	request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s/%d/input/%s/abort?json={\"parameter\":[]}", rootURL, jobPath, 1, "Eff7d5dba32b4da32d9a67a519434d3f"), nil)
	PrepareCommonPost(request, "", roundTripper, user, password, rootURL)
	return
}

// PrepareForSubmitProcessInput only for test
func PrepareForSubmitProcessInput(roundTripper *mhttp.MockRoundTripper, rootURL, jobPath, user, password string) (
	request *http.Request, response *http.Response) {
	request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s/%d/input/%s/proceed?json={\"parameter\":[]}", rootURL, jobPath, 1, "Eff7d5dba32b4da32d9a67a519434d3f"), nil)
	PrepareCommonPost(request, "", roundTripper, user, password, rootURL)
	return
}

// PrepareForBuildWithNoParams only for test
func PrepareForBuildWithNoParams(roundTripper *mhttp.MockRoundTripper, rootURL, jobName, user, password string) (
	request *http.Request, response *http.Response) {
	formData := url.Values{"json": {`{"parameter": []}`}}
	payload := strings.NewReader(formData.Encode())
	request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/job/%s/build", rootURL, jobName), payload)
	request.Header.Add(httpdownloader.ContentType, httpdownloader.ApplicationForm)
	response = PrepareCommonPost(request, "", roundTripper, user, password, rootURL)
	response.StatusCode = 201
	return
}

// PrepareForBuildWithParams only for test
func PrepareForBuildWithParams(roundTripper *mhttp.MockRoundTripper, rootURL, jobName, user, password string) (
	request *http.Request, response *http.Response) {
	formData := url.Values{"json": {`{"parameter": {"Description":"","name":"name","Type":"StringParameterDefinition","value":"value","DefaultParameterValue":{"Description":"","Value":null}}}`}}
	payload := strings.NewReader(formData.Encode())
	request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/job/%s/build", rootURL, jobName), payload)
	request.Header.Add(httpdownloader.ContentType, httpdownloader.ApplicationForm)
	response = PrepareCommonPost(request, "", roundTripper, user, password, rootURL)
	response.StatusCode = 201
	return
}

// PrepareForGetJob only for test
func PrepareForGetJob(roundTripper *mhttp.MockRoundTripper, rootURL, jobName, user, password string) (
	response *http.Response) {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/job/%s/api/json", rootURL, jobName), nil)
	response = &http.Response{
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{
  "name" : "%s",
  "builds" : [
    {
      "number" : 1,
      "url" : "http://localhost:8080/job/we/1/"
    },
    {
      "number" : 2,
      "url" : "http://localhost:8080/job/we/2/"
    }]
				}`, jobName))),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	}
	return
}

// PrepareForGetJobWithParams only for test
func PrepareForGetJobWithParams(roundTripper *mhttp.MockRoundTripper, rootURL, jobName, user, password string) {
	response := PrepareForGetJob(roundTripper, rootURL, jobName, user, password)
	response.Body = ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{
  "name" : "%s",
  "builds" : [
    {
      "number" : 1,
      "url" : "http://localhost:8080/job/we/1/"
    },
    {
      "number" : 2,
      "url" : "http://localhost:8080/job/we/2/"
    }],
  "property" : [
    {
      "_class" : "io.alauda.jenkins.devops.sync.WorkflowJobProperty"
    },
    {
      "parameterDefinitions" : [
        {
          "defaultParameterValue" : {
            "name" : "name",
            "value" : "jake"
          },
          "description" : "",
          "name" : "name",
          "type" : "StringParameterDefinition"
        }
      ]
    }
  ]
}`, jobName)))
}

// PrepareForGetBuild only for test
func PrepareForGetBuild(roundTripper *mhttp.MockRoundTripper, rootURL, jobName string, buildID int, user, password string) {
	api := ""
	if buildID == -1 {
		api = fmt.Sprintf("%s/job/%s/lastBuild/api/json", rootURL, jobName)
	} else {
		api = fmt.Sprintf("%s/job/%s/%d/api/json", rootURL, jobName, buildID)
	}
	request, _ := http.NewRequest(http.MethodGet, api, nil)
	response := &http.Response{
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(`
				{"displayName":"fake"}
				`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	}
}

// PrepareForJobLog only for test
func PrepareForJobLog(roundTripper *mhttp.MockRoundTripper, rootURL, jobName string, buildID int, user, password string) {
	var api string
	if buildID == -1 {
		api = fmt.Sprintf("%s/job/%s/lastBuild/logText/progressiveText?start=%d", rootURL, jobName, 0)
	} else {
		api = fmt.Sprintf("%s/job/%s/%d/logText/progressiveText?start=%d", rootURL, jobName, buildID, 0)
	}
	request, _ := http.NewRequest(http.MethodGet, api, nil)
	response := &http.Response{
		StatusCode: 200,
		Request:    request,
		Header: map[string][]string{
			"X-More-Data": {"false"},
			"X-Text-Size": {"8"},
		},
		Body: ioutil.NopCloser(bytes.NewBufferString("fake log")),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	}
}

// PrepareOneItem only for test
func PrepareOneItem(roundTripper *mhttp.MockRoundTripper, rootURL, name, kind, user, token string) {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/items/list?name=%s&type=%s&start=%d&limit=%d&parent=",
		rootURL, name, kind, 0, 50), nil)
	response := &http.Response{
		StatusCode: 200,
		Request:    request,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`[{"name":"fake","displayName":"fake","description":null,"type":"WorkflowJob","shortURL":"job/fake/","url":"job/fake/"}]`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && token != "" {
		request.SetBasicAuth(user, token)
	}
}

// PrepareForSearchViaAPI only for test
func PrepareForSearchViaAPI(roundTripper *mhttp.MockRoundTripper, rootURL, user, token string) {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/json", rootURL), nil)
	response := &http.Response{
		StatusCode: 200,
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"jobs":[{"_class":"com.cloudbees.hudson.plugins.folder.Folder","name":"folder","displayName":"folder",` +
			`"jobs":[{"_class":"org.jenkinsci.plugins.workflow.job.WorkflowJob","name":"deploy","displayName":"deploy","buildable":true},` +
			`{"_class":"hudson.model.FreeStyleProject","name":"other","displayName":"other","buildable":true}]}]}`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && token != "" {
		request.SetBasicAuth(user, token)
	}
}

// PrepareForParameterChoices only for test
func PrepareForParameterChoices(roundTripper *mhttp.MockRoundTripper, rootURL, jobPath string) {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s/api/json", rootURL, jobPath), nil)
	response := &http.Response{
		StatusCode: 200,
		Request:    request,
		Body: ioutil.NopCloser(bytes.NewBufferString(`{"property":[{},{"parameterDefinitions":[` +
			`{"name":"COUNTRY","type":"ChoiceParameterDefinition","choices":["KR","US"]},` +
			`{"name":"BRANCH_NAME","type":"GitParameterDefinition"},` +
			`{"name":"MEMO","type":"StringParameterDefinition"}]}]}`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)

	requestCrumb, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/crumbIssuer/api/json", rootURL), nil)
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(requestCrumb)).Return(&http.Response{
		StatusCode: 200,
		Request:    requestCrumb,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"crumbRequestField":"CrumbRequestField","crumb":"Crumb"}`)),
	}, nil)

	requestGit, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s/descriptorByName/%s/fillValueItems",
		rootURL, jobPath, gitParameterDescriptor), nil)
	requestGit.Header.Add("CrumbRequestField", "Crumb")
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(requestGit)).Return(&http.Response{
		StatusCode: 200,
		Request:    requestGit,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"values":[{"name":"main","value":"main"},{"name":"v1","value":"v1"}]}`)),
	}, nil)
}

// PrepareEmptyItems only for test
func PrepareEmptyItems(roundTripper *mhttp.MockRoundTripper, rootURL, name, kind, user, token string) {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/items/list?name=%s&type=%s&start=%d&limit=%d&parent=",
		rootURL, name, kind, 0, 50), nil)
	response := &http.Response{
		StatusCode: 200,
		Request:    request,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
	}
	roundTripper.EXPECT().
		RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	if user != "" && token != "" {
		request.SetBasicAuth(user, token)
	}
}

// PrepareForDisableJob only for test
func PrepareForDisableJob(roundTripper *mhttp.MockRoundTripper, rootURL, name, user, token string) {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/job/%s/disable", rootURL, name), nil)
	PrepareCommonPost(request, "", roundTripper, user, token, rootURL)
	//response := &http.Response{
	//	StatusCode: 200,
	//	Request:    request,
	//	Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
	//}
	//roundTripper.EXPECT().
	//	RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	//if user != "" && token != "" {
	//	request.SetBasicAuth(user, token)
	//}
}

// PrepareForEnableJob only for test
func PrepareForEnableJob(roundTripper *mhttp.MockRoundTripper, rootURL, name, user, token string) {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/job/%s/enable", rootURL, name), nil)
	PrepareCommonPost(request, "", roundTripper, user, token, rootURL)
	//response := &http.Response{
	//	StatusCode: 200,
	//	Request:    request,
	//	Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
	//}
	//roundTripper.EXPECT().
	//	RoundTrip(NewRequestMatcher(request)).Return(response, nil)
	//if user != "" && token != "" {
	//	request.SetBasicAuth(user, token)
	//}
}
