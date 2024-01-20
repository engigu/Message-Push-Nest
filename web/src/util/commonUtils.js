

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
  
  static getCurrentTimeStr = () => {
    const currentDate = new Date();
    const year = currentDate.getFullYear();
    const month = String(currentDate.getMonth() + 1).padStart(2, '0');
    const day = String(currentDate.getDate()).padStart(2, '0');
    const hours = String(currentDate.getHours()).padStart(2, '0');
    const minutes = String(currentDate.getMinutes()).padStart(2, '0');
    const seconds = String(currentDate.getSeconds()).padStart(2, '0');
    const formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    return formattedDate;
  }

}

export { CommonUtils };
