import { Post } from '../hooks/remoteCall';

const requestFundsPath = 'http://localhost:8000/api/v1/faucet/request'

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