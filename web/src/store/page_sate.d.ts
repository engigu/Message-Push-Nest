export interface PageState {
  isLogin: boolean;
  Token: string;
  isShowAddWayDialog: boolean;
  siteConfigData: any;
  ShowDialogData: any;
}

export interface PageStateActions {
  setIsLogin(state: boolean): void;
  setToken(token: string): void;
  setShowAddWayDialog(status: boolean): void;
  setSiteConfigData(configData: any): void;
}

export declare const usePageState: () => PageState & PageStateActions;