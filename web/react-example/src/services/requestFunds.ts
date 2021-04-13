import { Post } from '../hooks/remoteCall';
import { requestFundsPath } from './apiEndpoints';


export interface FundsRequest {
    walletAddress: string;
    captchaResponse: string;
}

export interface FundsResponse {
   amount: number;
   transactionHash: string;
}

export function requestFunds(req: FundsRequest): Promise<FundsResponse> {
    return Post<FundsResponse>(requestFundsPath, req);
}