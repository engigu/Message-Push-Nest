

const gethttpOrigin = () => {
    return window.location.origin
}

class ApiStrGenerate {


    static getCurlString(task_id, options) {
        let data = { task_id: task_id };
        data.text = 'Hello World!';
        if (options.html) {
            data.html = '<h1> Hello World! </h1>';
        }
        if (options.markdown) {
            data.html = '** Hello World! **';
        }
        let dataStr = JSON.stringify(data, null, 4)
        let example = `curl -X POST --location '${gethttpOrigin()}/api/v1/message/send' \\
        --header 'Accept: application/json'  \\
        --data '${dataStr}'`;
        return example;
    }

}

export { ApiStrGenerate };
