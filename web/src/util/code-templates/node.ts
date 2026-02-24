const getNodeFunction = (url: string, dataStr: string) => {
    return `var request = require('request');

function sendMessage(payload) {
    var options = {
        url: '${url}',
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload
    };

    return new Promise((resolve, reject) => {
        request(options, function (error, response, body) {
            if (error) reject(error);
            else resolve(body);
        });
    });
}

// 调用示例
var dataString = ${JSON.stringify(dataStr).slice(1, -1)};
sendMessage(dataString)
    .then(body => console.log(body))
    .catch(err => console.error(err));`;
};

const getNodeScript = (url: string, dataStr: string) => {
    return `var request = require('request');

var headers = {
    'Content-Type': 'application/json'
};

var dataString = ${JSON.stringify(dataStr).slice(1, -1)};

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
};

export const getNode = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getNodeFunction(url, dataStr) : getNodeScript(url, dataStr);
};
