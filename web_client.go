package benchmark

import (
    "net/http"
    "net/http/cookiejar"
    "net/url"
)

type WebClient struct {
    bacsBaseUrl string
    httpClient  *http.Client
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
