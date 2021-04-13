import { environment } from '../environment';

const apiBase: string = environment.apiEndpoint;
export const requestFundsPath = `${apiBase}/api/v1/faucet/request`;