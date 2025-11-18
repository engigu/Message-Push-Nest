# 调用示例

本页面提供各种编程语言的API调用示例。

## CURL

```bash
curl -X POST --location 'http://127.0.0.1:8000/api/v1/message/send' \
  --header 'Content-Type: application/json' \
  --data '{
    "token": "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    "title": "message title",
    "text": "Hello World!"
  }'
```

## Python

```python
import requests

headers = {
    'Content-Type': 'application/json',
}

json_data = {
    "token": "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    "title": "message title",
    "text": "Hello World!"
}

response = requests.post(
    'http://127.0.0.1:8000/api/v1/message/send',
    headers=headers,
    json=json_data
)

print("response:", response.json())
```

### 使用 requests 库

首先安装依赖：

```bash
pip install requests
```

## Go

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	client := &http.Client{}
	var data = strings.NewReader(`{
    "token": "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    "title": "message title",
    "text": "Hello World!"
}`)
	req, err := http.NewRequest("POST", "http://127.0.0.1:8000/api/v1/message/send", data)
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
	fmt.Printf("%s\n", bodyText)
}
```

## Java

```java
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;

public class MessageNestExample {
    public static void main(String[] args) throws IOException, InterruptedException {
        HttpClient client = HttpClient.newBuilder()
            .followRedirects(HttpClient.Redirect.NORMAL)
            .build();

        String jsonData = """
            {
                "token": "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
                "title": "message title",
                "text": "Hello World!"
            }
            """;

        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create("http://127.0.0.1:8000/api/v1/message/send"))
            .POST(BodyPublishers.ofString(jsonData))
            .setHeader("Content-Type", "application/json")
            .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        
        System.out.println(response.body());
    }
}
```

## Node.js

### 使用 request 库

```javascript
var request = require('request');

var headers = {
    'Content-Type': 'application/json'
};

var dataString = JSON.stringify({
    "token": "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    "title": "message title",
    "text": "Hello World!"
});

var options = {
    url: 'http://127.0.0.1:8000/api/v1/message/send',
    method: 'POST',
    headers: headers,
    body: dataString
};

function callback(error, response, body) {
    if (!error && response.statusCode == 200) {
        console.log(body);
    }
}

request(options, callback);
```

### 使用 axios 库

```javascript
const axios = require('axios');

const data = {
    token: "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    title: "message title",
    text: "Hello World!"
};

axios.post('http://127.0.0.1:8000/api/v1/message/send', data, {
    headers: {
        'Content-Type': 'application/json'
    }
})
.then(response => {
    console.log('response:', response.data);
})
.catch(error => {
    console.error('error:', error);
});
```

### 使用 fetch (Node.js 18+)

```javascript
const data = {
    token: "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    title: "message title",
    text: "Hello World!"
};

fetch('http://127.0.0.1:8000/api/v1/message/send', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(data)
})
.then(response => response.json())
.then(data => {
    console.log('response:', data);
})
.catch(error => {
    console.error('error:', error);
});
```

## PHP

```php
<?php
$ch = curl_init();

$data = array(
    "token" => "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
    "title" => "message title",
    "text" => "Hello World!"
);

curl_setopt($ch, CURLOPT_URL, 'http://127.0.0.1:8000/api/v1/message/send');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
]);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);

$response = curl_exec($ch);

if (curl_errno($ch)) {
    echo 'Error:' . curl_error($ch);
} else {
    echo $response;
}

curl_close($ch);
?>
```

## C#

```csharp
using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

class Program
{
    static async Task Main(string[] args)
    {
        using var client = new HttpClient();
        
        var data = new
        {
            token = "a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e",
            title = "message title",
            text = "Hello World!"
        };
        
        var json = JsonSerializer.Serialize(data);
        var content = new StringContent(json, Encoding.UTF8, "application/json");
        
        var response = await client.PostAsync(
            "http://127.0.0.1:8000/api/v1/message/send",
            content
        );
        
        var responseString = await response.Content.ReadAsStringAsync();
        Console.WriteLine(responseString);
    }
}
```

## Ruby

```ruby
require 'net/http'
require 'json'
require 'uri'

uri = URI('http://127.0.0.1:8000/api/v1/message/send')
http = Net::HTTP.new(uri.host, uri.port)

request = Net::HTTP::Post.new(uri.path, {
  'Content-Type' => 'application/json'
})

request.body = {
  token: 'a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e',
  title: 'message title',
  text: 'Hello World!'
}.to_json

response = http.request(request)
puts response.body
```

## 注意事项

::: tip 提示
- 将示例中的 `http://127.0.0.1:8000` 替换为你的实际服务地址
- 将 `a3541c2f0d3e1b4a5c6d7e8f9a0b1c2d3e` 替换为你在管理后台创建的实际 Token
- 建议在生产环境中使用 HTTPS
:::
