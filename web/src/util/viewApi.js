

const gethttpOrigin = () => {
    return window.location.origin
}

class ApiStrGenerate {

    static getDataString(task_id, options) {
        let data = { task_id: task_id };
        data.title = 'message title';
        data.text = 'Hello World!';
        if (options.html) {
            data.html = '<h1> Hello World! </h1>';
        }
        if (options.markdown) {
            data.markdown = '**Hello World!**';
        }
        if (options.url) {
            data.url = 'https://github.com';
        }
        let dataStr = JSON.stringify(data, null, 4);
        return dataStr
    }

    static getCurlString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `curl -X POST --location '${gethttpOrigin()}/api/v1/message/send' \\
        --header 'Content-Type: application/json' \\
        --data '${dataStr}'`;
        return example;
    }

    static getGolangString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `package main

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
    req, err := http.NewRequest("POST", "${gethttpOrigin()}/api/v1/message/send", data)
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
}
        `;
        return example;
    }

    static getPythonString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `import requests

headers = {
    'Content-Type': 'application/json',
}
json_data = ${dataStr}
response = requests.post('${gethttpOrigin()}/api/v1/message/send', headers=headers, json=json_data)

print("response:", response.json())
`;
        return example;
    }

    static getJaveString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;

HttpClient client = HttpClient.newBuilder()
    .followRedirects(HttpClient.Redirect.NORMAL)
    .build();

HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${gethttpOrigin()}/api/v1/message/send"))
    .POST(BodyPublishers.ofString(${JSON.stringify(dataStr).trim('\"')}))
    .setHeader("Content-Type", "application/json")
    .build();

HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
`;
        return example;
    }

    static getRustString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `extern crate reqwest;
use reqwest::header;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut headers = header::HeaderMap::new();
    headers.insert("Content-Type", "application/json".parse().unwrap());

    let client = reqwest::blocking::Client::new();
    let res = client.post("${gethttpOrigin()}/api/v1/message/send")
        .headers(headers)
        .body(r#"
${dataStr}
"#
        )
        .send()?
        .text()?;
    println!("{}", res);

    Ok(())
}
`;
        return example;
    }

    static getPHPString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `<?php
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, '${gethttpOrigin()}/api/v1/message/send');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
]);
curl_setopt($ch, CURLOPT_POSTFIELDS, ${JSON.stringify(dataStr).trim('\"')});
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);

$response = curl_exec($ch);

curl_close($ch);
        `;
        return example;
    }

    static getNodeString(task_id, options) {
        let dataStr = ApiStrGenerate.getDataString(task_id, options);
        let example = `var request = require('request');

var headers = {
    'Content-Type': 'application/json'
};

var dataString = ${JSON.stringify(dataStr).trim('\"')};

var options = {
    url: '${gethttpOrigin()}/api/v1/message/send',
    method: 'POST',
    headers: headers,
    body: dataString
};

function callback(error, response, body) {
    if (!error && response.statusCode == 200) {
    
    }
}

request(options, callback);
        `;
        return example;
    }

}

export { ApiStrGenerate };
