package benchmark

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/http/cookiejar"
    "net/url"
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
    resp.Body.Close()
    return nil
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

func (c *WebClient) AcmMonitor() (string, error) {
    resp, err := c.httpClient.Get(
        c.URLf("/Monitor/AcmMonitor?contestId=%d&showLeftMenu=True", c.contestId))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("Unable to read monitor: %v", resp.Status)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}
