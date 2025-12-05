const gethttpOrigin = () => {
    return window.location.origin
}

// ==================== 公共加密工具 ====================

class TokenEncryption {
    // 根据字符串内容生成确定性 salt（范围 0~255）
    static getDeterministicSalt(text) {
        let sum = 0;
        for (let i = 0; i < text.length; i++) {
            sum = (sum + text.charCodeAt(i) * (i + 1)) & 0xFF;
        }
        return sum;
    }

    // 加密：首字节为salt，后续为按位异或后的数据
    static encryptHex(text, key) {
        const salt = TokenEncryption.getDeterministicSalt(text);
        let result = salt.toString(16).padStart(2, '0');
        for (let i = 0; i < text.length; i++) {
            const code = text.charCodeAt(i) ^ (key & 0xFF) ^ ((salt + i) & 0xFF);
            result += code.toString(16).padStart(2, '0');
        }
        return result;
    }
}

// ==================== 公共代码模板生成器 ====================

class CodeTemplates {
    static getCurl(url, dataStr) {
        return `curl -X POST --location '${url}' \\
--header 'Content-Type: application/json' \\
--data '${dataStr}'`;
    }

    static getGolang(url, dataStr) {
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
    }

    static getPython(url, dataStr) {
        return `import requests

headers = {
    'Content-Type': 'application/json',
}
json_data = ${dataStr}
response = requests.post('${url}', headers=headers, json=json_data)

print("response:", response.json())`;
    }

    static getJava(url, dataStr) {
        return `import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;

HttpClient client = HttpClient.newBuilder()
    .followRedirects(HttpClient.Redirect.NORMAL)
    .build();

HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${url}"))
    .POST(BodyPublishers.ofString(${JSON.stringify(dataStr).trim('"')}))
    .setHeader("Content-Type", "application/json")
    .build();

HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());`;
    }

    static getRust(url, dataStr) {
        return `extern crate reqwest;
use reqwest::header;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut headers = header::HeaderMap::new();
    headers.insert("Content-Type", "application/json".parse().unwrap());

    let client = reqwest::blocking::Client::new();
    let res = client.post("${url}")
        .headers(headers)
        .body(r#"
${dataStr}
"#
        )
        .send()?
        .text()?;
    println!("{}", res);

    Ok(())
}`;
    }

    static getPHP(url, dataStr) {
        return `<?php
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, '${url}');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
]);
curl_setopt($ch, CURLOPT_POSTFIELDS, ${JSON.stringify(dataStr).trim('"')});
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);

$response = curl_exec($ch);

curl_close($ch);`;
    }

    static getNode(url, dataStr) {
        return `var request = require('request');

var headers = {
    'Content-Type': 'application/json'
};

var dataString = ${JSON.stringify(dataStr).trim('"')};

var options = {
    url: '${url}',
    method: 'POST',
    headers: headers,
    body: dataString
};

function callback(error, response, body) {
    if (!error && response.statusCode == 200) {
        console.log(body);
    }
}

request(options, callback);`;
    }
}

// ==================== 发信任务 API (V1) ====================

class ApiStrGenerate {
    static getDataString(task_id, options) {
        let data = { token: TokenEncryption.encryptHex(task_id, 71) };
        data.title = 'message title';
        data.text = 'Hello World!';
        if (options.html) data.html = '<h1> Hello World! </h1>';
        if (options.markdown) data.markdown = '**Hello World!**';
        if (options.url) data.url = 'https://github.com';
        if (options.at_mobiles) data.at_mobiles = ['13800138000', '13900139000'];
        if (options.at_user_ids) data.at_user_ids = ['zhangsan', 'lisi'];
        if (options.at_all) data.at_all = true;
        return JSON.stringify(data, null, 4);
    }

    static getApiUrl() {
        return `${gethttpOrigin()}/api/v1/message/send`;
    }

    static getCurlString(task_id, options) {
        return CodeTemplates.getCurl(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getGolangString(task_id, options) {
        return CodeTemplates.getGolang(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getPythonString(task_id, options) {
        return CodeTemplates.getPython(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getJaveString(task_id, options) {
        return CodeTemplates.getJava(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getRustString(task_id, options) {
        return CodeTemplates.getRust(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getPHPString(task_id, options) {
        return CodeTemplates.getPHP(this.getApiUrl(), this.getDataString(task_id, options));
    }

    static getNodeString(task_id, options) {
        return CodeTemplates.getNode(this.getApiUrl(), this.getDataString(task_id, options));
    }
}

// ==================== 模板 API (V2) ====================

class TemplateApiStrGenerate {
    static getTemplateDataString(template_id, placeholders_json) {
        // 解析占位符配置
        let placeholders = {};
        try {
            const placeholdersList = JSON.parse(placeholders_json || '[]');
            // 根据占位符配置生成示例值
            placeholdersList.forEach(p => {
                placeholders[p.key] = p.default || `mock_${p.key}`;
            });
        } catch (e) {
            // 如果解析失败，使用默认示例
            placeholders = {
                'username': 'John Doe',
                'email': 'john@example.com',
                'phone': '13800138000'
            };
        }

        let data = {
            token: TokenEncryption.encryptHex(template_id, 71),
            title: 'message title',
            placeholders: placeholders
        };
        return JSON.stringify(data, null, 4);
    }

    static getApiUrl() {
        return `${gethttpOrigin()}/api/v2/message/send`;
    }

    static getCurlString(template_id, placeholders_json) {
        return CodeTemplates.getCurl(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getGolangString(template_id, placeholders_json) {
        return CodeTemplates.getGolang(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getPythonString(template_id, placeholders_json) {
        return CodeTemplates.getPython(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getJavaString(template_id, placeholders_json) {
        return CodeTemplates.getJava(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getRustString(template_id, placeholders_json) {
        return CodeTemplates.getRust(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getPHPString(template_id, placeholders_json) {
        return CodeTemplates.getPHP(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }

    static getNodeString(template_id, placeholders_json) {
        return CodeTemplates.getNode(this.getApiUrl(), this.getTemplateDataString(template_id, placeholders_json));
    }
}

export { ApiStrGenerate, TemplateApiStrGenerate };
