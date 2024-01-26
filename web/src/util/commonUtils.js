

import { CONSTANT } from '@/constant'

class CommonUtils {

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

  static formatWayName = (type) => {
    return CONSTANT.WAYS_DATA_MAP[type].label;
  }

  // 获取东八区时间字符串 
  static getCurrentTimeStr = () => {
    let time = new Date();
    const timezoneOffset = 8;
    const utc = time.getTime() + time.getTimezoneOffset() * 60000;
    const chinaDate = new Date(utc + timezoneOffset * 60 * 60 * 1000);

    const year = chinaDate.getFullYear();
    const month = String(chinaDate.getMonth() + 1).padStart(2, '0');
    const day = String(chinaDate.getDate()).padStart(2, '0');
    const hour = String(chinaDate.getHours()).padStart(2, '0');
    const minute = String(chinaDate.getMinutes()).padStart(2, '0');
    const second = String(chinaDate.getSeconds()).padStart(2, '0');

    const formattedDateTime = `${year}-${month}-${day} ${hour}:${minute}:${second}`;
    return formattedDateTime;
  }

}

export { CommonUtils };
