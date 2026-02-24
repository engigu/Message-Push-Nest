const getPythonFunction = (url: string, dataStr: string) => {
    return `import requests

def send_message(payload):
    url = '${url}'
    headers = {
        'Content-Type': 'application/json',
    }
    response = requests.post(url, headers=headers, json=payload)
    return response.json()

# 调用示例
json_data = ${dataStr}
print("response:", send_message(json_data))`;
};

const getPythonScript = (url: string, dataStr: string) => {
    return `import requests

headers = {
    'Content-Type': 'application/json',
}
json_data = ${dataStr}
response = requests.post('${url}', headers=headers, json=json_data)

print("response:", response.json())`;
};

export const getPython = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getPythonFunction(url, dataStr) : getPythonScript(url, dataStr);
};
