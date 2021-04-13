import axios from 'axios';

export async function PostWithOptionalResponse<T = undefined>(path: string, body?: any) {
    const res = await axios.post<T>(path, body);
    return res.data;
}

export async function Post<T>(path: string, body?: any) {
    const res = await PostWithOptionalResponse<T>(path, body);
    if (!res) {
        throw new Error("unexpected type of response");
    }
    return res;
}