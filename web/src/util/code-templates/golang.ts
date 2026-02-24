const getGolangFunction = (url: string, dataStr: string) => {
    return `package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
)

func SendMessage(payload string) (string, error) {
    url := "${url}"
    client := &http.Client{}
    data := strings.NewReader(payload)
    req, err := http.NewRequest("POST", url, data)
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    bodyText, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(bodyText), nil
}

func main() {
    var data = \`${dataStr}\`
    res, err := SendMessage(data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\\n", res)
}`;
};

const getGolangScript = (url: string, dataStr: string) => {
    return `package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
)

func main() {
    client := &http.Client{}
    var data = strings.NewReader(\`${dataStr}\`)
    req, err := http.NewRequest("POST", "${url}", data)
    if err != nil {
        log.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    bodyText, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\\n", bodyText)
}`;
};

export const getGolang = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getGolangFunction(url, dataStr) : getGolangScript(url, dataStr);
};
