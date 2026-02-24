const getPHPFunction = (url: string, dataStr: string) => {
    return `<?php
function sendMessage($payload) {
    $url = '${url}';
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Content-Type: application/json',
    ]);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $payload);
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);

    $response = curl_exec($ch);
    curl_close($ch);
    return $response;
}

// 调用示例
$data = ${JSON.stringify(dataStr).slice(1, -1)};
echo sendMessage($data);`;
};

const getPHPScript = (url: string, dataStr: string) => {
    return `<?php
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, '${url}');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
]);
curl_setopt($ch, CURLOPT_POSTFIELDS, ${JSON.stringify(dataStr).slice(1, -1)});
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);

$response = curl_exec($ch);

curl_close($ch);`;
};

export const getPHP = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getPHPFunction(url, dataStr) : getPHPScript(url, dataStr);
};
