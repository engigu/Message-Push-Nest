export interface ConstantType {
  PAGE: number;
  TOTAL: number;
  DEFALUT_SITE_CONFIG: string;
  LOG_TASK_ID: string;
  HOSTED_MSG_TASK_ID: string;
  STORE_TOKEN_NAME: string;
  STORE_CUSTOM_NAME: string;
  NO_AUTH_URL: string[];
  CHANNEL_CATEGORIES: Array<{
    id: string;
    label: string;
  }>;
  WAYS_DATA: Array<{
    type: string;
    label: string;
    meta?: {
      subTitle?: string;
      category?: string;
      categoryId?: string;
    };
    subTitle?: string;
    category?: string;
    categoryId?: string;
    inputs?: any[];
    tips?: any;
    dynamicRecipient?: any;
    [key: string]: any;
  }>;
  WAYS_DATA_MAP: Record<string, any>;
  API_VIEW_DATA: any[];
}

export declare const CONSTANT: ConstantType;