import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { from, Observable, throwError } from 'rxjs';
import { catchError, switchMap } from 'rxjs/operators';
import { ReCaptchaV3Service } from 'ngx-captcha';

import { EnvironmenterService } from './environmenter.service';

export interface FundsRequest {
  walletAddress: string;
  captchaResponse: string;
}

export interface FundsResponse {
  amount: string;
  transactionHash: string;
}

@Injectable({
  providedIn: 'root'
})
export class FaucetService {
  apiEndpoint: string;
  siteKey: string;
  constructor(
    private readonly environmenter: EnvironmenterService,
    private readonly http: HttpClient,
    private readonly reCaptcha: ReCaptchaV3Service,
  ) {
    this.apiEndpoint = this.environmenter.env.apiEndpoint;
    this.siteKey = this.environmenter.env.catpchaSiteKey;
  }
  requestFunds(address: string): Observable<FundsResponse> {
    return from(this.reCaptcha.executeAsPromise(
      this.siteKey, address, { useGlobalDomain: false },
    )).pipe(
      switchMap(token => {
        const req: FundsRequest = {
          walletAddress: address,
          captchaResponse: token,
        };
        return this.http.post<FundsResponse>(`${this.apiEndpoint}/api/v1/faucet/request`, req);
      }),
      catchError(err => {
        return throwError(err);
      })
    );
  }
}
