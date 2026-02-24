const getRustFunction = (url: string, dataStr: string) => {
    return `extern crate reqwest;
use reqwest::header;

fn send_message(payload: &str) -> Result<String, Box<dyn std::error::Error>> {
    let url = "${url}";
    let mut headers = header::HeaderMap::new();
    headers.insert("Content-Type", "application/json".parse().unwrap());

    let client = reqwest::blocking::Client::new();
    let res = client.post(url)
        .headers(headers)
        .body(payload.to_string())
        .send()?
        .text()?;
    Ok(res)
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let payload = r#"
${dataStr}
"#;
    println!("{}", send_message(payload)?);
    Ok(())
}`;
};

const getRustScript = (url: string, dataStr: string) => {
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
};

export const getRust = (url: string, dataStr: string, isFunction: boolean = false) => {
    return isFunction ? getRustFunction(url, dataStr) : getRustScript(url, dataStr);
};
