// 类型声明文件
export interface ConstantType {
  PAGE: number;
  TOTAL: number;
  DEFALUT_SITE_CONFIG: string;
  LOG_TASK_ID: string;
  STORE_TOKEN_NAME: string;
  STORE_CUSTOM_NAME: string;
  NO_AUTH_URL: string[];
  WAYS_DATA: Array<{
    type: string;
    [key: string]: any;
  }>;
}

export declare const CONSTANT: ConstantType;