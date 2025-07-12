/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export type LoginCallbackData = ResponseSuccessResponse;

export type LoginCallbackError = ResponseErrorResponse;

export type LoginRedirectData = ResOauthRedirect;

export type LoginRedirectError = ResponseErrorResponse;

export interface PayloadOauthCallback {
  code: string;
}

export interface PayloadOauthRedirect {
  redirectUrl: string;
}

export interface PayloadOverview {
  histories: PayloadOverviewHistoryItem[];
  poolTokens: PayloadPoolTokenCategoryItem[];
  tokenCount: number;
  totalCompleted: number;
  totalFailed: number;
  totalPending: number;
}

export interface PayloadOverviewHistoryItem {
  completed: number;
  failed: number;
  pending: number;
  submitted: number;
}

export interface PayloadOverviewRequest {
  userId: number;
}

export interface PayloadPoolTokenCategoryItem {
  categoryId: number;
  categoryName: string;
  tokenCount: number;
}

export interface PayloadStateResponse {
  displayName: string;
  email: string;
  isAdmin: boolean;
  photoUrl: string;
  userId: number;
}

export interface PayloadTaskCategoryItem {
  createdAt: string;
  id: number;
  name: string;
  updatedAt: string;
}

export interface PayloadTaskCategoryListResponse {
  categories: PayloadTaskCategoryItem[];
}

export interface PayloadTaskDetailRequest {
  taskId: number;
}

export interface PayloadTaskDetailResponse {
  category: PayloadTaskCategoryItem;
  categoryId: number;
  content: string;
  createdAt: string;
  failedReason: string;
  id: number;
  isRaw: boolean;
  source: string;
  status: string;
  title: string;
  tokenCount: number;
  type: string;
  updatedAt: string;
  uploadId: number;
  user: PayloadUserListItem;
  userId: number;
}

export interface PayloadTaskListItem {
  categoryId: number;
  createdAt: string;
  failedReason: string;
  id: number;
  source: string;
  status: string;
  tokenCount: number;
  type: string;
  updatedAt: string;
  uploadId: number;
  userId: number;
}

export interface PayloadTaskListRequest {
  /**
   * @min 1
   * @max 100
   */
  limit: number;
  /** @min 0 */
  offset: number;
  uploadId: number;
  userId: number;
}

export interface PayloadTaskListResponse {
  count: number;
  tasks: PayloadTaskListItem[];
}

export interface PayloadTaskSubmitBatchResponse {
  tasks: PsqlTask[];
  tasksCreated: number;
}

export interface PayloadTaskSubmitRequest {
  category: string;
  source: string;
  type: PayloadTaskSubmitRequestTypeEnum;
}

export enum PayloadTaskSubmitRequestTypeEnum {
  Web = "web",
  Doc = "doc",
  Youtube = "youtube",
}

export interface PayloadTaskSubmitResponse {
  taskId: number;
}

export interface PayloadTaskUploadItem {
  createdAt: string;
  id: number;
  updatedAt: string;
  userId: number;
}

export interface PayloadTaskUploadListResponse {
  uploads: PayloadTaskUploadItem[];
}

export interface PayloadUserListItem {
  createdAt: string;
  email: string;
  firstname: string;
  id: number;
  isAdmin: boolean;
  lastname: string;
  oid: string;
  photoUrl: string;
  updatedAt: string;
}

export interface PayloadUserListResponse {
  users: PayloadUserListItem[];
}

export interface PsqlTask {
  categoryId: number;
  content: string;
  createdAt: string;
  failedReason: string;
  id: number;
  isRaw: boolean;
  source: string;
  status: string;
  title: string;
  tokenCount: number;
  type: string;
  updatedAt: string;
  uploadId: number;
  userId: number;
}

export interface ResOauthRedirect {
  code: string;
  data: PayloadOauthRedirect;
  message: string;
  success: boolean;
}

export interface ResOverview {
  code: string;
  data: PayloadOverview;
  message: string;
  success: boolean;
}

export interface ResStateResponse {
  code: string;
  data: PayloadStateResponse;
  message: string;
  success: boolean;
}

export interface ResTaskCategoryListResponse {
  code: string;
  data: PayloadTaskCategoryListResponse;
  message: string;
  success: boolean;
}

export interface ResTaskDetailResponse {
  code: string;
  data: PayloadTaskDetailResponse;
  message: string;
  success: boolean;
}

export interface ResTaskListResponse {
  code: string;
  data: PayloadTaskListResponse;
  message: string;
  success: boolean;
}

export interface ResTaskSubmitBatchResponse {
  code: string;
  data: PayloadTaskSubmitBatchResponse;
  message: string;
  success: boolean;
}

export interface ResTaskSubmitResponse {
  code: string;
  data: PayloadTaskSubmitResponse;
  message: string;
  success: boolean;
}

export interface ResTaskUploadListResponse {
  code: string;
  data: PayloadTaskUploadListResponse;
  message: string;
  success: boolean;
}

export interface ResUserListResponse {
  code: string;
  data: PayloadUserListResponse;
  message: string;
  success: boolean;
}

export interface ResponseErrorResponse {
  code: string;
  error: string;
  message: string;
  success: boolean;
}

export interface ResponseSuccessResponse {
  code: string;
  data: any;
  message: string;
  success: boolean;
}

export type StateData = ResStateResponse;

export type StateError = ResponseErrorResponse;

export type StateOverviewData = ResOverview;

export type StateOverviewError = ResponseErrorResponse;

export type TaskCategoryListData = ResTaskCategoryListResponse;

export type TaskCategoryListError = ResponseErrorResponse;

export type TaskDetailData = ResTaskDetailResponse;

export type TaskDetailError = ResponseErrorResponse;

export type TaskListData = ResTaskListResponse;

export type TaskListError = ResponseErrorResponse;

export type TaskSubmitBatchData = ResTaskSubmitBatchResponse;

export type TaskSubmitBatchError = ResponseErrorResponse;

export type TaskSubmitData = ResTaskSubmitResponse;

export type TaskSubmitError = ResponseErrorResponse;

export type TaskUploadListData = ResTaskUploadListResponse;

export type TaskUploadListError = ResponseErrorResponse;

export type UserListData = ResUserListResponse;

export type UserListError = ResponseErrorResponse;

export namespace Admin {
  /**
   * No description
   * @tags admin
   * @name UserList
   * @request POST:/admin/user/list
   * @response `200` `UserListData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace UserList {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = UserListData;
  }
}

export namespace Public {
  /**
   * No description
   * @tags public
   * @name LoginCallback
   * @request POST:/public/login/callback
   * @response `200` `LoginCallbackData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace LoginCallback {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = PayloadOauthCallback;
    export type RequestHeaders = {};
    export type ResponseBody = LoginCallbackData;
  }

  /**
   * No description
   * @tags public
   * @name LoginRedirect
   * @request GET:/public/login/redirect
   * @response `200` `LoginRedirectData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace LoginRedirect {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = LoginRedirectData;
  }
}

export namespace State {
  /**
   * No description
   * @tags state
   * @name State
   * @request POST:/state/state
   * @response `200` `StateData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace State {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = StateData;
  }

  /**
   * No description
   * @tags state
   * @name StateOverview
   * @request POST:/state/overview
   * @response `200` `StateOverviewData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace StateOverview {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = PayloadOverviewRequest;
    export type RequestHeaders = {};
    export type ResponseBody = StateOverviewData;
  }
}

export namespace Task {
  /**
   * No description
   * @tags task
   * @name TaskCategoryList
   * @request POST:/task/category/list
   * @response `200` `TaskCategoryListData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskCategoryList {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = TaskCategoryListData;
  }

  /**
   * No description
   * @tags task
   * @name TaskDetail
   * @request POST:/task/detail
   * @response `200` `TaskDetailData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskDetail {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = PayloadTaskDetailRequest;
    export type RequestHeaders = {};
    export type ResponseBody = TaskDetailData;
  }

  /**
   * No description
   * @tags task
   * @name TaskList
   * @request POST:/task/list
   * @response `200` `TaskListData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskList {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = PayloadTaskListRequest;
    export type RequestHeaders = {};
    export type ResponseBody = TaskListData;
  }

  /**
   * No description
   * @tags task
   * @name TaskSubmit
   * @request POST:/task/submit
   * @response `200` `TaskSubmitData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskSubmit {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = PayloadTaskSubmitRequest;
    export type RequestHeaders = {};
    export type ResponseBody = TaskSubmitData;
  }

  /**
   * No description
   * @tags task
   * @name TaskSubmitBatch
   * @request POST:/task/submit/batch
   * @response `200` `TaskSubmitBatchData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskSubmitBatch {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = TaskSubmitBatchData;
  }

  /**
   * No description
   * @tags task
   * @name TaskUploadList
   * @request POST:/task/upload/list
   * @response `200` `TaskUploadListData` OK
   * @response `400` `ResponseErrorResponse` Bad Request
   */
  export namespace TaskUploadList {
    export type RequestParams = {};
    export type RequestQuery = {};
    export type RequestBody = never;
    export type RequestHeaders = {};
    export type ResponseBody = TaskUploadListData;
  }
}

import type {
  AxiosInstance,
  AxiosRequestConfig,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  JsonApi = "application/vnd.api+json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "//localhost:3000/api",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<T> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance
      .request({
        ...requestParams,
        headers: {
          ...(requestParams.headers || {}),
          ...(type ? { "Content-Type": type } : {}),
        },
        params: query,
        responseType: responseFormat,
        data: body,
        url: path,
      })
      .then((response) => response.data);
  };
}

/**
 * @title Backend API
 * @version 1.0
 * @license Apache 2.0 (http://www.apache.org/licenses/LICENSE-2.0.html)
 * @baseUrl //localhost:3000/api
 * @contact API Support <support@swagger.io> (http://www.swagger.io/support)
 *
 * The Swagger API documentation for backend
 */
export class Backend<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  admin = {
    /**
     * No description
     *
     * @tags admin
     * @name UserList
     * @request POST:/admin/user/list
     * @response `200` `UserListData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    userList: (params: RequestParams = {}) =>
      this.request<UserListData, UserListError>({
        path: `/admin/user/list`,
        method: "POST",
        ...params,
      }),
  };
  public = {
    /**
     * No description
     *
     * @tags public
     * @name LoginCallback
     * @request POST:/public/login/callback
     * @response `200` `LoginCallbackData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    loginCallback: (body: PayloadOauthCallback, params: RequestParams = {}) =>
      this.request<LoginCallbackData, LoginCallbackError>({
        path: `/public/login/callback`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags public
     * @name LoginRedirect
     * @request GET:/public/login/redirect
     * @response `200` `LoginRedirectData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    loginRedirect: (params: RequestParams = {}) =>
      this.request<LoginRedirectData, LoginRedirectError>({
        path: `/public/login/redirect`,
        method: "GET",
        ...params,
      }),
  };
  state = {
    /**
     * No description
     *
     * @tags state
     * @name State
     * @request POST:/state/state
     * @response `200` `StateData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    state: (params: RequestParams = {}) =>
      this.request<StateData, StateError>({
        path: `/state/state`,
        method: "POST",
        ...params,
      }),

    /**
     * No description
     *
     * @tags state
     * @name StateOverview
     * @request POST:/state/overview
     * @response `200` `StateOverviewData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    stateOverview: (body: PayloadOverviewRequest, params: RequestParams = {}) =>
      this.request<StateOverviewData, StateOverviewError>({
        path: `/state/overview`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),
  };
  task = {
    /**
     * No description
     *
     * @tags task
     * @name TaskCategoryList
     * @request POST:/task/category/list
     * @response `200` `TaskCategoryListData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskCategoryList: (params: RequestParams = {}) =>
      this.request<TaskCategoryListData, TaskCategoryListError>({
        path: `/task/category/list`,
        method: "POST",
        ...params,
      }),

    /**
     * No description
     *
     * @tags task
     * @name TaskDetail
     * @request POST:/task/detail
     * @response `200` `TaskDetailData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskDetail: (body: PayloadTaskDetailRequest, params: RequestParams = {}) =>
      this.request<TaskDetailData, TaskDetailError>({
        path: `/task/detail`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags task
     * @name TaskList
     * @request POST:/task/list
     * @response `200` `TaskListData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskList: (body: PayloadTaskListRequest, params: RequestParams = {}) =>
      this.request<TaskListData, TaskListError>({
        path: `/task/list`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags task
     * @name TaskSubmit
     * @request POST:/task/submit
     * @response `200` `TaskSubmitData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskSubmit: (body: PayloadTaskSubmitRequest, params: RequestParams = {}) =>
      this.request<TaskSubmitData, TaskSubmitError>({
        path: `/task/submit`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        ...params,
      }),

    /**
     * No description
     *
     * @tags task
     * @name TaskSubmitBatch
     * @request POST:/task/submit/batch
     * @response `200` `TaskSubmitBatchData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskSubmitBatch: (params: RequestParams = {}) =>
      this.request<TaskSubmitBatchData, TaskSubmitBatchError>({
        path: `/task/submit/batch`,
        method: "POST",
        ...params,
      }),

    /**
     * No description
     *
     * @tags task
     * @name TaskUploadList
     * @request POST:/task/upload/list
     * @response `200` `TaskUploadListData` OK
     * @response `400` `ResponseErrorResponse` Bad Request
     */
    taskUploadList: (params: RequestParams = {}) =>
      this.request<TaskUploadListData, TaskUploadListError>({
        path: `/task/upload/list`,
        method: "POST",
        ...params,
      }),
  };
}
