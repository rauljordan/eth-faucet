import axios from 'axios';

export async function PostWithOptionalResponse<T = undefined>(path: string, body?: any) {
    console.log('right before');
    const res = await axios.post<T>(path, body);
    console.log(res);
    if (res.status !== 200) {
        throw new Error(`failed with ${res}`);
    }
    return res.data;
}

export async function Post<T>(path: string, body?: any) {
    const res = await PostWithOptionalResponse<T>(path, body);
    if (!res) {
        throw new Error("unexpected type of response");
    }
    return res;
}