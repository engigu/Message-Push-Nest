const getCurlFunction = (url: string, dataStr: string) => {
    return `send_message() {
  local url='${url}'
  curl -X POST --location "$url" \\
  --header 'Content-Type: application/json' \\
  --data "$1"
}

# 调用示例
send_message '${dataStr}'`;
};

const getCurlScript = (url: string, dataStr: string) => {
    return `curl -X POST --location '${url}' \\
--header 'Content-Type: application/json' \\
--data '${dataStr}'`;
};

export const getCurl = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getCurlFunction(url, dataStr) : getCurlScript(url, dataStr);
};
