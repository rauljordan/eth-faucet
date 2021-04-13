import { AxiosResponse } from "axios";

export interface Response<T> {
    status: "OK";
    data: T;
}

export function toData<T>(resp: AxiosResponse<T>): T | undefined {
    if (resp.data) {
        return resp.data as T;
    }
    return undefined;
}