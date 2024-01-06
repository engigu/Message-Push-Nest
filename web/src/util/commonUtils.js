


class CommonUtils {


    // static getCurlString(task_id, options) {
    //     let data = { task_id: task_id };
    //     data.title = 'message title';
    //     data.text = 'Hello World!';
    //     if (options.html) {
    //         data.html = '<h1> Hello World! </h1>';
    //     }
    //     if (options.markdown) {
    //         data.html = '**Hello World!**';
    //     }
    //     let dataStr = JSON.stringify(data, null, 4);
    //     let example = `curl -X POST --location '${gethttpOrigin()}/api/v1/message/send' \\
    //     --header 'Content-Type: application/json' \\
    //     --data '${dataStr}'`;
    //     return example;
    // }

    static formatInsConfigDisplay = (scope) => {
        if (!scope.row.config) {
          return ""
        }
        if (scope.row.way_type == "Email") {
          let config = JSON.parse(scope.row.config)
          let info = `发送账号：${config.to_account}`
          return info
        } else {
          return "暂无"
        }
      }

}

export { CommonUtils };
