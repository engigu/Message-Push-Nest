

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

}

export { CommonUtils };
