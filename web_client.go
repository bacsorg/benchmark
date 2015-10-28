package benchmark

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	authCookie = ".ASPXAUTH"
)

type WebClient struct {
	bacsBaseUrl string
	httpClient  *http.Client
	contestId   int
}

func NewWebClient(bacsBaseUrl string) (*WebClient, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &WebClient{
		bacsBaseUrl: bacsBaseUrl,
		httpClient: &http.Client{
			Jar: cookieJar,
		},
	}, nil
}

func (c *WebClient) URL(relative string) string {
	return c.bacsBaseUrl + relative
}

func (c *WebClient) URLf(format string, param ...interface{}) string {
	return c.URL(fmt.Sprintf(format, param...))
}

func (c *WebClient) Login(username, password string) error {
	resp, err := c.httpClient.Get(c.URL("/Account/LogOn"))
	if err != nil {
		return err
	}
	resp.Body.Close()
	resp, err = c.httpClient.PostForm(c.URL("/Account/LogOn"),
		url.Values{
			"Login":      {username},
			"Password":   {password},
			"RememberMe": {"false"},
			"logon":      {"Вход"},
		})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// check that login succeeded
	url, err := url.Parse(c.bacsBaseUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode/100 == 5 {
		return fmt.Errorf("unable to login: %d %q", resp.StatusCode, resp.Status)
	}
	for _, cookie := range c.httpClient.Jar.Cookies(url) {
		if cookie.Name == authCookie {
			return nil
		}
	}
	return fmt.Errorf("unable to login: %q cookie not found", authCookie)
}

func (c *WebClient) EnterContest(contestId int) error {
	resp, err := c.httpClient.Get(
		c.URLf("/Contest/EnterContest?contestID=%d", contestId))
	if err != nil {
		return err
	}
	resp.Body.Close()
	c.contestId = contestId
	return nil
}

func (c *WebClient) monitor(name string, params ...string) (string, error) {
	url := c.URLf("/Monitor/%s?contestId=%d", name, c.contestId)
	for _, param := range params {
		url += "&" + param
	}
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to read monitor: %v", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *WebClient) AcmMonitor() (string, error) {
	return c.monitor("AcmMonitor", "showLeftMenu=True")
}

func (c *WebClient) SchoolFinalMonitor() (string, error) {
	return c.monitor("SchoolFinalMonitor")
}

func (c *WebClient) MySchoolFinalSubmits() (string, error) {
	return c.monitor("MySchoolFinalSubmits")
}

func (c *WebClient) Submit(problem, compiler, solution string) error {
	resp, err := c.httpClient.PostForm(
		c.URLf("/Contest/Submit?contestId=%d", c.contestId),
		url.Values{
			"Problem":          {problem},
			"Compiler":         {CompilerId(compiler)},
			"SolutionText":     {solution},
			"SolutionFileType": {"Text"},
		})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
