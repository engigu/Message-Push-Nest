import { AxiosInstance, AxiosResponse } from 'axios';

export interface ApiResponse<T = any> {
  code: number;
  msg: string;
  data: T;
}

export declare const request: AxiosInstance;
export declare function handleException(error: any): void;
export declare function logout(): void;